package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type RetryPolicy struct {
	MaxRetries   int
	InitialDelay time.Duration
}

type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

type GoogleOAuthConfig struct {
	GoogleCredentialsJSON string
	GoogleAccessToken     string
	GoogleRefreshToken    string
	GoogleEmailSender     string
}

type ZeroBounceConfig struct {
	ApiKey string
}

type Config struct {
	QueueURL    string
	RedisURL    string
	ProviderMap map[string]string
	SenderMap   map[string]string
	RetryPolicy RetryPolicy
	WorkerCount int
	Validator   struct{ DisposableDomains []string }

	SMTP        SMTPConfig
	GoogleOAuth GoogleOAuthConfig

	QuotaScoreThreshold float64
	QuotaScaleFactor    float64

	ZeroBounce ZeroBounceConfig
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // ignore errors if .env is missing

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigName("config")
	v.AddConfigPath(".")

	// Set defaults
	v.SetDefault("WORKER_COUNT", 5)
	v.SetDefault("RETRY_POLICY_MAX_RETRIES", 3)
	v.SetDefault("RETRY_POLICY_INITIAL_DELAY", "1s")
	v.SetDefault("QUOTA_SCORE_THRESHOLD", 0.8)
	v.SetDefault("QUOTA_SCALE_FACTOR", 1.5)

	v.BindEnv("QUOTA_SCORE_THRESHOLD")
	v.BindEnv("QUOTA_SCALE_FACTOR")

	v.BindEnv("GOOGLE_CREDENTIALS_JSON")
	v.BindEnv("GOOGLE_ACCESS_TOKEN")
	v.BindEnv("GOOGLE_REFRESH_TOKEN")
	v.BindEnv("GOOGLE_EMAIL_SENDER")
	v.BindEnv("ZERO_BOUNCE_API_KEY")

	// Unmarshal values
	cfg := &Config{}
	cfg.QueueURL = v.GetString("QUEUE_URL")
	cfg.RedisURL = v.GetString("REDIS_URL")

	cfg.ProviderMap = v.GetStringMapString("PROVIDER_MAP")
	cfg.SenderMap = v.GetStringMapString("TENANT_SENDER_MAP")

	cfg.WorkerCount = v.GetInt("WORKER_COUNT")

	cfg.RetryPolicy.MaxRetries = v.GetInt("RETRY_POLICY_MAX_RETRIES")
	d, _ := time.ParseDuration(v.GetString("RETRY_POLICY_INITIAL_DELAY"))
	cfg.RetryPolicy.InitialDelay = d
	cfg.Validator.DisposableDomains = v.GetStringSlice("VALIDATOR_DISPOSABLE_DOMAINS")

	cfg.SMTP.Host = v.GetString("SMTP_HOST")
	cfg.SMTP.Port = v.GetString("SMTP_PORT")
	cfg.SMTP.User = v.GetString("SMTP_USER")
	cfg.SMTP.Pass = v.GetString("SMTP_PASS")
	cfg.SMTP.From = v.GetString("SMTP_FROM")

	cfg.QuotaScaleFactor = v.GetFloat64("QUOTA_SCORE_THRESHOLD")
	cfg.QuotaScoreThreshold = v.GetFloat64("QUOTA_SCORE_THRESHOLD")

	cfg.GoogleOAuth.GoogleCredentialsJSON = v.GetString("GOOGLE_CREDENTIALS_JSON")
	cfg.GoogleOAuth.GoogleAccessToken = v.GetString("GOOGLE_ACCESS_TOKEN")
	cfg.GoogleOAuth.GoogleRefreshToken = v.GetString("GOOGLE_REFRESH_TOKEN")
	cfg.GoogleOAuth.GoogleEmailSender = v.GetString("GOOGLE_EMAIL_SENDER")

	cfg.ZeroBounce.ApiKey = v.GetString("ZERO_BOUNCE_API_KEY")

	return cfg, nil
}
