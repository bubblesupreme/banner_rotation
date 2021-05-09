package server

import (
	"banner_rotation/internal/app"
	"context"
	"net"
	"net/http"
	"strconv"

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
	r.HandleFunc("/get_banner", app.GetBanner).Methods("POST")
	r.HandleFunc("/new_slot", app.AddSlot).Methods("POST")
	r.HandleFunc("/add_banner", app.AddBanner).Methods("POST")
	r.HandleFunc("/add_relation", app.AddRelation).Methods("POST")
	r.HandleFunc("/remove_banner", app.RemoveBanner).Methods("DELETE")
	r.HandleFunc("/remove_slot", app.RemoveSlot).Methods("DELETE")
	r.HandleFunc("/remove_relation", app.RemoveRelation).Methods("DELETE")
	r.HandleFunc("/click", app.Click).Methods("POST")
	r.HandleFunc("/show", app.Show).Methods("POST")
	r.HandleFunc("/all_banners", app.GetAllBanners).Methods("GET")
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
