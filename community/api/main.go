package main

import (
	"net/http"

	"github.com/akrylysov/algnhsa"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// main is called when a new lambda starts, so don't
// expect to have something done for every query here.
func main() {
	// init go-chi router
	r := chi.NewRouter()
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Your chi is strong"))
	})
	algnhsa.ListenAndServe(r, nil)
}

// apiResponse is the response to the API.
type apiResponse struct {
	Status      int    `json:"status_code,omitempty"`
	URL         string `json:"url,omitempty"`
	RequestBody string `json:"request_body,omitempty"`
}

// Render is used by go-chi-render to render the JSON response.
func (a apiResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, a.Status)
	return nil
}
