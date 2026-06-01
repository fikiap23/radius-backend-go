package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/users/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleProvider struct {
	cfg    config.OAuthProviderConfig
	oauth2 *oauth2.Config
}

func NewGoogleProvider(cfg config.OAuthProviderConfig) Provider {
	return &googleProvider{
		cfg: cfg,
		oauth2: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"openid", "email", "profile"},
		},
	}
}

func (p *googleProvider) Name() domain.OAuthProvider {
	return domain.OAuthProviderGoogle
}

func (p *googleProvider) Enabled() bool {
	return p.cfg.ClientID != "" && p.cfg.ClientSecret != ""
}

func (p *googleProvider) AuthURL(state, redirectURI string) string {
	cfg := *p.oauth2
	cfg.RedirectURL = redirectURI
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *googleProvider) Exchange(ctx context.Context, code, redirectURI string) (*UserInfo, error) {
	cfg := *p.oauth2
	cfg.RedirectURL = redirectURI

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange google code: %w", err)
	}

	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("google userinfo status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
	}

	if payload.ID == "" || payload.Email == "" {
		return nil, fmt.Errorf("google userinfo missing id or email")
	}

	info := &UserInfo{
		ProviderUserID: payload.ID,
		Email:          strings.ToLower(strings.TrimSpace(payload.Email)),
		Name:           strings.TrimSpace(payload.Name),
		EmailVerified:  payload.VerifiedEmail,
	}
	if payload.Picture != "" {
		picture := payload.Picture
		info.AvatarURL = &picture
	}
	if info.Name == "" {
		info.Name = strings.Split(info.Email, "@")[0]
	}

	return info, nil
}
