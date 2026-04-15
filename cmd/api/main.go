package main

import (
	"expvar"
	"runtime"
	"socialv3/internal/auth"
	db2 "socialv3/internal/db"
	"socialv3/internal/env"
	"socialv3/internal/mailer"
	"socialv3/internal/ratelimiter"
	"socialv3/internal/store"
	cache2 "socialv3/internal/store/cache"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const version = "1.0.0"

//	@title			Social API
//	@version		1.0.0
//	@description	API untuk aplikasi sosial: manajemen users, posts, follow/unfollow, dan feed.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	https://github.com/your-org/socialv3
//	@contact.email	support@example.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Bearer token untuk autentikasi
func main() {
	cfg := config{
		env:         env.GetString("ENV", "development"),
		addr:        env.GetString("ADDR", ":8000"),
		baseUrl:     env.GetString("BASE_URL", "http://127.0.0.1:8000"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://social_user:social_password@localhost:5432/social_db?sslmode=disable"),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 5),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 5),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		mail: mailConfig{
			exp: time.Hour * 24 * 3,
			config: mailer.MailtrapConfig{
				APIKey:         env.GetString("MAIL_MAILTRAP_API_KEY", "d8862be33a858d80236678e306c71a7b"),
				FromEmail:      env.GetString("MAIL_MAILTRAP_FROM_EMAIL", mailer.FromEmail),
				FromName:       env.GetString("MAIL_MAILTRAP_FROM_NAME", mailer.FromName),
				SandboxInboxID: env.GetString("MAIL_MAILTRAP_SANDBOX_ID", "231735"),
			},
			// mailer: mailer.NewMailtrapMailer(mailer.MailtrapConfig{
			// 	FromEmail: env.GetString("MAIL_FROM_EMAIL", "noreply@example.com"),
			// 	FromName:  env.GetString("MAIL_FROM_NAME", "Example"),
			// 	APIKey:    env.GetString("MAIL_API_KEY", "07e6821ad0e59517566bc87ed6f9f160"),
			// 	SandboxInboxID: env.GetString("MAIL_SANDBOX_INBOX_ID", "231735"),
			// }),
		},
		auth: authConfig{
			basic: authBasicConfig{
				username: env.GetString("AUTH_BASIC_USERNAME", "root"),
				password: env.GetString("AUTH_BASIC_PASSWORD", "root"),
			},
			jwt: authJWTConfig{
				secret: env.GetString("AUTH_JWT_SECRET", "secret"),
				exp:    time.Hour * 24 * 7,
				issuer: env.GetString("AUTH_JWT_ISSUER", "socialv3"),
			},
		},
		redis: redisConfig{
			addr:     env.GetString("REDIS_ADDR", "127.0.0.1:6379"),
			password: env.GetString("REDIS_PASSWORD", ""),
			db:       env.GetInt("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", false),
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUEST_COUNT", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATELIMITER_ENABLED", true),
		},
	}

	// logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// redis
	var rdb *redis.Client
	if cfg.redis.enabled {
		rdb = cache2.NewRedisClient(cfg.redis.addr, cfg.redis.password, cfg.redis.db)
		logger.Info("redis cache established")
	}
	cacheStorage := cache2.NewRedisStorage(rdb)

	// db
	db, err := db2.New(cfg.db.addr, cfg.db.maxIdleConns, cfg.db.maxOpenConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("Connected to database")
	storage := store.NewStorage(db)

	// mailer
	mailer_ := mailer.NewMailtrapMailerWithConfig(cfg.mail.config)

	// jwt
	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.jwt.secret, cfg.auth.jwt.issuer, cfg.auth.jwt.issuer)

	// ratelimiter
	ratelimit := ratelimiter.NewFixedWindowRateLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	app := &application{
		config:        cfg,
		store:         storage,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailer_,
		authenticator: jwtAuthenticator,
		rateLimiter:   ratelimit,
	}

	// metrics
	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	logger.Fatal(app.run(app.mount()))
}
