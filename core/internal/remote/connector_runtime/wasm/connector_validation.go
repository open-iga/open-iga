package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/tetratelabs/wazero"
)

var CustomSectionName = "openiga:manifest"

const (
	maxConnectorSize = 32 * 1024 * 1024 // 32 MiB cap on a downloaded connector
	fetchTimeout     = 30 * time.Second
)

type ConnectorRuntime struct {
	logger     *slog.Logger
	httpClient *http.Client
}

var _ contract.ConnectorRuntime = (*ConnectorRuntime)(nil)

func NewConnectorRuntime(logger *slog.Logger) *ConnectorRuntime {
	return &ConnectorRuntime{logger: logger, httpClient: newSecureClient()}
}

// maxRedirects bounds redirect following; GitHub release URLs 302 to a signed CDN host.
const maxRedirects = 5

// newSecureClient blocks non-public addresses at dial time (SSRF guard, incl. DNS
// rebinding since the resolved IP is checked). Redirects are followed but kept on
// https and re-dialed through the same guard, so every hop is SSRF-checked.
func newSecureClient() *http.Client {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return fmt.Errorf("invalid dial address: %w", err)
			}
			// TODO: if a internal registry or VCS is used, by design this will be blocked now. Check if this is required in the future
			if ip := net.ParseIP(host); ip == nil || !isPublicIP(ip) {
				return errors.New("blocked non-public address")
			}
			return nil
		},
	}
	return &http.Client{
		Timeout:   fetchTimeout,
		Transport: &http.Transport{DialContext: dialer.DialContext},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "https" {
				return errors.New("redirect to non-https blocked")
			}
			return nil
		},
	}
}

// isPublicIP this is introduced to prevent SSRF attack
func isPublicIP(ip net.IP) bool {
	switch {
	case ip.IsLoopback(), ip.IsPrivate(), ip.IsUnspecified(),
		ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(), ip.IsMulticast():
		return false
	default:
		return true
	}
}

// ValidateConnectorByURL downloads the connector, verifies its hash, and extracts the
// manifest from the wasm custom section.
func (c *ConnectorRuntime) ValidateConnectorByURL(ctx context.Context, connectorURL string, hash string) (*domain.ConnectorSpec, error) {
	if hash == "" {
		return nil, fmt.Errorf("hash required for the connector url %s", connectorURL)
	}

	c.logger.Info("validating connector by URL", "url", connectorURL)
	wasmData, err := c.fetchConnector(ctx, connectorURL)
	if err != nil {
		return nil, err
	}

	if sha256Hex(wasmData) != hash {
		return nil, errors.New("connector hash mismatch")
	}

	customSection, err := c.readCustomSection(ctx, connectorURL, wasmData)
	if err != nil {
		return nil, fmt.Errorf("custom section is invalid: %w", err)
	}

	spec, err := domain.NewConnectorSpec(customSection)
	if err != nil {
		return nil, fmt.Errorf("invalid connector spec: %w", err)
	}

	return spec, nil
}

// fetchConnector downloads the connector over HTTPS with a size cap. Network-level
// errors are logged and returned generically so internal reachability is not leaked.
func (c *ConnectorRuntime) fetchConnector(ctx context.Context, rawURL string) ([]byte, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid connector url: %w", err)
	}
	if parsed.Scheme != "https" {
		return nil, errors.New("connector url must use https")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("connector fetch failed", "url", rawURL, "err", err)
		return nil, errors.New("failed to fetch connector")
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching connector", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxConnectorSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read connector: %w", err)
	}
	if len(data) > maxConnectorSize {
		return nil, fmt.Errorf("connector exceeds max size of %d bytes", maxConnectorSize)
	}

	return data, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// readCustomSection extracts the manifest custom section; Extism does not expose it,
// so the module is parsed with wazero.
func (c *ConnectorRuntime) readCustomSection(ctx context.Context, connectorURL string, wasmData []byte) ([]byte, error) {
	wazeroRuntime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCustomSections(true))
	defer common.WithErrorLogged(ctx, c.logger, wazeroRuntime.Close)
	c.logger.Info("reading custom section", "connectorUrl", connectorURL)

	module, err := wazeroRuntime.CompileModule(ctx, wasmData)
	if err != nil {
		return nil, fmt.Errorf("failed to compile wasm module: %w", err)
	}
	defer common.WithErrorLogged(ctx, c.logger, module.Close)

	for _, sec := range module.CustomSections() {
		if sec.Name() == CustomSectionName {
			return sec.Data(), nil
		}
	}

	return nil, fmt.Errorf("failed to find custom section with name %s", CustomSectionName)
}
