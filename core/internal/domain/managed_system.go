package domain

import "errors"

var (
	ErrConnectorNotFound       = errors.New("connector not found")
	ErrConnectorNotOnboardable = errors.New("connector cannot be onboarded")
)

type ManagedSystem struct {
	Id            string
	Name          string
	ConnectorUrl  string
	ConnectorHash string
}
