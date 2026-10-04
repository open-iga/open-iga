package domain

import "errors"

var (
	ErrConnectorNotFound       = errors.New("connector not found")
	ErrConnectorNotOnboardable = errors.New("connector cannot be onboarded")
	ErrManagedSystemNameExists = errors.New("managed system name already exists")
)

type ManagedSystem struct {
	Id            string
	Name          string
	ConnectorUrl  string
	ConnectorHash string
}
