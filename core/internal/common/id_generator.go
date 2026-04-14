package common

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// GenerateHighEntropyID generates a 256 bit high entropy random string suitable for use as a state parameter in OAuth2 flows
// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
func GenerateHighEntropyID() string {
	b := make([]byte, 32) // 32 bytes = 256 bits

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic("crypto/rand.ReadFull failed: " + err.Error()) // panicking here as this is the critical error we can't recover from
	}

	return base64.RawURLEncoding.EncodeToString(b)
}
