package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
)

var templateUrl = regexp.MustCompile(`^https?://[\w.\-{}]+(/.*)?$`)

type Config struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    *bool  `json:"required"` // pointer is used as nil is the default value unlike
}

func (c Config) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Description, validation.Required),
		validation.Field(&c.Required, validation.NotNil),
	)
}

type Endpoint struct {
	Method      string `json:"method"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

func (e Endpoint) Validate() error {
	return validation.ValidateStruct(
		&e,
		validation.Field(&e.Method, validation.Required,
			validation.In("GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "get", "post", "put", "delete", "patch", "head")),
		validation.Field(&e.URL, validation.Required, validation.Match(templateUrl)),
		validation.Field(&e.Description, validation.Required),
	)
}

type Operation struct {
	Endpoints   []Endpoint `json:"endpoints"`
	Description string     `json:"description"`
	Config      []Config   `json:"config"`
}

func (o Operation) Validate() error {
	return validation.ValidateStruct(
		&o,
		validation.Field(&o.Endpoints, validation.Required, validation.Each()),
		validation.Field(&o.Description, validation.Required),
		validation.Field(&o.Config, validation.Required, validation.Each()),
	)
}

type AccountActions struct {
	Create  *Operation `json:"create"`
	Enable  *Operation `json:"enable"`
	Disable *Operation `json:"disable"`
	Delete  *Operation `json:"delete"`
	Read    *Operation `json:"read"`
}

func (a AccountActions) Validate() error {
	return validation.ValidateStruct(
		&a,
		validation.Field(&a.Create, validation.Required),
		validation.Field(&a.Enable, validation.Required),
		validation.Field(&a.Disable, validation.Required),
		validation.Field(&a.Delete, validation.Required),
		validation.Field(&a.Read, validation.Required),
	)
}

type UserEntitlements struct {
	Discover *Operation `json:"discover"`
	Grant    *Operation `json:"grant"`
	Revoke   *Operation `json:"revoke"`
	Read     *Operation `json:"read"`
}

func (e UserEntitlements) Validate() error {
	return validation.ValidateStruct(
		&e,
		validation.Field(&e.Discover, validation.Required),
		validation.Field(&e.Grant, validation.Required),
		validation.Field(&e.Revoke, validation.Required),
		validation.Field(&e.Read, validation.Required),
	)
}

type ConnectorSpec struct {
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	Config         []Config                    `json:"config"`
	AllowedDomains []string                    `json:"allowedDomains"`
	Actions        map[string]AccountActions   `json:"actions"`
	Entitlements   map[string]UserEntitlements `json:"entitlements"`
}

func (c *ConnectorSpec) Validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Description, validation.Required),
		validation.Field(&c.Config, validation.Required, validation.Each()),
		validation.Field(&c.AllowedDomains, validation.Required, validation.Each(validation.Required)),
		validation.Field(&c.Actions, validation.Required, validation.Each()),
		validation.Field(&c.Entitlements, validation.Required, validation.Each()),
	)
}

// NewConnectorSpec decodes the custom section byte and returns the connector spec
func NewConnectorSpec(customSection []byte) (*ConnectorSpec, error) {
	var connectorSpec ConnectorSpec

	dec := json.NewDecoder(bytes.NewReader(customSection))
	if err := dec.Decode(&connectorSpec); err != nil {
		return nil, fmt.Errorf("invalid connector spec: %w", err)
	}

	if err := connectorSpec.Validate(); err != nil {
		return nil, fmt.Errorf("connector spec validation failed: %w", err)
	}

	return &connectorSpec, nil
}

// MissingRequiredEnvVars returns the names of required config entries that the core's
// environment has unset or empty, matching how endpoint templates treat an empty value. Config
// entries not marked required are allowed to be absent.
func (c *ConnectorSpec) MissingRequiredEnvVars() []string {
	missing := make([]string, 0)
	for _, cfg := range c.Config {
		if cfg.Required == nil || !*cfg.Required {
			continue
		}
		if value, ok := os.LookupEnv(cfg.Name); !ok || value == "" {
			missing = append(missing, cfg.Name)
		}
	}
	return missing
}

// endpointVar matches a `{{VAR}}` placeholder in an endpoint URL (e.g. https://sts.{{AWS_REGION}}...).
var endpointVar = regexp.MustCompile(`{{\s*([A-Za-z0-9_]+)\s*}}`)

// ResolvedEndpoint is a spec endpoint with its {{VAR}} placeholders substituted.
type ResolvedEndpoint struct {
	Method string
	URL    string
}

// OperationKey is the key of an operation in the ResolveEndpoints and ScopedConfigs results:
// "<kind>.<resource>.<operation>", e.g. "account-action.iam-user.create". The kind is part of it
// because a resource name or an operation like read can exist under both account actions and
// entitlements.
func OperationKey(kind ConnectorKind, resource string, operation string) string {
	return string(kind) + "." + resource + "." + operation
}

// ScopedConfigs returns, per operation (keyed by OperationKey), the env values the Host must send
// for an invocation: the module-level config plus that operation's own declared config. Only keys
// set in the environment are included (required ones are guaranteed present by the onboarding
// gate). Scoping is the Host's isolation boundary — an operation only ever receives the config
// keys it declared, so e.g. an entitlement never sees an account-action's role ARN.
//
// The result is shared by every invocation, so callers must treat it as read-only.
func (c *ConnectorSpec) ScopedConfigs() map[string]map[string]string {
	scoped := map[string]map[string]string{}
	c.forEachOperation(func(kind ConnectorKind, resource, operation string, op *Operation) {
		config := map[string]string{}
		for _, declared := range [][]Config{c.Config, op.Config} {
			for _, cfg := range declared {
				if value, ok := os.LookupEnv(cfg.Name); ok {
					config[cfg.Name] = value
				}
			}
		}
		scoped[OperationKey(kind, resource, operation)] = config
	})
	return scoped
}

// forEachOperation visits every declared operation across actions and entitlements.
func (c *ConnectorSpec) forEachOperation(visit func(kind ConnectorKind, resource, operation string, op *Operation)) {
	for resource, actions := range c.Actions {
		for name, op := range map[string]*Operation{
			"create": actions.Create, "enable": actions.Enable, "disable": actions.Disable,
			"delete": actions.Delete, "read": actions.Read,
		} {
			if op != nil {
				visit(AccountActionKind, resource, name, op)
			}
		}
	}
	for resource, ents := range c.Entitlements {
		for name, op := range map[string]*Operation{
			"discover": ents.Discover, "grant": ents.Grant, "revoke": ents.Revoke, "read": ents.Read,
		} {
			if op != nil {
				visit(EntitlementKind, resource, name, op)
			}
		}
	}
}

// substituteURL replaces each {{VAR}} in rawURL with its env value. The connector SDK guarantees
// every endpoint var is a required config entry, so the MissingRequiredEnvVars gate guarantees it
// is set by the time endpoints are resolved.
func substituteURL(rawURL string, logger *slog.Logger) string {
	resolved := rawURL
	for _, match := range endpointVar.FindAllStringSubmatch(rawURL, -1) {
		placeholder, name := match[0], match[1] // "{{AWS_REGION}}", "AWS_REGION"
		value := os.Getenv(name)
		if value == "" {
			logger.Error("endpoint template var is unset or empty; connectors must be built with the connector SDK", "var", name, "url", rawURL)
		}
		resolved = strings.ReplaceAll(resolved, placeholder, value)
	}
	return resolved
}

// ResolveEndpoints substitutes every endpoint's {{VAR}} placeholders from the core's environment
// and returns the concrete endpoints, keyed by OperationKey(kind, resource, operation), so egress
// can be gated per operation.
func (c *ConnectorSpec) ResolveEndpoints(logger *slog.Logger) map[string][]ResolvedEndpoint {
	resolved := map[string][]ResolvedEndpoint{}
	c.forEachOperation(func(kind ConnectorKind, resource, operation string, op *Operation) {
		key := OperationKey(kind, resource, operation)
		for _, endpoint := range op.Endpoints {
			resolved[key] = append(resolved[key], ResolvedEndpoint{
				Method: strings.ToUpper(endpoint.Method), URL: substituteURL(endpoint.URL, logger),
			})
		}
	})
	return resolved
}
