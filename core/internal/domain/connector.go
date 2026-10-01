package domain

type ConnectorValidationStatus string

const (
	ConnectorValidationPending ConnectorValidationStatus = "pending"
	ConnectorValidationSuccess ConnectorValidationStatus = "success"
	ConnectorValidationFailed  ConnectorValidationStatus = "failed"
)

type ConnectorValidationResult struct {
	Status        ConnectorValidationStatus
	ConnectorSpec *ConnectorSpec
	Error         error
}
