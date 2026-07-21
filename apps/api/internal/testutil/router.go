package testutil

import (
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/server"
)

func NewRouter() http.Handler {

	return server.NewRouter()

}
