package handlers

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func BenchmarkHealthHandler(b *testing.B) {

	req := testutil.NewRequest(http.MethodGet, "/health")

	b.ReportAllocs()

	for b.Loop() {

		rr := testutil.NewRecorder()

		HealthHandler(rr, req)
	}
}
