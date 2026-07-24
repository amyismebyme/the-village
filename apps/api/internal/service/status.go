package service

import (
	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
)

type StatusService struct {
	cfg *config.Config
}

func NewStatusService(cfg *config.Config) *StatusService {
	return &StatusService{
		cfg: cfg,
	}
}

func (s *StatusService) Status() model.HealthStatus {
	return model.HealthStatus{
		Service:     "village-api",
		Status:      "running",
		Version:     appruntime.BuildVersion,
		Environment: s.cfg.Environment,
		Uptime:      appruntime.Uptime().String(),
	}
}
