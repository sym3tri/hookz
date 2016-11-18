package server

import (
	"net/http"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/pkg/httputil"
	httprouter "github.com/julienschmidt/httprouter"
)

type Server struct {
	log       *logrus.Entry
	endpoints []Endpoint
	Version   string
}

func New(logger *logrus.Entry, endpoints []Endpoint) *Server {
	return &Server{
		log:       logger,
		endpoints: endpoints,
	}
}

func (s *Server) HTTPHandler() http.Handler {
	//checks := []health.Checkable{}
	//for _, h := range s.wizHandlers {
	//checks = append(checks, h.Health()...)
	//}

	r := httprouter.New()
	//r.Handler("GET", "/health", health.Checker{
	//Checks: checks,
	//})

	r.HandlerFunc("GET", "/version", s.versionHandler)
	r.HandlerFunc("GET", "/metrics", s.metricsHandler)
	r.HandlerFunc("GET", "/", s.indexHandler)
	r.NotFound = s.notFoundHandler

	for _, e := range s.endpoints {
		s.log.WithField("endpoint", e.Name).Info("adding endpoint path")
		r.HandlerFunc("POST", path.Join("/endpoints", e.Path), makeEndpointHandler(e))
		r.HandlerFunc("GET", path.Join("/endpoints", e.Path), makeEndpointInfo(e))
	}

	//for _, h := range s.wizHandlers {
	//s.logger.WithField("handlerPath", h.Path()).Info("Adding handler path")
	//mux.Handle(h.Path(), h)
	//}

	return http.Handler(r)
}

func makeEndpointInfo(e Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSONResponse(w, http.StatusOK, e.Sanitize())
	}
}

func makeEndpointHandler(e Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(e.Name))
	}
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hookz"))
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("TODO: put prometheus metrics here"))
}

func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSONResponse(w, http.StatusOK, struct {
		Version string `json:"version"`
	}{
		Version: s.Version,
	})
}

func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	s.log.WithField("path", r.URL.Path).Debug("Path not found")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}
