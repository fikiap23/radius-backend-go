package dto

type RegisterInput struct {
	Body struct {
		Name     string `json:"name" doc:"Display name" minLength:"2" maxLength:"255"`
		Email    string `json:"email" doc:"Email address" format:"email"`
		Password string `json:"password" doc:"Password" minLength:"8" maxLength:"72"`
	}
}

type LoginInput struct {
	Body struct {
		Email    string `json:"email" doc:"Email address" format:"email"`
		Password string `json:"password" doc:"Password" minLength:"8" maxLength:"72"`
	}
}

type AuthResult struct {
	AccessToken string      `json:"accessToken"`
	TokenType   string      `json:"tokenType"`
	ExpiresIn   int64       `json:"expiresIn"`
	User        UserProfile `json:"user"`
}
