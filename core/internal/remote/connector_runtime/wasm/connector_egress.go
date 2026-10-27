package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	extism "github.com/extism/go-sdk"
	"github.com/gobwas/glob"
	"github.com/open-iga/core/internal/domain"
)

// maxEgressResponseBytes caps a connector HTTP response read into guest memory.
const maxEgressResponseBytes = 16 * 1024 * 1024

// maxConnectorMemoryPages caps each instance's wasm linear memory (page = 64 KiB): 800 × 64 KiB = 50 MiB.
const maxConnectorMemoryPages = 800

// egressRequest / egressResponse are the JSON the guest `fetch` polyfill exchanges with the host.
type egressRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type egressResponse struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
	Error  string `json:"error,omitempty"`
}

// hostFunctions are the capabilities exposed to connectors (namespace extism:host/user). fetch is
// the sole egress: it routes connector HTTP through the host's SSRF-guarded client and allows only
// the connector's declared domains and the resolved endpoints of the operation being executed.
func (c *ConnectorRuntime) hostFunctions(spec *domain.ConnectorSpec) []extism.HostFunction {
	globs := compileGlobs(spec.AllowedDomains)

	fetch := extism.NewHostFunctionWithStack(
		"fetch",
		func(ctx context.Context, cp *extism.CurrentPlugin, stack []uint64) {
			stack[0] = c.handleFetch(ctx, cp, globs, stack[0])
		},
		[]extism.ValueType{extism.ValueTypeI64},
		[]extism.ValueType{extism.ValueTypeI64},
	)

	// TODO: remove sendEmail. Email capability withdrawn (phishing vector). No-op stub kept only so
	// connectors still importing it instantiate; returns ok without sending.
	sendEmail := extism.NewHostFunctionWithStack(
		"sendEmail",
		func(_ context.Context, _ *extism.CurrentPlugin, stack []uint64) {
			stack[0] = 0 // 0 = ok
		},
		[]extism.ValueType{extism.ValueTypeI64},
		[]extism.ValueType{extism.ValueTypeI64},
	)

	return []extism.HostFunction{fetch, sendEmail}
}

// handleFetch reads the request from guest memory, performs the gated request, and writes the
// response back into guest memory, returning its offset (0 on a memory write failure).
func (c *ConnectorRuntime) handleFetch(ctx context.Context, cp *extism.CurrentPlugin, globs []glob.Glob, reqOffset uint64) uint64 {
	response := c.doFetch(ctx, cp, globs, reqOffset)

	out, err := json.Marshal(response)
	if err != nil {
		out = []byte(`{"status":500,"error":"failed to marshal response"}`)
	}
	offset, err := cp.WriteBytes(out)
	if err != nil {
		c.logger.Error("failed to write egress response to guest memory", "err", err)
		return 0
	}
	return offset
}

// doFetch applies both egress filters then performs the request. Any denial or failure is returned
// as an egressResponse with Error set, never a panic — the connector sees a failed fetch.
func (c *ConnectorRuntime) doFetch(ctx context.Context, cp *extism.CurrentPlugin, globs []glob.Glob, reqOffset uint64) egressResponse {
	reqBytes, err := cp.ReadBytes(reqOffset)
	if err != nil {
		return egressResponse{Status: http.StatusBadRequest, Error: fmt.Sprintf("failed to read request from guest memory at offset %d: %v", reqOffset, err)}
	}
	var request egressRequest
	if err := json.Unmarshal(reqBytes, &request); err != nil {
		return egressResponse{Status: http.StatusBadRequest, Error: fmt.Sprintf("invalid request json: %v", err)}
	}

	parsed, err := url.Parse(request.URL)
	if err != nil || parsed.Scheme != "https" {
		return egressResponse{Status: http.StatusForbidden, Error: fmt.Sprintf("url %q must be a valid https url", request.URL)}
	}

	// Domain level: host must match one of the connector's allowed domains.
	if !matchesAnyGlob(globs, parsed.Hostname()) {
		return egressResponse{Status: http.StatusForbidden, Error: fmt.Sprintf("egress blocked: host %s not in allowed domains %s", parsed.Hostname(), globs)}
	}

	// Action level: host+method must match a resolved endpoint of the operation being executed
	// endpoints are set on the context by Execute
	allowedEndpoints := allowedEndpointsFrom(ctx)
	method := strings.ToUpper(request.Method)
	if method == "" {
		method = http.MethodGet
	}
	if !allowedByEndpoints(allowedEndpoints, parsed.Hostname(), method) {
		return egressResponse{Status: http.StatusForbidden, Error: fmt.Sprintf("egress blocked: %s method on %s does not match any endpoint declared for this operation", method, parsed.Hostname())}
	}

	var body io.Reader
	if request.Body != "" {
		body = strings.NewReader(request.Body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, request.URL, body)
	if err != nil {
		return egressResponse{Status: http.StatusInternalServerError, Error: fmt.Sprintf("failed to build %s request to %s: %v", method, parsed.Hostname(), err)}
	}
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error("connector egress request failed", "host", parsed.Hostname(), "err", err)
		return egressResponse{Status: http.StatusBadGateway, Error: fmt.Sprintf("%s request to %s failed: %v", method, parsed.Hostname(), err)}
	}
	defer func() { _ = httpResp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(httpResp.Body, maxEgressResponseBytes))
	if err != nil {
		return egressResponse{Status: http.StatusBadGateway, Error: fmt.Sprintf("failed to read response from %s: %v", parsed.Hostname(), err)}
	}
	return egressResponse{Status: httpResp.StatusCode, Body: string(data)}
}

// allowedByEndpoints reports whether host+method matches any resolved endpoint. Empty endpoints
// (no operation context, or an operation with none declared) denies — egress fails closed.
func allowedByEndpoints(endpoints []domain.ResolvedEndpoint, host string, method string) bool {
	for _, endpoint := range endpoints {
		parsed, err := url.Parse(endpoint.URL)
		if err != nil {
			continue
		}
		if parsed.Hostname() == host && (endpoint.Method == "" || endpoint.Method == method) {
			return true
		}
	}
	return false
}

func compileGlobs(patterns []string) []glob.Glob {
	globs := make([]glob.Glob, 0, len(patterns))
	for _, pattern := range patterns {
		if compiled, err := glob.Compile(pattern); err == nil {
			globs = append(globs, compiled)
		}
	}
	return globs
}

func matchesAnyGlob(globs []glob.Glob, host string) bool {
	for _, g := range globs {
		if g.Match(host) {
			return true
		}
	}
	return false
}
