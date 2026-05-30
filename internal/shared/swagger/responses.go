package swagger

// Shared Swagger response schemas — reuse in @Success / @Failure annotations.

type Err struct {
	IsSuccess bool   `json:"isSuccess" example:"false"`
	Message   string `json:"message" example:"VALIDATION_ERROR"`
}

type HealthOK struct {
	Status string `json:"status" example:"ok"`
}

type UserData struct {
	ID     string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name   string `json:"name" example:"Jane Doe"`
	Email  string `json:"email" example:"jane@example.com"`
	Locale string `json:"locale" example:"en"`
}

type AuthData struct {
	AccessToken string   `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string   `json:"tokenType" example:"Bearer"`
	ExpiresIn   int64    `json:"expiresIn" example:"604800"`
	User        UserData `json:"user"`
}

type AuthOK struct {
	IsSuccess bool     `json:"isSuccess" example:"true"`
	Message   string   `json:"message" example:"OK"`
	Data      AuthData `json:"data"`
}

type UserOK struct {
	IsSuccess bool     `json:"isSuccess" example:"true"`
	Message   string   `json:"message" example:"OK"`
	Data      UserData `json:"data"`
}
