package handlers

import (
	"encoding/json"
	"net/http"

	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
)

type StatusResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	BuildTime string `json:"BuildTime"`
	GoVersion string `json:"go_version"`
	Uptime    string `json:"uptime"`
	StartedAt string `json:"started_at"`
	GitCommit string `json:"git_commit"`
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {

	response := StatusResponse{
		Status:    "running",
		Version:   appruntime.BuildVersion,
		BuildTime: appruntime.BuildTime.Format("2006-01-02T15:04:05Z07:00"),
		GoVersion: appruntime.GoVersion(),
		Uptime:    appruntime.Uptime().String(),
		StartedAt: appruntime.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		GitCommit: appruntime.GitCommit,
	}

	w.Header().Set("Content-Type", "application/json")


    if err := json.NewEncoder(w).Encode(response); err != nil {
    	http.Error(w, "failed to encode response", http.StatusInternalServerError)
    	return
    }
}
