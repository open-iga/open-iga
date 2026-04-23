package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// GenerateHighEntropyID generates a 256 bit high entropy random string suitable for use as a state parameter in OAuth2 flows
// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
func GenerateHighEntropyID() (string, error) {
	return generateHighEntropyId(rand.Reader)
}

func generateHighEntropyId(reader io.Reader) (string, error) {
	b := make([]byte, 32)

	if _, err := io.ReadFull(reader, b); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
