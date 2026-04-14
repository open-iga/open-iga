package domain

type OauthUser struct {
	Email     string
	FirstName string
	LastName  string
}

func NewOauthUser(firstName string, lastname string, email string) *OauthUser {
	return &OauthUser{
		Email:     email,
		FirstName: firstName,
		LastName:  lastname,
	}
}
