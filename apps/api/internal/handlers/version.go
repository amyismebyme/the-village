package handlers

import (
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
)

type VersionResponse struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	response := VersionResponse{
		Version:   appruntime.BuildVersion,
		GitCommit: appruntime.GitCommit,
		BuildDate: appruntime.BuildTime.Format(
			"2006-01-02T15:04:05Z07:00",
		),
	}

	httputil.WriteJSON(
		w,
		http.StatusOK,
		response,
	)
}
