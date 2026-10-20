package domain

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
