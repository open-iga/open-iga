package wasm

import (
	"testing"

	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMatchesAnyGlob(t *testing.T) {
	globs := compileGlobs([]string{"*.amazonaws.com"})

	assert.True(t, matchesAnyGlob(globs, "sts.us-east-1.amazonaws.com"))
	assert.True(t, matchesAnyGlob(globs, "iam.amazonaws.com"))
	assert.False(t, matchesAnyGlob(globs, "evil.com"))
	assert.False(t, matchesAnyGlob(globs, "amazonaws.com.evil.com"))
}

func TestAllowedByEndpoints(t *testing.T) {
	endpoints := []domain.ResolvedEndpoint{
		{Method: "POST", URL: "https://iam.amazonaws.com/"},
		{Method: "POST", URL: "https://sts.us-east-1.amazonaws.com/"},
	}

	t.Run("allows a declared host+method", func(t *testing.T) {
		assert.True(t, allowedByEndpoints(endpoints, "iam.amazonaws.com", "POST"))
	})

	t.Run("denies an undeclared host even within the domain", func(t *testing.T) {
		assert.False(t, allowedByEndpoints(endpoints, "ec2.amazonaws.com", "POST"))
	})

	t.Run("denies a mismatched method", func(t *testing.T) {
		assert.False(t, allowedByEndpoints(endpoints, "iam.amazonaws.com", "GET"))
	})

	t.Run("fails closed on empty endpoints", func(t *testing.T) {
		assert.False(t, allowedByEndpoints(nil, "iam.amazonaws.com", "POST"))
	})
}
