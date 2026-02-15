package main

import (
	"fmt"
	"net/http"
	"socialv3/internal/store"
	"time"

	"socialv3/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
}

type config struct {
	addr    string
	db      dbConfig
	env     string
	baseUrl string
	mail    mailConfig
}

type mailConfig struct {
	exp time.Duration
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
	r.Get("/v1/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(docsUrl), //The url pointing to API definition
	))
	r.Get("/v1/health", app.healthCheckHandler)

	r.Route("/v1/posts", func(r chi.Router) {
		r.Post("/", app.createPostHandler)

		r.Route("/{postID}", func(r chi.Router) {
			r.Use(app.postsContextMiddleware)

			r.Get("/", app.getPostHandler)
			r.Put("/", app.updatePostHandler)
			r.Delete("/", app.deletePostHandler)
		})
	})

	r.Route("/v1/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(app.userContextMiddleware)

			r.Get("/", app.getUserHandler)
			r.Put("/follow", app.followUserHandler)
			r.Put("/unfollow", app.unfollowUserHandler)
		})

		r.Group(func(r chi.Router) {
			r.Get("/feed", app.getUserFeedHandler)
		})
	})

	r.Route("/v1/authentication", func(r chi.Router) {
		r.Post("/user", app.authRegisterHandler)
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	//
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.baseUrl
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)
	return srv.ListenAndServe()
}
