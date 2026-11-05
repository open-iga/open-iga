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

// connectorExecutionTimeoutMs is the maximum time one dispatch call may run. Setting it is also
// what lets a cancelled context stop a running guest; never set it to 0.
const connectorExecutionTimeoutMs = 60_000

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

// egressError logs and builds a failed fetch response with the given message.
func (c *ConnectorRuntime) egressError(status int, message string) egressResponse {
	c.logger.Error("connector egress failed", "status", status, "error", message)
	return egressResponse{Status: status, Body: "", Error: message}
}

// hostFunctions are the capabilities exposed to connectors (namespace extism:host/user). fetch is
// the sole egress: it routes connector HTTP through the host's SSRF-guarded client and allows only
// the connector's declared domains and the resolved endpoints of the operation being executed.
func (c *ConnectorRuntime) hostFunctions(spec *domain.ConnectorSpec) []extism.HostFunction {
	globs := compileGlobs(spec.AllowedDomains)

	fetch := extism.NewHostFunctionWithStack(
		"fetch",
		func(ctx context.Context, cp *extism.CurrentPlugin, stack []uint64) {
			stack[0] = c.handleFetch(ctx, cp, spec.AllowedDomains, globs, stack[0])
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
func (c *ConnectorRuntime) handleFetch(ctx context.Context, cp *extism.CurrentPlugin, allowedDomains []string, globs []glob.Glob, reqOffset uint64) uint64 {
	response := c.doFetch(ctx, cp, allowedDomains, globs, reqOffset)

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
func (c *ConnectorRuntime) doFetch(ctx context.Context, cp *extism.CurrentPlugin, allowedDomains []string, globs []glob.Glob, reqOffset uint64) egressResponse {
	reqBytes, err := cp.ReadBytes(reqOffset)
	if err != nil {
		return c.egressError(http.StatusBadRequest, fmt.Sprintf("failed to read request from guest memory at offset %d: %v", reqOffset, err))
	}
	var request egressRequest
	if err := json.Unmarshal(reqBytes, &request); err != nil {
		return c.egressError(http.StatusBadRequest, fmt.Sprintf("invalid request json: %v", err))
	}

	parsed, err := url.Parse(request.URL)
	if err != nil || parsed.Scheme != "https" {
		return c.egressError(http.StatusForbidden, fmt.Sprintf("url %q must be a valid https url", request.URL))
	}

	// Domain level: host must match one of the connector's allowed domains.
	if !matchesAnyGlob(globs, parsed.Hostname()) {
		return c.egressError(http.StatusForbidden, fmt.Sprintf("egress blocked: host %s is not in the connector's allowed domains %v", parsed.Hostname(), allowedDomains))
	}

	// Action level: host+method must match a resolved endpoint of the operation being executed
	// endpoints are set on the context by Execute
	allowedEndpoints := allowedEndpointsFrom(ctx)
	method := strings.ToUpper(request.Method)
	if method == "" {
		method = http.MethodGet
	}
	if !allowedByEndpoints(allowedEndpoints, parsed.Hostname(), method) {
		return c.egressError(http.StatusForbidden, fmt.Sprintf("egress blocked: %s method on %s does not match any endpoint declared for this operation", method, parsed.Hostname()))
	}

	return c.send(ctx, request, method, parsed.Hostname())
}

// send performs the already-gated request and returns the upstream status and body.
func (c *ConnectorRuntime) send(ctx context.Context, request egressRequest, method string, host string) egressResponse {
	var body io.Reader
	if request.Body != "" {
		body = strings.NewReader(request.Body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, request.URL, body)
	if err != nil {
		return c.egressError(http.StatusInternalServerError, fmt.Sprintf("failed to build %s request to %s: %v", method, host, err))
	}
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Redirects bypass the domain and endpoint filters (checked on request.URL only), so they are
	// not followed: the connector receives the 3xx and must call the new URL through fetch, where
	// it is gated. A shallow copy keeps the shared transport; the wasm download still follows redirects.
	client := *c.httpClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return c.egressError(http.StatusBadGateway, fmt.Sprintf("%s request to %s failed: %v", method, host, err))
	}
	defer func() { _ = httpResp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(httpResp.Body, maxEgressResponseBytes))
	if err != nil {
		return c.egressError(http.StatusBadGateway, fmt.Sprintf("failed to read response from %s: %v", host, err))
	}
	return egressResponse{Status: httpResp.StatusCode, Body: string(data), Error: ""}
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
