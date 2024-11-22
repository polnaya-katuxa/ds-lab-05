package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/auth"
	openapiGenerated "github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/generated/openapi"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/logic"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/openapi"
	repositoryPostgres "github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/repository/postgres"
	"github.com/pressly/goose/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := readConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	logger, err := initLogger(cfg)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.Postgres.toDSN()))
	if err != nil {
		return fmt.Errorf("open postgres connection: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("up migrations: %w", err)
	}

	repo := repositoryPostgres.New(db)
	logic := logic.New(repo)

	e := echo.New()
	e.Use(auth.CreateMiddleware(cfg.JWKsURL, cfg.ServicePassword))
	server := openapi.New(logic)
	openapiGenerated.RegisterHandlers(e, server)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c

		logger.Info("shutting down")

		rawDB, err := db.DB()
		if err == nil {
			rawDB.Close()
		}

		e.Close()
	}()

	logger.Infow("starting service", "port", cfg.Port)

	g := new(errgroup.Group)
	g.Go(func() error {
		if err := e.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve echo server: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("errgroup: %w", err)
	}

	return nil
}

func readConfig() (*config, error) {
	cfgFile := flag.String("config", "/config.yaml", "path to config")
	flag.Parse()

	viper.SetConfigName(filepath.Base(*cfgFile))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Dir(*cfgFile))

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("read in config: %w", err)
	}

	var cfg config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &cfg, nil
}

func initLogger(cfg *config) (*zap.SugaredLogger, error) {
	lvl, err := zap.ParseAtomicLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("parse level: %w", err)
	}

	logConfig := zap.Config{
		Level:    lvl,
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			TimeKey:      "time",
			CallerKey:    "caller",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			EncodeTime:   zapcore.RFC3339TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		OutputPaths: []string{"stdout"},
	}

	logger, err := logConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return logger.Sugar(), nil
}

type db struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     int
}

func (d *db) toDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", d.Host,
		d.Port, d.User, d.Password, d.DBName)
}

type config struct {
	Postgres        db
	Port            int
	LogLevel        string
	JWKsURL         string
	ServicePassword string
}
