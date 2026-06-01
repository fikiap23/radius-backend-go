package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
}

type AppConfig struct {
	Name     string
	Env      string
	LogLevel string
}

type HTTPConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type JWTConfig struct {
	SecretKey string
	Issuer    string
	Expiry    time.Duration
}

type OAuthConfig struct {
	StateExpiry         time.Duration
	AllowedRedirectURIs []string
	Google              OAuthProviderConfig
	GitHub              OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	v.SetEnvPrefix("RADIUS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	for _, key := range v.AllKeys() {
		_ = v.BindEnv(key)
	}

	if err := normalizeJWTExpiry(v); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	expiry, err := parseFlexibleDuration(v.GetString("jwt.expiry"))
	if err != nil {
		return nil, fmt.Errorf("jwt.expiry: %w", err)
	}
	cfg.JWT.Expiry = expiry

	oauthStateExpiry, err := parseFlexibleDuration(v.GetString("oauth.stateexpiry"))
	if err != nil {
		return nil, fmt.Errorf("oauth.stateexpiry: %w", err)
	}
	cfg.OAuth.StateExpiry = oauthStateExpiry
	normalizeAllowedRedirectURIs(v, &cfg.OAuth)

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "radius-backend")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.loglevel", "info")

	v.SetDefault("http.port", 8080)
	v.SetDefault("http.readtimeout", "15s")
	v.SetDefault("http.writetimeout", "15s")
	v.SetDefault("http.shutdowntimeout", "30s")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.maxopenconns", 25)
	v.SetDefault("database.maxidleconns", 10)
	v.SetDefault("database.connmaxlifetime", "5m")

	v.SetDefault("jwt.secretkey", "")
	v.SetDefault("jwt.expiry", "24h")
	v.SetDefault("jwt.issuer", "radius-backend")

	v.SetDefault("oauth.stateexpiry", "10m")
	v.SetDefault("oauth.allowedredirecturis", []string{})
	v.SetDefault("oauth.google.clientid", "")
	v.SetDefault("oauth.google.clientsecret", "")
	v.SetDefault("oauth.github.clientid", "")
	v.SetDefault("oauth.github.clientsecret", "")
}

func (c *Config) validate() error {
	if c.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database.password is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if c.JWT.SecretKey == "" {
		return fmt.Errorf("jwt.secretkey is required")
	}
	return nil
}

// normalizeJWTExpiry converts day-based values (7d) to hours so viper can unmarshal.
func normalizeJWTExpiry(v *viper.Viper) error {
	raw := strings.TrimSpace(v.GetString("jwt.expiry"))
	if raw == "" || !strings.HasSuffix(raw, "d") {
		return nil
	}

	expiry, err := parseFlexibleDuration(raw)
	if err != nil {
		return fmt.Errorf("jwt.expiry: %w", err)
	}
	v.Set("jwt.expiry", expiry.String())
	return nil
}

// parseFlexibleDuration supports Go durations (24h, 30m) and day suffix (7d).
func parseFlexibleDuration(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("duration is required")
	}

	if strings.HasSuffix(raw, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(raw, "d"))
		if err != nil || days <= 0 {
			return 0, fmt.Errorf("invalid day duration %q", raw)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", raw, err)
	}
	return d, nil
}

func normalizeAllowedRedirectURIs(v *viper.Viper, oauth *OAuthConfig) {
	raw := strings.TrimSpace(v.GetString("oauth.allowedredirecturis"))
	if raw == "" && len(oauth.AllowedRedirectURIs) == 1 && strings.Contains(oauth.AllowedRedirectURIs[0], ",") {
		raw = oauth.AllowedRedirectURIs[0]
	}
	if raw == "" {
		return
	}

	parts := strings.Split(raw, ",")
	uris := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			uris = append(uris, trimmed)
		}
	}
	oauth.AllowedRedirectURIs = uris
}
