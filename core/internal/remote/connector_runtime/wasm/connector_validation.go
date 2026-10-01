package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/extism/go-sdk"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/tetratelabs/wazero"
)

var (
	CustomSectionName = "openiga:manifest"
)

type ConnectorRuntime struct {
	logger *slog.Logger
}

var _ contract.ConnectorRuntime = (*ConnectorRuntime)(nil)

func NewConnectorRuntime(logger *slog.Logger) *ConnectorRuntime {
	return &ConnectorRuntime{logger: logger}
}

// ValidateConnectorByURL validates the WASM connector based on the URL; on top of this, this function also validates the custom section
// Validation: Download and verify the hash with Extism -> Get the custom section with Wazero runtime
func (c *ConnectorRuntime) ValidateConnectorByURL(ctx context.Context, url string, hash string) (*domain.ConnectorSpec, error) {
	if hash == "" {
		return nil, fmt.Errorf("hash required for the connectionURL %s", url)
	}

	c.logger.Info("validating connector by URL", "url", url, "hash", hash)
	wasmData, err := extism.WasmUrl{Url: url}.ToWasmData(ctx)
	if err != nil {
		return nil, fmt.Errorf(`failed to fetch wasm data from url "%s". reason: %w`, url, err)
	}

	sum := sha256.Sum256(wasmData.Data)
	if actualSha := hex.EncodeToString(sum[:]); actualSha != hash {
		return nil, fmt.Errorf(`expected hash "%s", got "%s"`, hash, actualSha)
	}

	customSection, err := c.readCustomSection(ctx, url, wasmData.Data)
	if err != nil {
		return nil, fmt.Errorf("custom section at %s is invalid: %w", url, err)
	}

	return domain.NewConnectorSpec(customSection)
}

// readCustomSection Extism does not expose method to get custom section; with this the custom section is extracted
// and the manifest is returned for the users for approval process; later the approved configs and endpoints are used
func (c *ConnectorRuntime) readCustomSection(ctx context.Context, connectorUrl string, wasmData []byte) ([]byte, error) {
	wazeroRuntime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCustomSections(true))
	defer common.WithErrorLogged(ctx, c.logger, wazeroRuntime.Close)
	c.logger.Info("Reading custom section", "connectorUrl", connectorUrl)

	module, err := wazeroRuntime.CompileModule(ctx, wasmData)
	if err != nil {
		return nil, fmt.Errorf("failed to compile wasm module: %w", err)
	}
	defer common.WithErrorLogged(ctx, c.logger, module.Close)

	for _, sec := range module.CustomSections() {
		if sec.Name() == CustomSectionName {
			c.logger.Info("Found custom section", "name", sec.Name(), "connectorUrl", connectorUrl)
			return sec.Data(), nil
		}
	}

	return nil, fmt.Errorf("failed to find custom section with name %s", CustomSectionName)
}
