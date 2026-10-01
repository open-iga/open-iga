package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation"
)

var templateUrl = regexp.MustCompile(`^https?://[\w.\-{}]+(/.*)?$`)

type Config struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    *bool  `json:"required"` // pointer is used as nil is the default value unlike
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
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
	return validation.ValidateStruct(&e,
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
	return validation.ValidateStruct(&o,
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
	return validation.ValidateStruct(&a,
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
	return validation.ValidateStruct(&e,
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

func (c ConnectorSpec) Validate() error {
	return validation.ValidateStruct(&c,
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
