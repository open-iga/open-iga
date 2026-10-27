package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
)

var _ contract.ConnectorRuntime = (*ConnectorRuntime)(nil)

func NewConnectorRuntime(logger *slog.Logger) *ConnectorRuntime {
	return &ConnectorRuntime{logger: logger, httpClient: newSecureClient(), verifiedWasmBytes: common.SyncMap[string, []byte]{}}
}

// ConnectorDispatchEnvelope is the JSON sent as the wasm `dispatch` input: an invocation plus the
// config the runtime scoped to it. The json tags are the connector SDK dispatcher's wire contract.
type ConnectorDispatchEnvelope struct {
	Kind            domain.ConnectorKind `json:"__kind"`
	ManagedResource string               `json:"__managedResource"`
	Type            string               `json:"__type"`
	Config          map[string]string    `json:"config"`
	Input           any                  `json:"input"`
}

// connectorEntry is a one-shot future for a compile: ready closes when the compile finishes,
// after which compiled/err are set and safe to read (the close establishes happens-before).
// Execute waits on ready instead of racing the background compile. configs and endpoints are
// resolved synchronously up front (not gated by the compile), keyed by domain.OperationKey:
// configs are what the dispatcher receives, endpoints gate egress per operation. Both are
// read-only and shared by every Execute.
type connectorEntry struct {
	ready     chan struct{}
	compiled  *extism.CompiledPlugin
	err       error
	configs   map[string]map[string]string
	endpoints map[string][]domain.ResolvedEndpoint
}

// connectors keeps connectorEntry values keyed by url+hash so repeat invocations skip the
// CPU-expensive wasm compile. Hash in the key means a connector upgrade compiles fresh. A
// CompiledPlugin is safe to share; callers derive a per-invocation Instance (own linear memory)
// from it, since its Call is not goroutine-safe.
var connectors common.SyncMap[string, *connectorEntry]

func cacheKey(connectorURL string, hash string) string {
	return connectorURL + "@" + hash
}

// CompileConnector resolves the connector's endpoints from the environment and caches a compiled
// plugin under url+hash. Called post-approval (onboarding) and at startup — by which point the
// required env is guaranteed set (MissingRequiredEnvVars), and every endpoint var is required.
//
// Endpoints resolve synchronously (fast, and available immediately for egress gating). The
// CPU-heavy compile — fetch+verify+NewCompiledPlugin — runs in the background on a detached
// context so startup/onboarding returns fast; the entry is published before it starts, so an
// Execute arriving mid-compile waits (on ready) rather than missing. A failed compile is dropped
// so a later call retries. No-op if already cached.
func (c *ConnectorRuntime) CompileConnector(ctx context.Context, connectorURL string, hash string, spec *domain.ConnectorSpec) {
	key := cacheKey(connectorURL, hash)
	entry := &connectorEntry{ready: make(chan struct{}), configs: spec.ScopedConfigs(), endpoints: spec.ResolveEndpoints(c.logger), err: nil, compiled: nil}
	if _, loaded := connectors.LoadOrStore(key, entry); loaded {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
		defer cancel()

		c.logger.Info("compiling connector into cache", "key", key)

		var wasmData []byte
		if wasmData, entry.err = c.fetchAndVerify(ctx, connectorURL, hash); entry.err == nil {
			manifest := extism.Manifest{
				Wasm: []extism.Wasm{extism.WasmData{Data: wasmData}},
				// Extism's built-in HTTP is disabled so all egress must go through our gated fetch host function
				AllowedHosts: nil,
				Memory:       &extism.ManifestMemory{MaxPages: maxConnectorMemoryPages},
			}
			entry.compiled, entry.err = extism.NewCompiledPlugin(ctx, manifest, extism.PluginConfig{EnableWasi: true}, c.hostFunctions(spec))
		}
		close(entry.ready)

		if entry.err != nil {
			c.logger.Error("failed to compile connector", "url", connectorURL, "hash", hash, "err", entry.err)
			connectors.Delete(key)
			return
		}
		c.logger.Info("compiled connector into cache", "key", key)
	}()
}

// Execute runs one invocation on the connector cached for url+hash. It derives a fresh instance
// (own linear memory, so concurrent Executes don't collide), calls the wasm `dispatch` export
// with the invocation, and returns the connector's raw JSON output.
// TODO: Gate execution until all the connectors are complied and limit maximum number of execution
func (c *ConnectorRuntime) Execute(ctx context.Context, connectorURL string, hash string, invocation domain.ConnectorInvocation) ([]byte, error) {
	entry, ok := connectors.Load(cacheKey(connectorURL, hash))
	if !ok {
		return nil, fmt.Errorf("connector not loaded for %s", connectorURL)
	}

	// Wait for an in-flight background compile rather than racing it
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting for connector %s: %w", connectorURL, ctx.Err())
	case <-entry.ready:
	}
	if entry.err != nil {
		return nil, fmt.Errorf("connector not compiled for %s: %w", connectorURL, entry.err)
	}

	// Config is the Host's isolation boundary: the runtime looks it up itself, scoped to this
	// operation (module + operation config, from env) — never taken from the caller.
	key := domain.OperationKey(invocation.Kind, invocation.ManagedResource, invocation.Type)
	config, ok := entry.configs[key]
	if !ok {
		return nil, fmt.Errorf("connector %s declares no operation %s", connectorURL, key)
	}
	envelope := ConnectorDispatchEnvelope{
		Kind:            invocation.Kind,
		ManagedResource: invocation.ManagedResource,
		Type:            invocation.Type,
		Config:          config,
		Input:           invocation.Input,
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal dispatch envelope: %w", err)
	}

	// Scope egress to this operation's endpoints: the fetch host function reads them off the ctx.
	ctx = withAllowedEndpoints(ctx, entry.endpoints[key])

	instance, err := entry.compiled.Instance(ctx, extism.PluginInstanceConfig{})
	if err != nil {
		return nil, fmt.Errorf("instantiate connector: %w", err)
	}
	defer func() { _ = instance.Close(ctx) }()

	// dispatch returns 0 ok / 1 error; on error the output holds {"error": ...}.
	exit, out, err := instance.CallWithContext(ctx, "dispatch", payload)
	if err != nil {
		return nil, fmt.Errorf("connector dispatch: %w", err)
	}
	if exit != 0 {
		return nil, fmt.Errorf("connector dispatch failed: %s", out)
	}

	return out, nil
}
