package handlers

//Version Handler for the app
import (
	"encoding/json"
	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
	"net/http"
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
		BuildDate: appruntime.BuildTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		http.Error(
			w,
			"failed to encode Version response",
			http.StatusInternalServerError,
		)
		return
	}
}
