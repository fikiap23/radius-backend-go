package dto

type SSOAuthURLResult struct {
	AuthURL string `json:"authUrl"`
	State   string `json:"state"`
}

type GoogleSSOAuthURLInput struct {
	RedirectURI string `query:"redirect_uri" doc:"OAuth callback URL registered with Google" required:"true"`
}

type GitHubSSOAuthURLInput struct {
	RedirectURI string `query:"redirect_uri" doc:"OAuth callback URL registered with GitHub" required:"true"`
}

type GoogleSSOCallbackInput struct {
	Body struct {
		Code  string `json:"code" doc:"Authorization code from Google" minLength:"1"`
		State string `json:"state" doc:"State returned from authorize step" minLength:"1"`
	}
}

type GitHubSSOCallbackInput struct {
	Body struct {
		Code  string `json:"code" doc:"Authorization code from GitHub" minLength:"1"`
		State string `json:"state" doc:"State returned from authorize step" minLength:"1"`
	}
}
