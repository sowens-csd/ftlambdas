package main

import (
	"io"
	"net/http"

	"github.com/akrylysov/algnhsa"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	folktells "github.com/sowens-csd/folktells-server"
	"github.com/sowens-csd/folktells-server/awsproxy"
)

// main is called when a new lambda starts, so don't
// expect to have something done for every query here.
func main() {
	// init go-chi router
	r := chi.NewRouter()
	r.Route("/folk", func(r chi.Router) {

		r.Post("/", createFolk)
		r.Get("/", searchFolk)

		// Subrouters:
		r.Route("/{folkID}", func(r chi.Router) {
			r.Get("/", getFolk)
			r.Put("/", updateFolk)
			r.Delete("/", deleteFolk)
		})
	})
	// r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte(fmt.Sprintf("Triggered GET %s %s %s %s\n", r.Method, r.URL, r.Proto, r.URL.Path)))
	// 	render.Status(r, http.StatusNotFound)
	// })
	algnhsa.ListenAndServe(r, nil)
}

func createFolk(w http.ResponseWriter, r *http.Request) {
	proxyReq, ok := algnhsa.ProxyRequestFromContext(r.Context())
	if ok {
		ftCtx, errResp := awsproxy.NewFromContextAndJWT(r.Context(), awsproxy.Request(proxyReq))
		if nil != errResp {
			render.Status(r, 401)
			return
		}
		ftCtx.RequestLogger.Info().Msg("Create")
		bodyReader := r.Body
		defer bodyReader.Close()
		body, err := io.ReadAll(bodyReader)
		if nil != err {
			render.Status(r, http.StatusInternalServerError)
			return
		}
		ftCtx.RequestLogger.Info().Msg("About to AddManagedUser")
		folktells.AddManagedUser(ftCtx, string(body))
		w.Write([]byte("Created"))
		render.Status(r, 200)
	}

}

func searchFolk(w http.ResponseWriter, r *http.Request) {
	proxyReq, ok := algnhsa.ProxyRequestFromContext(r.Context())
	if ok {
		ftCtx, errResp := awsproxy.NewFromContextAndJWT(r.Context(), awsproxy.Request(proxyReq))
		if nil != errResp {
			render.Status(r, 401)
			return
		}
		ftCtx.RequestLogger.Info().Msg("search")
		w.Write([]byte("search"))
		render.Status(r, 200)
	}
}

func getFolk(w http.ResponseWriter, r *http.Request) {
	proxyReq, ok := algnhsa.ProxyRequestFromContext(r.Context())
	if ok {
		ftCtx, errResp := awsproxy.NewFromContextAndJWT(r.Context(), awsproxy.Request(proxyReq))
		if nil != errResp {
			render.Status(r, 401)
			return
		}
		ftCtx.RequestLogger.Info().Msg("get")
		w.Write([]byte("get"))
		render.Status(r, 200)
	}
}

func updateFolk(w http.ResponseWriter, r *http.Request) {
	proxyReq, ok := algnhsa.ProxyRequestFromContext(r.Context())
	if ok {
		ftCtx, errResp := awsproxy.NewFromContextAndJWT(r.Context(), awsproxy.Request(proxyReq))
		if nil != errResp {
			render.Status(r, 401)
			return
		}
		ftCtx.RequestLogger.Info().Msg("update")
		w.Write([]byte("update"))
		render.Status(r, 200)
	}
}

func deleteFolk(w http.ResponseWriter, r *http.Request) {
	proxyReq, ok := algnhsa.ProxyRequestFromContext(r.Context())
	if ok {
		ftCtx, errResp := awsproxy.NewFromContextAndJWT(r.Context(), awsproxy.Request(proxyReq))
		if nil != errResp {
			render.Status(r, 401)
			return
		}
		ftCtx.RequestLogger.Info().Msg("delete")
		w.Write([]byte("delete"))
		render.Status(r, 200)
	}
}
