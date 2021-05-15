package server

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/bubblesupreme/banner_rotation/internal/app"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type Server struct {
	port   int
	router *mux.Router
	server *http.Server
	app    *app.BannersApp
}

func NewServer(app *app.BannersApp, port int) *Server {
	r := mux.NewRouter()
	r.HandleFunc("/get_banner", app.GetBanner).Methods("POST")
	r.HandleFunc("/banner", app.AddBanner).Methods("POST")
	r.HandleFunc("/banner", app.RemoveBanner).Methods("DELETE")

	r.HandleFunc("/slot", app.AddSlot).Methods("POST")
	r.HandleFunc("/slot", app.RemoveSlot).Methods("DELETE")

	r.HandleFunc("/relation", app.AddRelation).Methods("POST")
	r.HandleFunc("/relation", app.RemoveRelation).Methods("DELETE")

	r.HandleFunc("/group", app.AddGroup).Methods("POST")
	r.HandleFunc("/group", app.RemoveGroup).Methods("DELETE")

	r.HandleFunc("/click", app.Click).Methods("POST")
	r.HandleFunc("/show", app.Show).Methods("POST")
	r.HandleFunc("/all_banners", app.GetAllBanners).Methods("GET")
	r.HandleFunc("/all_groups", app.GetAllGroups).Methods("GET")

	r.Use(jsonHeaderMiddleware, loggingMiddleware)
	http.Handle("/", r)
	return &Server{
		port:   port,
		router: r,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    net.JoinHostPort("", strconv.Itoa(s.port)),
		Handler: s.router,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	log.Info("starting server on " + s.server.Addr)

	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("server was stopped")
	return s.server.Shutdown(ctx)
}
