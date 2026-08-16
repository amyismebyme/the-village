package httputil

import "net/http"

type ResponseRecorder struct {
	http.ResponseWriter
	Status      int
	WroteHeader bool
}

func NewResponseRecorder(
	w http.ResponseWriter,
) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		Status:         http.StatusOK,
	}
}

func (r *ResponseRecorder) WriteHeader(code int) {
	if r.WroteHeader {
		return
	}

	r.Status = code
	r.WroteHeader = true

	r.ResponseWriter.WriteHeader(code)
}

func (r *ResponseRecorder) Write(
	body []byte,
) (int, error) {
	if !r.WroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(body)
}

func (r *ResponseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
