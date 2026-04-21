package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorReader struct{}

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestGenerateHighEntropyID(t *testing.T) {
	t.Run("returns a base64url encoded string", func(t *testing.T) {
		id, err := GenerateHighEntropyID()
		assert.NoError(t, err)
		assert.Len(t, id, 43) // 32 bytes → 43 base64url chars (no padding)
	})

	t.Run("returns error when reader fails", func(t *testing.T) {
		id, err := generateHighEntropyId(&errorReader{})
		assert.Error(t, err)
		assert.Empty(t, id)
	})
}
