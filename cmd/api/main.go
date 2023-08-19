package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"movie-api/internal/data"
	"movie-api/internal/jsonlog"
	"movie-api/internal/mailer"
	"os"
	"sync"
	"time"
)

const Version = "1.0.0"

type config struct {
	port        int
	environment string
	db          struct {
		dns string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	logger *jsonlog.Logger
	config config
	models data.Models
	mailer *mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "api-port", 4000, "Server port for API")
	flag.StringVar(&cfg.environment, "api-environment", "development", "Specifies API env mode")
	flag.StringVar(&cfg.db.dns, "db-dns", os.Getenv("MOVIES_DB"), "Describes API db connection link")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "bf56d873e43eb7", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "fbe40984f0556d", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "greemlight.team@email.com", "SMTP sender")
	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}
	defer db.Close()
	logger.PrintInfo("Database connection pool successfully established!!!", nil)

	app := &application{
		logger: logger,
		config: cfg,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.db.dns)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
