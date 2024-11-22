package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/auth"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/clients"
	openapiGenerated "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi"
	cars_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/cars-service"
	payment_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/payment-service"
	rental_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/rental-service"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/openapi"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/repository/kafka/retryqueue"
	circuit "github.com/rubyist/circuitbreaker"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := readConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	logger, err := initLogger(cfg)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	optCarsServiceClient := cars_service.WithHTTPClient(circuit.NewHTTPClient(0, 10, nil))
	carsServiceGeneratedClient, err := cars_service.NewClient(cfg.Services.Cars, optCarsServiceClient)
	if err != nil {
		return fmt.Errorf("init cars service client: %w", err)
	}
	carsServiceClient := clients.NewCarsServiceClient(carsServiceGeneratedClient, cfg.ServicePassword)

	optRentalServiceClient := rental_service.WithHTTPClient(circuit.NewHTTPClient(0, 10, nil))
	rentalServiceGeneratedClient, err := rental_service.NewClient(cfg.Services.Rental, optRentalServiceClient)
	if err != nil {
		return fmt.Errorf("init rental service client: %w", err)
	}
	rentalServiceClient := clients.NewRentalServiceClient(rentalServiceGeneratedClient)

	optPaymentServiceClient := payment_service.WithHTTPClient(circuit.NewHTTPClient(0, 10, nil))
	paymentServiceGeneratedClient, err := payment_service.NewClient(cfg.Services.Payment, optPaymentServiceClient)
	if err != nil {
		return fmt.Errorf("init rental service client: %w", err)
	}
	paymentServiceClient := clients.NewPaymentServiceClient(paymentServiceGeneratedClient, cfg.ServicePassword)

	retryQueueProducer, err := retryqueue.NewRetryQueueProducer(cfg.Kafka.Brokers, cfg.Kafka.CarsServiceRetryTopic, cfg.Kafka.PaymentServiceRetryTopic, logger)
	if err != nil {
		return fmt.Errorf("init retry queue producer: %w", err)
	}

	carsRetryQueueConsumer, err := retryqueue.NewCarsRetryQueueConsumer(ctx, retryQueueProducer, cfg.Kafka.Brokers, cfg.Kafka.CarsServiceRetryTopic, carsServiceClient, logger)
	if err != nil {
		return fmt.Errorf("init cars retry queue consumer: %w", err)
	}

	paymentRetryQueueConsumer, err := retryqueue.NewPaymentRetryQueueConsumer(ctx, retryQueueProducer, cfg.Kafka.Brokers, cfg.Kafka.PaymentServiceRetryTopic, paymentServiceClient, logger)
	if err != nil {
		return fmt.Errorf("init payment retry queue consumer: %w", err)
	}

	e := echo.New()
	e.Use(auth.CreateMiddleware(cfg.JWKsURL))
	server := openapi.New(carsServiceClient, paymentServiceClient, rentalServiceClient, retryQueueProducer)
	openapiGenerated.RegisterHandlers(e, server)

	go func() {
		<-ctx.Done()

		logger.Info("shutting down")

		e.Close()
		carsRetryQueueConsumer.Stop()
		paymentRetryQueueConsumer.Stop()
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

type config struct {
	Services        services
	Port            int
	LogLevel        string
	Kafka           kafka
	JWKsURL         string
	ServicePassword string
}

type services struct {
	Cars    string
	Rental  string
	Payment string
}

type kafka struct {
	Brokers                  []string
	CarsServiceRetryTopic    string
	PaymentServiceRetryTopic string
}
