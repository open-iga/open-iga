package wasm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDispatchError(t *testing.T) {
	assert.EqualError(t, dispatchError([]byte(`{"error":"Error: no such user"}`), nil), "connector dispatch failed: Error: no such user")

	runtimeErr := errors.New("wasm trap")
	err := dispatchError([]byte(`{"error":"Error: boom"}`), runtimeErr)
	assert.EqualError(t, err, "connector dispatch failed: Error: boom: wasm trap")
	assert.ErrorIs(t, err, runtimeErr)

	assert.EqualError(t, dispatchError([]byte("not json"), nil), "connector dispatch failed: not json")
	assert.EqualError(t, dispatchError(nil, nil), "connector dispatch failed: ")
}

func TestConnectorErrorMessage(t *testing.T) {
	message, failed := connectorErrorMessage([]byte(`{"error":"Error: Config \"AWS_USER_MANAGEMENT_ROLE\" is required but missing from the Host config"}`))
	assert.True(t, failed)
	assert.Equal(t, `Error: Config "AWS_USER_MANAGEMENT_ROLE" is required but missing from the Host config`, message)

	_, failed = connectorErrorMessage([]byte(`{"accountId":"arn:1","created":true}`))
	assert.False(t, failed)
	_, failed = connectorErrorMessage(nil)
	assert.False(t, failed)
}
