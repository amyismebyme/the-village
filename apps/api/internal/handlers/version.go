package handlers
//Version Handler for the app
import (
	"encoding/json"
	"net/http"
)

type VersionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {

	response := VersionResponse{
		Version:   "0.1.2",
		Commit:    "dev",
		BuildDate: "2026-07-13T00:00:00Z",
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
