package server

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"banner_rotation/internal/app"
	"github.com/gorilla/mux"
)

type Server struct {
	port   int
	host   string
	router *mux.Router
	server *http.Server
	app    *app.BannersApp
}

func NewServer(app *app.BannersApp, port int, host string) *Server {
	r := mux.NewRouter()
	r.HandleFunc("/banner", app.GetBanner).Methods("POST")
	r.HandleFunc("/add_site", app.AddSite).Methods("POST")
	r.HandleFunc("/add_banner", app.AddBanner).Methods("POST")
	r.Use(jsonHeaderMiddleware, loggingMiddleware)
	http.Handle("/", r)
	return &Server{
		port:   port,
		host:   host,
		router: r,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    net.JoinHostPort(s.host, strconv.Itoa(s.port)),
		Handler: s.router,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
