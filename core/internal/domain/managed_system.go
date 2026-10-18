package domain

import (
	"errors"
	"time"
)

var (
	ErrConnectorNotFound       = errors.New("connector not found")
	ErrConnectorNotOnboardable = errors.New("connector cannot be onboarded")
	ErrManagedSystemNameExists = errors.New("managed system name already exists")
	ErrManagedSystemNotFound   = errors.New("managed system not found")
)

type ManagedSystem struct {
	Id            string
	Name          string
	ConnectorUrl  string
	ConnectorHash string
	ConnectorSpec *ConnectorSpec
	CreatedAt     time.Time
}

type ManagedSystemTenant struct {
	Id              string
	ManagedSystemId string
	ExternalId      string
	Name            string
	CreatedAt       time.Time
}
