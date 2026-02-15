package main

import (
	db2 "socialv3/internal/db"
	"socialv3/internal/env"
	"socialv3/internal/store"
	"time"

	"go.uber.org/zap"
)

const version = "1.0.0"

//	@title			Social API
//	@version		1.0
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
		env:     env.GetString("ENV", "development"),
		addr:    env.GetString("ADDR", ":8000"),
		baseUrl: env.GetString("BASE_URL", "http://127.0.0.1:8000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://social_user:social_password@localhost:5432/social_db?sslmode=disable"),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 5),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 5),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		mail: mailConfig{
			exp: time.Hour * 24 * 3,
		},
	}

	// logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db2.New(cfg.db.addr, cfg.db.maxIdleConns, cfg.db.maxOpenConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("Connected to database")
	storage := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  storage,
		logger: logger,
	}

	logger.Fatal(app.run(app.mount()))
}
