package domain

// ConnectorKind distinguishes the two operation families a connector exposes. `read` exists in
// both, so kind is what disambiguates them at dispatch.
type ConnectorKind string

const (
	AccountActionKind ConnectorKind = "account-action"
	EntitlementKind   ConnectorKind = "entitlement"
)

// ConnectorInvocation is a caller's request to run one operation on a connector: which managed
// resource and operation (Type, e.g. create/grant), and the handler input. It has no config —
// config is never the caller's to set; the runtime looks it up itself, scoped to this operation,
// when it builds the wire envelope.
type ConnectorInvocation struct {
	Kind            ConnectorKind
	ManagedResource string
	Type            string
	Input           any
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
	Error         error
}

func (c *ConnectorValidationResult) CanBeOnboarded() bool {
	return c.Status == ConnectorValidationSuccess
}
