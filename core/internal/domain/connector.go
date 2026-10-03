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
	Error         error
}

func (c *ConnectorValidationResult) CanBeOnboarded() bool {
	return c.Status == ConnectorValidationSuccess
}
