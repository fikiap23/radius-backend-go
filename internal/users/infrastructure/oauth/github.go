package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubProvider struct {
	cfg    config.OAuthProviderConfig
	oauth2 *oauth2.Config
}

func NewGitHubProvider(cfg config.OAuthProviderConfig) Provider {
	return &githubProvider{
		cfg: cfg,
		oauth2: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     github.Endpoint,
			Scopes:       []string{"read:user", "user:email"},
		},
	}
}

func (p *githubProvider) Name() entities.OAuthProvider {
	return entities.OAuthProviderGitHub
}

func (p *githubProvider) Enabled() bool {
	return p.cfg.ClientID != "" && p.cfg.ClientSecret != ""
}

func (p *githubProvider) AuthURL(state, redirectURI string) string {
	cfg := *p.oauth2
	cfg.RedirectURL = redirectURI
	return cfg.AuthCodeURL(state)
}

func (p *githubProvider) Exchange(ctx context.Context, code, redirectURI string) (*UserInfo, error) {
	cfg := *p.oauth2
	cfg.RedirectURL = redirectURI

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange github code: %w", err)
	}

	client := cfg.Client(ctx, token)

	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("fetch github user: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(userResp.Body, 512))
		return nil, fmt.Errorf("github user status %d: %s", userResp.StatusCode, strings.TrimSpace(string(body)))
	}

	var userPayload struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userPayload); err != nil {
		return nil, fmt.Errorf("decode github user: %w", err)
	}

	email := strings.ToLower(strings.TrimSpace(userPayload.Email))
	verified := email != ""

	if email == "" {
		email, verified, err = p.fetchPrimaryEmail(ctx, client)
		if err != nil {
			return nil, err
		}
	}

	if userPayload.ID == 0 || email == "" {
		return nil, fmt.Errorf("github user missing id or email")
	}

	name := strings.TrimSpace(userPayload.Name)
	if name == "" {
		name = userPayload.Login
	}
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	info := &UserInfo{
		ProviderUserID: strconv.FormatInt(userPayload.ID, 10),
		Email:          email,
		Name:           name,
		EmailVerified:  verified,
	}
	if userPayload.AvatarURL != "" {
		avatar := userPayload.AvatarURL
		info.AvatarURL = &avatar
	}

	return info, nil
}

func (p *githubProvider) fetchPrimaryEmail(ctx context.Context, client *http.Client) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", false, fmt.Errorf("build github emails request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("fetch github emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", false, fmt.Errorf("github emails status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", false, fmt.Errorf("decode github emails: %w", err)
	}

	for _, item := range emails {
		if item.Primary && item.Verified {
			return strings.ToLower(strings.TrimSpace(item.Email)), true, nil
		}
	}
	for _, item := range emails {
		if item.Verified {
			return strings.ToLower(strings.TrimSpace(item.Email)), true, nil
		}
	}
	for _, item := range emails {
		if item.Primary {
			return strings.ToLower(strings.TrimSpace(item.Email)), item.Verified, nil
		}
	}
	if len(emails) > 0 {
		return strings.ToLower(strings.TrimSpace(emails[0].Email)), emails[0].Verified, nil
	}

	return "", false, fmt.Errorf("github user has no email")
}
