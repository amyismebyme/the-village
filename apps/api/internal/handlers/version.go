package handlers

import (
	"net/http"

	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
)

type VersionResponse struct {
	Version   string `json:"version"`
	GitCommit string `json:"GitCommit"`
	BuildDate string `json:"buildDate"`
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	response := VersionResponse{
		Version:   appruntime.BuildVersion,
		GitCommit: appruntime.GitCommit,
		BuildDate: appruntime.BuildTime.Format(
			"2006-01-02T15:04:05Z07:00",
		),
	}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}
