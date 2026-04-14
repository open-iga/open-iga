package application

import (
	"github.com/open-iga/core/internal/application/oauth"
)

type RuntimeApplication struct {
	LoginService *oauth.LoginService
}
