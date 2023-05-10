package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"movies-api/internal/jsonlog"
	"movies-api/internal/mailer"
	"movies-api/internal/models/acttokens"
	"movies-api/internal/models/movies"
	"movies-api/internal/models/permissions"
	"movies-api/internal/models/users"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
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
	cors struct {
		trustedOrigins []string
	}
}

type app struct {
	config config
	logger *jsonlog.Logger
	err    CustomError
	mailer mailer.Mailer
	wg     sync.WaitGroup

	movieService       *movies.MovieService
	userService        *users.UserService
	actTokenService    *acttokens.ActTokenService
	permissionsService *permissions.PermissionsService
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 5000, "Your server port")
	flag.StringVar(&cfg.env, "env", "dev", "Your environment: dev|prod|stage")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://greenlight:pa55word@localhost/movies-api?sslmode=disable", "PostgreSQL connect url")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 50, "PostgreSQL maximum open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 50, "PostgreSQL maximum idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL maximum connections idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 50, "Rate limiter requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 1000, "Rate limiter burst requests")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enabled rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "94984a9872e2b4", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-pass", "932626d34748d6", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Movies API <no-reply@moviesapi.net>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version: \t%s\n", version)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)

	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("Connected to DB", nil)

	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &app{
		config:             cfg,
		err:                CustomError{logger: logger},
		logger:             jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		mailer:             mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		movieService:       movies.NewMovieService(db),
		userService:        users.NewUserService(db),
		actTokenService:    acttokens.NewActTokenService(db),
		permissionsService: permissions.NewPermissionsService(db),
	}

	err = app.serve()

	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	dur, err := time.ParseDuration(cfg.db.maxIdleTime)

	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(dur)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}
