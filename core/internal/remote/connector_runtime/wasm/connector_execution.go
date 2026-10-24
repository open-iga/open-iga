package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	extism "github.com/extism/go-sdk"
	"github.com/open-iga/core/internal/domain"
)

// connectorEntry is a one-shot future for a compile: ready closes when the compile finishes,
// after which compiled/err are set and safe to read (the close establishes happens-before).
// Execute waits on ready instead of racing the background compile.
type connectorEntry struct {
	ready    chan struct{}
	compiled *extism.CompiledPlugin
	err      error
}

// connectorCache keeps compiled connectors keyed by url+hash so repeat invocations skip the
// CPU-expensive wasm compile. Hash in the key means a connector upgrade compiles fresh.
// A CompiledPlugin is safe to share; callers derive a per-invocation Instance (own linear
// memory) from it, since its Call is not goroutine-safe.
type connectorCache struct {
	mu      sync.Mutex
	entries map[string]*connectorEntry
}

func newConnectorCache() *connectorCache {
	return &connectorCache{entries: make(map[string]*connectorEntry)}
}

func cacheKey(connectorURL string, hash string) string {
	return connectorURL + "@" + hash
}

// cacheConnector compiles the verified wasm and stores it under url+hash, in the background.
// Called from validation, which already holds the verified bytes, so execution never re-downloads.
// spec.AllowedDomains gates the connector's outbound HTTP.
//
// Compile is CPU-heavy, so it runs off the caller's path (startup/onboarding returns fast) on a
// detached context (the caller's timeout/cancel must not abort a warm-up). The entry is published
// before the compile starts so an Execute that arrives mid-compile waits rather than missing; a
// failed compile is dropped from the cache so a later validation retries.
func (c *ConnectorRuntime) cacheConnector(ctx context.Context, connectorURL string, hash string, wasmData []byte, spec *domain.ConnectorSpec) {
	key := cacheKey(connectorURL, hash)

	c.cache.mu.Lock()
	if _, exists := c.cache.entries[key]; exists {
		c.cache.mu.Unlock()
		return
	}
	entry := &connectorEntry{ready: make(chan struct{})}
	c.cache.entries[key] = entry
	c.cache.mu.Unlock()

	go func() {
		ctx := context.WithoutCancel(ctx)
		c.logger.Info("compiling connector into cache", "key", key)

		manifest := extism.Manifest{
			Wasm:         []extism.Wasm{extism.WasmData{Data: wasmData}},
			AllowedHosts: spec.AllowedDomains,
		}

		entry.compiled, entry.err = extism.NewCompiledPlugin(ctx, manifest, extism.PluginConfig{EnableWasi: true}, c.hostFunctions())
		close(entry.ready)

		if entry.err != nil {
			c.logger.Error("failed to compile connector", "url", connectorURL, "hash", hash, "err", entry.err)
			c.cache.mu.Lock()
			delete(c.cache.entries, key)
			c.cache.mu.Unlock()
			return
		}
		c.logger.Info("compiled connector into cache", "key", key)
	}()
}

// Execute runs one invocation on the connector cached for url+hash. It derives a fresh instance
// (own linear memory, so concurrent Executes don't collide), calls the wasm `dispatch` export
// with the envelope, and returns the connector's raw JSON output.
func (c *ConnectorRuntime) Execute(ctx context.Context, connectorURL string, hash string, envelope domain.ConnectorDispatchEnvelope) ([]byte, error) {
	c.cache.mu.Lock()
	entry, ok := c.cache.entries[cacheKey(connectorURL, hash)]
	c.cache.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("connector not loaded for %s", connectorURL)
	}

	// Wait for an in-flight background compile rather than racing it.
	select {
	case <-entry.ready:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if entry.err != nil {
		return nil, fmt.Errorf("connector not compiled for %s: %w", connectorURL, entry.err)
	}
	compiled := entry.compiled

	if envelope.Config == nil {
		envelope.Config = map[string]string{} // dispatcher requires an object, not null
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal dispatch envelope: %w", err)
	}

	instance, err := compiled.Instance(ctx, extism.PluginInstanceConfig{})
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

// hostFunctions are the capabilities exposed to connectors (namespace extism:host/user).
//
// TODO: remove sendEmail. Email capability withdrawn (phishing vector). No-op stub kept only so connectors still importing it instantiate; returns ok without sending.
func (c *ConnectorRuntime) hostFunctions() []extism.HostFunction {
	sendEmail := extism.NewHostFunctionWithStack(
		"sendEmail",
		func(_ context.Context, _ *extism.CurrentPlugin, stack []uint64) {
			stack[0] = 0 // 0 = ok
		},
		[]extism.ValueType{extism.ValueTypeI64},
		[]extism.ValueType{extism.ValueTypeI64},
	)
	return []extism.HostFunction{sendEmail}
}
