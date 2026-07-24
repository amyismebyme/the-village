package model

type HealthStatus struct {
	Service     string `json:"service"`
	Status      string `json:"status"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Uptime      string `json:"uptime"`
}
