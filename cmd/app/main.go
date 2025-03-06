package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	embedded "github.com/garnizeh/go-web-boilerplate/embedded"
	"github.com/garnizeh/go-web-boilerplate/internal/debug"
	"github.com/garnizeh/go-web-boilerplate/internal/web"
	"github.com/garnizeh/go-web-boilerplate/pkg/logger"
	"github.com/garnizeh/go-web-boilerplate/pkg/mailer"
	"github.com/garnizeh/go-web-boilerplate/pkg/securepass"
	"github.com/garnizeh/go-web-boilerplate/pkg/sessionmanager"
	"github.com/garnizeh/go-web-boilerplate/service"
	"github.com/garnizeh/go-web-boilerplate/storage"
	"github.com/garnizeh/go-web-boilerplate/storage/datastore"

	"github.com/ardanlabs/conf/v3"
)

var (
	build   = "develop"
	appName = "boilerplate"
)

func main() {
	prefix := strings.ToUpper(appName)

	// -------------------------------------------------------------------------
	// Logger Support

	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			// Do something here
			log.Info(ctx, "******* SEND ALERT *******")
		},
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, prefix, events)

	// -------------------------------------------------------------------------
	// Run

	ctx := context.Background()

	if err := run(ctx, log, prefix); err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger, prefix string) error {
	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Load Configuration

	cfg := struct {
		conf.Version
		Web struct {
			DomainName         string        `conf:"default:localhost"`
			Port               string        `conf:"default:3000"`
			BindAddress        string        `conf:"default:0.0.0.0"`
			ReadTimeout        time.Duration `conf:"default:5s"`
			WriteTimeout       time.Duration `conf:"default:10s"`
			IdleTimeout        time.Duration `conf:"default:120s"`
			ShutdownTimeout    time.Duration `conf:"default:20s"`
			CORSAllowedOrigins []string      `conf:"default:*,mask"`
		}
		Debug struct {
			Host string `conf:"default:0.0.0.0:3010"`
		}
		DBApp struct {
			DSN string `conf:"default:tmp/data/app.db"`
		}
		DBSessions struct {
			DSN string `conf:"default:tmp/data/sessions.db"`
		}
		Mailer struct {
			Host        string `conf:"default:localhost"`
			Port        int    `conf:"default:1025"`
			Username    string `conf:"default:test"`
			Password    string `conf:"default:test,mask"`
			Encryption  string `conf:"default:none"`
			FromAddress string `conf:"default:contact@boilerplate.com"`
			FromName    string `conf:"default:boilerplate Support"`
			JobsSize    int    `conf:"default:20"`
			ResultsSize int    `conf:"default:20"`
		}
		SecurePass struct {
			Time    uint32 `conf:"default:4"`
			SaltLen uint32 `conf:"default:32"`
			Memory  uint32 `conf:"default:65536"` // 64*1024
			Threads uint8  `conf:"default:4"`
			KeyLen  uint32 `conf:"default:256"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  appName,
		},
	}

	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		return fmt.Errorf("failed to parse config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "app started", "app name", appName, "build", cfg.Build)
	defer log.Info(ctx, "app finished")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("failed to generate config for output: %w", err)
	}

	log.Info(ctx, "startup", "config", out)

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support")

	dbApp, err := storage.NewDB(
		cfg.DBApp.DSN,
		datastore.Migrations,
		datastore.Factory,
	)
	if err != nil {
		return fmt.Errorf("failed to connect to the app database: %w", err)
	}
	defer func() {
		log.Info(ctx, "shutdown", "status", "stopping app database")
		if err := dbApp.Close(); err != nil {
			log.Error(ctx, "shutdown", "status", "failed to close the app database", "error", err)
		}
	}()

	if err := dbApp.RDBMS().Ping(); err != nil {
		return fmt.Errorf("failed to ping the app database: %w", err)
	}

	log.Info(ctx, "startup", "status", "connected to the app database")

	dbSessions, err := storage.NewDBSqlite(cfg.DBSessions.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to the sessions database: %w", err)
	}
	defer func() {
		log.Info(ctx, "shutdown", "status", "stopping sessions database")
		if err := dbSessions.Close(); err != nil {
			log.Error(ctx, "shutdown", "status", "failed to close the sessions database", "error", err)
		}
	}()

	if err := dbSessions.Ping(); err != nil {
		return fmt.Errorf("failed to ping the sessions database: %w", err)
	}

	if err := storage.MigrateSessions(dbSessions); err != nil {
		return fmt.Errorf("failed to migrate the sessions database: %w", err)
	}

	log.Info(ctx, "startup", "status", "connected to the sessions database")

	// -------------------------------------------------------------------------
	// Password Encryption Support

	log.Info(ctx, "startup", "status", "initializing password encryption support")

	securepass := securepass.New(
		cfg.SecurePass.Time,
		cfg.SecurePass.SaltLen,
		cfg.SecurePass.Memory,
		cfg.SecurePass.Threads,
		cfg.SecurePass.KeyLen,
	)

	// -------------------------------------------------------------------------
	// SMTP Support

	log.Info(ctx, "startup", "status", "initializing smtp support")

	scheme := "https"
	if cfg.Web.DomainName == "localhost" {
		scheme = "http"
	}
	baseURL := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%s", cfg.Web.DomainName, cfg.Web.Port),
	}

	mailer := mailer.New(mailer.Config{
		TemplatesFS: embedded.Mails(),
		Host:        cfg.Mailer.Host,
		Port:        cfg.Mailer.Port,
		Username:    cfg.Mailer.Username,
		Password:    cfg.Mailer.Password,
		Encryption:  cfg.Mailer.Encryption,
		FromAddress: cfg.Mailer.FromAddress,
		FromName:    cfg.Mailer.FromName,
		BaseURL:     baseURL.String(),
		JobsSize:    cfg.Mailer.JobsSize,
		ResultsSize: cfg.Mailer.ResultsSize,
	})

	// -------------------------------------------------------------------------
	// Service Support

	log.Info(ctx, "startup", "status", "initializing service support")

	service := service.New(securepass, mailer, dbApp)

	// -------------------------------------------------------------------------
	// Session Management Support

	log.Info(ctx, "startup", "status", "initializing session management support")

	sessionManager := sessionmanager.New(dbSessions)
	defer sessionManager.Close()

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug router started")

		mux, err := debug.Mux()
		if err != nil {
			panic(fmt.Sprintf("failed to create the debug mux: %v", err))
		}

		expvar.NewString("build").Set(build)

		server := http.Server{
			Addr:              cfg.Debug.Host,
			Handler:           mux,
			ReadHeaderTimeout: 3 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       120 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "shutdown", "status", "debug server closed", "error", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start Web Service

	log.Info(ctx, "startup", "status", "initializing web server")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverCfg := web.Config{
		AppName:     appName,
		DomainName:  cfg.Web.DomainName,
		Port:        cfg.Web.Port,
		BindAddress: cfg.Web.BindAddress,
		SessionsDSN: cfg.DBSessions.DSN,

		ReadTimeout:     cfg.Web.ReadTimeout,
		WriteTimeout:    cfg.Web.WriteTimeout,
		IdleTimeout:     cfg.Web.IdleTimeout,
		ShutdownTimeout: cfg.Web.ShutdownTimeout,

		CORSAllowedOrigins: cfg.Web.CORSAllowedOrigins,

		SessionManager: sessionManager,
	}

	serverErrors := make(chan error, 1)

	app := web.NewServer(serverCfg, service)

	go func() {
		log.Info(ctx, "startup", "status", "starting web server", "host", serverCfg.FullDomain())

		serverErrors <- app.Start(serverCfg.Address())
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "starting app shutdown", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "app shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		log.Info(ctx, "shutdown", "status", "stopping web server")

		if err := app.Shutdown(ctx); err != nil {
			app.Close()
			return fmt.Errorf("failed to stop the server gracefully: %w", err)
		}
	}

	return nil
}
