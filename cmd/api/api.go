package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"socialv3/internal/auth"
	"socialv3/internal/mailer"
	"socialv3/internal/ratelimiter"
	"socialv3/internal/store"
	"socialv3/internal/store/cache"
	"syscall"
	"time"

	"socialv3/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	cacheStorage  cache.Storage
	rateLimiter   ratelimiter.Limiter
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	baseUrl     string
	frontendURL string
	mail        mailConfig
	auth        authConfig
	redis       redisConfig
	rateLimiter ratelimiter.Config
}

type redisConfig struct {
	addr     string
	password string
	db       int
	enabled  bool
}

type authConfig struct {
	basic authBasicConfig
	jwt   authJWTConfig
}

type authJWTConfig struct {
	secret string
	exp    time.Duration
	issuer string
}

type authBasicConfig struct {
	username string
	password string
}

type mailConfig struct {
	exp    time.Duration
	config mailer.MailtrapConfig
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
	r.Use(app.RateLimiterMiddleware)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	r.Get("/v1/health", app.healthCheckHandler)

	r.With(app.BasicAuthMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)
	docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
	r.Get("/v1/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(docsUrl), //The url pointing to API definition
	))

	r.Route("/v1/posts", func(r chi.Router) {
		r.Use(app.authJWTMiddleware)
		r.Post("/", app.createPostHandler)

		r.Route("/{postID}", func(r chi.Router) {
			r.Use(app.postsContextMiddleware)

			r.Get("/", app.getPostHandler)
			r.Put("/", app.checkPostOwnership("moderator", app.updatePostHandler))
			r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
		})
	})

	r.Route("/v1/users", func(r chi.Router) {
		r.Put("/activate/{token}", app.activateUserHandler)

		r.Route("/{userID}", func(r chi.Router) {
			r.Use(app.authJWTMiddleware)

			r.Get("/", app.getUserHandler)
			r.Put("/follow", app.followUserHandler)
			r.Put("/unfollow", app.unfollowUserHandler)
		})

		r.Group(func(r chi.Router) {
			r.Use(app.authJWTMiddleware)
			r.Get("/feed", app.getUserFeedHandler)
		})
	})

	r.Route("/v1/authentication", func(r chi.Router) {
		r.Post("/user", app.authRegisterHandler)
		r.Post("/token", app.createTokenHandler)
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

	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("shutting down server", "signal", s.String())
		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}
	app.logger.Infow("server has stopped", "addr", app.config.addr)
	return nil
}
