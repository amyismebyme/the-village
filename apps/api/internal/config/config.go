package config

// Package config manages application configuration.
// Manages different params like ports and timeouts and log settings
import (
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"os"
	"strconv"
	"time"
)

type WorkerConfig struct {
	Enabled bool
	Reddit  RedditWorkerConfig
}

type RedditWorkerConfig struct {
	Subreddit      string
	Limit          int
	IngestInterval time.Duration
}

type ExternalConfig struct {
	RequestTimeout time.Duration
	Reddit         RedditConfig
}

// Config holds all application configuration.
type Config struct {
	Port            string
	Environment     string
	LogLevel        string
	LogFormat       string
	ReadTimeout     time.Duration
	RequestTimeout  time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	Database        database.Config
	External        ExternalConfig
	Worker          WorkerConfig
}

type RedditConfig struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	UserAgent    string
	BaseURL      string
	AuthBaseURL  string
}

// Load reads environment variables.
func Load() Config {

	cfg := Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogFormat:       getEnv("LOG_FORMAT", "json"),
		ReadTimeout:     getDuration("READ_TIMEOUT", 10),
		RequestTimeout:  getDuration("REQUEST_TIMEOUT", 35),
		WriteTimeout:    getDuration("WRITE_TIMEOUT", 40),
		IdleTimeout:     getDuration("IDLE_TIMEOUT", 60),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 15),
		External: ExternalConfig{
			RequestTimeout: getDuration(
				"EXTERNAL_REQUEST_TIMEOUT",
				15,
			),

			Reddit: RedditConfig{
				Enabled: getBool(
					"REDDIT_ENABLED",
					false,
				),
				ClientID: getEnv(
					"REDDIT_CLIENT_ID",
					"",
				),
				ClientSecret: getEnv(
					"REDDIT_CLIENT_SECRET",
					"",
				),
				UserAgent: getEnv(
					"REDDIT_USER_AGENT",
					"",
				),
				BaseURL: getEnv(
					"REDDIT_BASE_URL",
					"https://oauth.reddit.com",
				),
				AuthBaseURL: getEnv(
					"REDDIT_AUTH_BASE_URL",
					"https://www.reddit.com",
				),
			},
		},

		Worker: WorkerConfig{
			Enabled: getBool(
				"WORKER_ENABLED",
				false,
			),

			Reddit: RedditWorkerConfig{
				Subreddit: getEnv(
					"REDDIT_INGEST_SUBREDDIT",
					"toronto",
				),

				Limit: getInt(
					"REDDIT_INGEST_LIMIT",
					25,
				),

				IngestInterval: getDuration(
					"REDDIT_INGEST_INTERVAL",
					300,
				),
			},
		},

		Database: database.Config{
			Host: getEnv("DB_HOST", "localhost"),
			Port: getInt("DB_PORT", 5432),

			User:     getEnv("DB_USER", "village"),
			Password: getEnv("DB_PASSWORD", "village"),
			Name:     getEnv("DB_NAME", "village"),

			SSLMode: getEnv("DB_SSLMODE", "disable"),

			MaxConns:          int32(getInt("DB_MAX_CONNS", 10)),
			MinConns:          int32(getInt("DB_MIN_CONNS", 1)),
			MaxConnLifetime:   getDuration("DB_MAX_CONN_LIFETIME", 3600),
			MaxConnIdleTime:   getDuration("DB_MAX_CONN_IDLE_TIME", 300),
			HealthCheckPeriod: getDuration("DB_HEALTH_CHECK_PERIOD", 60),
			QueryTimeout: getDuration(
				"DB_QUERY_TIMEOUT",
				30,
			),
		},
	}

	return cfg
}

func getEnv(key, defaultValue string) string {

	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func getDuration(key string, defaultSeconds int) time.Duration {

	value := os.Getenv(key)

	if value == "" {
		return time.Duration(defaultSeconds) * time.Second
	}

	seconds, err := strconv.Atoi(value)

	if err != nil {
		return time.Duration(defaultSeconds) * time.Second
	}

	return time.Duration(seconds) * time.Second
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return v
}

func getBool(
	key string,
	defaultValue bool,
) bool {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
