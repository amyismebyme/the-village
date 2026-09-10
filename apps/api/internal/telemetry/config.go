package telemetry

import (
	"strconv"
	"strings"
)

type Config struct {
	Enabled         bool
	ServiceName     string
	ServiceVersion  string
	Environment     string
	Endpoint        string
	Insecure        bool
	Sampler         string
	SamplerArgument float64
}

func LoadFromEnv(getenv func(string) string) Config {
	return Config{
		Enabled:         getBool(getenv("OTEL_ENABLED"), false),
		ServiceName:     getEnv(getenv("OTEL_SERVICE_NAME"), "village-api"),
		ServiceVersion:  getEnv(getenv("OTEL_SERVICE_VERSION"), "dev"),
		Environment:     getEnv(getenv("OTEL_ENVIRONMENT"), "development"),
		Endpoint:        getEnv(getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), "http://localhost:4318/v1/traces"),
		Insecure:        getBool(getenv("OTEL_EXPORTER_OTLP_INSECURE"), true),
		Sampler:         strings.ToLower(getEnv(getenv("OTEL_TRACES_SAMPLER"), "parentbased_traceidratio")),
		SamplerArgument: getFloat(getenv("OTEL_TRACES_SAMPLER_ARG"), 1),
	}
}

func getEnv(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func getBool(value string, fallback bool) bool {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getFloat(value string, fallback float64) float64 {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
