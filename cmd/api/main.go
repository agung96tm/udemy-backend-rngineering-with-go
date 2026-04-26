package main

import (
	"socialv2/internal/auth"
	"socialv2/internal/db"
	"socialv2/internal/env"
	"socialv2/internal/mailer"
	"socialv2/internal/store"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

const version = "1.0"

//	@title			Social API
//	@description	API for Social
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		petstore.swagger.io
//	@BasePath	/v1

// @securityDefinitions.apiKey	APIKeyAuth
// @in							header
// @name						Authorization
// @description				Your API key for authentication
func main() {
	env.InitEnv()
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiUrl:      env.GetString("API_URL", "http://localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:8000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://user:password@host:port/db?sslmode=disable"),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 10),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 10),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		version: env.GetString("VERSION", "latest"),
		env:     env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, // 3days
			fromEmail: env.GetString("MAIL_FROM_EMAIL", "agung@example.com"),
			mailtrap: mailtrapConfig{
				host:     env.GetString("MAILTRAP_HOST", "sandbox.smtp.mailtrap.io"),
				port:     env.GetString("MAILTRAP_PORT", "25"),
				username: env.GetString("MAILTRAP_USERNAME", "42730f78202ad6"),
				password: env.GetString("MAILTRAP_PASSWORD", "116fcda2659b10"),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				username: env.GetString("AUTH_BASIC_USERNAME", ""),
				password: env.GetString("AUTH_BASIC_PASSWORD", ""),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", ""),
				exp:    time.Hour * 24 * 7, // 7 days
				iss:    "socialv2",
			},
		},
	}

	// logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// db
	datab, err := db.NewDB(cfg.db.addr, cfg.db.maxIdleConns, cfg.db.maxOpenConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer datab.Close()
	strg := store.NewStorage(datab)
	logger.Info("database connection pool established")

	// mail
	mailr2 := mailer.NewMailtrap(
		cfg.mail.mailtrap.host,
		cfg.mail.mailtrap.username,
		cfg.mail.mailtrap.password,
		cfg.mail.mailtrap.port,
	)

	// auth
	auths := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	app := &application{
		config:        cfg,
		store:         strg,
		logger:        logger,
		mailer:        mailr2,
		authenticator: auths,
	}

	logger.Fatalln(app.run(app.mount()))
}
