// Copyright 2016 Fraunhofer Institute for Applied Information Technology FIT

package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type router struct {
	*mux.Router
}

func newRouter() *router {
	return &router{mux.NewRouter().StrictSlash(false).SkipClean(true)}
}

func (r *router) get(path string, handler http.Handler) {
	r.Methods("GET").Path(path).Handler(handler)
	r.Methods("GET").Path(fmt.Sprintf("%s/", path)).Handler(handler)
}

func (r *router) post(path string, handler http.Handler) {
	r.Methods("POST").Path(path).Handler(handler)
	r.Methods("POST").Path(fmt.Sprintf("%s/", path)).Handler(handler)
}

func (r *router) put(path string, handler http.Handler) {
	r.Methods("PUT").Path(path).Handler(handler)
	r.Methods("PUT").Path(fmt.Sprintf("%s/", path)).Handler(handler)
}

func (r *router) delete(path string, handler http.Handler) {
	r.Methods("DELETE").Path(path).Handler(handler)
	r.Methods("DELETE").Path(fmt.Sprintf("%s/", path)).Handler(handler)
}

func (r *router) patch(path string, handler http.Handler) {
	r.Methods("PATCH").Path(path).Handler(handler)
	r.Methods("PATCH").Path(fmt.Sprintf("%s/", path)).Handler(handler)
}

func (r *router) head(path string, handler http.Handler) {
	r.Methods("HEAD").Path(path).Handler(handler)
	r.Methods("HEAD").Path(fmt.Sprintf("%s/", path)).Handler(handler)
}

func (r *router) options(path string, handler http.Handler) {
	r.Methods("OPTIONS").Path(path).Handler(handler)
	r.Methods("OPTIONS").Path(fmt.Sprintf("%s/", path)).Handler(handler)
}

// Add headers to handler's chain
func commonHeaders(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Headers for HTTP access control (CORS)
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization")

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
