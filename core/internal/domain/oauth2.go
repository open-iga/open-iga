package domain

type OauthUser struct {
	Email     string
	FirstName string
	LastName  string
}

type ConsentDetails struct {
	AuthCodeURL string
	State       string
}

func NewConsentDetails(authCodeUrl string, state string) *ConsentDetails {
	return &ConsentDetails{
		AuthCodeURL: authCodeUrl,
		State:       state,
	}
}

func NewOauthUser(firstName string, lastname string, email string) *OauthUser {
	return &OauthUser{
		Email:     email,
		FirstName: firstName,
		LastName:  lastname,
	}
}
