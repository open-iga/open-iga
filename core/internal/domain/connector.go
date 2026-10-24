package domain

// ConnectorKind distinguishes the two operation families a connector exposes. `read` exists in
// both, so kind is what disambiguates them at dispatch.
type ConnectorKind string

const (
	AccountActionKind ConnectorKind = "account-action"
	EntitlementKind   ConnectorKind = "entitlement"
)

// ConnectorDispatchEnvelope is one invocation routed to a connector: which managed resource and
// operation to run (Type, e.g. create/grant), the scoped config, and the handler input. The
// remote marshals it straight to the wasm `dispatch` input, so the json tags are the dispatcher's
// wire contract.
//
// Note: this is the extism wire format living in domain. Fine while wasm is the only runtime;
// split into a domain type + wire DTO if a non-wasm runtime ever appears.
type ConnectorDispatchEnvelope struct {
	Kind            ConnectorKind     `json:"__kind"`
	ManagedResource string            `json:"__managedResource"`
	Type            string            `json:"__type"`
	Config          map[string]string `json:"config"`
	Input           any               `json:"input"`
}

type ConnectorValidationStatus string

const (
	ConnectorValidationPending ConnectorValidationStatus = "pending"
	ConnectorValidationSuccess ConnectorValidationStatus = "success"
	ConnectorValidationFailed  ConnectorValidationStatus = "failed"
)

type ConnectorValidationResult struct {
	Status        ConnectorValidationStatus
	ConnectorUrl  string
	ConnectorHash string
	ConnectorSpec *ConnectorSpec
	// MissingEnvVars are required env variables the connector needs that the core's
	// environment does not have set. Onboarding is blocked while this is non-empty.
	MissingEnvVars []string
	Error          error
}

func (c *ConnectorValidationResult) CanBeOnboarded() bool {
	return c.Status == ConnectorValidationSuccess
}
