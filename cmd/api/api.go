package main

import (
	"fmt"
	"socialv2/docs"
	"socialv2/internal/auth"
	"socialv2/internal/mailer"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"net/http"
	"socialv2/internal/store"
	"time"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr        string
	env         string
	version     string
	db          dbConfig
	apiUrl      string
	frontendURL string
	mail        mailConfig
	auth        authConfig
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}
type basicConfig struct {
	username string
	password string
}

type mailConfig struct {
	mailtrap  mailtrapConfig
	exp       time.Duration
	fromEmail string
}

type mailtrapConfig struct {
	host     string
	username string
	password string
	port     string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthHandler)

		docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsUrl)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Get("/", app.listPostHandler)
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/", app.detailPostHandler)
				r.Patch("/", app.updatePostHandler)
				r.Delete("/", app.deletePostHandler)
			})
		})

		r.Route("/comments", func(r chi.Router) {
			r.Post("/", app.createCommentHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Get("/feeds", app.getUserFeedHandler)
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiUrl
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	app.logger.Info("Listening on " + app.config.addr)
	return srv.ListenAndServe()
}
