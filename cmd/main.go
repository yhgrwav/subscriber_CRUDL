package main

import (
	"context"
	"errors"
	stdhttp "net/http"

	"os"
	"os/signal"
	"time"

	"testovoe_again/docs"
	"testovoe_again/internal/config"
	deliveryhttp "testovoe_again/internal/delivery/http"
	"testovoe_again/internal/logger"
	"testovoe_again/internal/repository"
	"testovoe_again/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// @title           Subscription Service API
// @version         1.0
// @description     сервис для управления подписками
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, closeLog, err := logger.NewLogger(cfg.Logger.Level)
	if err != nil {
		panic(err)
	}
	defer closeLog()
	defer log.Sync()

	e := echo.New()

	e.Validator = &deliveryhttp.Validator{Validater: validator.New()}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Server.ReadTimeout = cfg.HTTPReadTimeout()
	e.Server.WriteTimeout = cfg.HTTPWriteTimeout()

	db, err := repository.Connect(cfg.DB.PostgresDSN)
	if err != nil {
		log.Fatal("не удалось подключиться к базе данных", zap.Error(err))
	}
	db.SetMaxOpenConns(cfg.DB.MaxOpen)
	db.SetMaxIdleConns(cfg.DB.MaxIdle)
	db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetimeMin) * time.Minute)

	repo := repository.NewPostgresRepo(db, log)
	svc := service.NewSubscriptionService(log, repo)
	handler := deliveryhttp.NewHandler(log, svc)

	handler.Routing(e)
	docs.SwaggerInfo.Host = cfg.Swagger.Host
	docs.SwaggerInfo.BasePath = cfg.Swagger.BasePath

	go func() {
		if err := e.Start(":" + cfg.HTTP.AppPort); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			log.Fatal("выключение сервера...", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err.Error())
	}
}
