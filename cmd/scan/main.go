package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tariel-x/scan/internal/api"
	"github.com/tariel-x/scan/internal/scan"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func main() {
	c := dig.New()

	constructors := []interface{}{
		NewLogger,
		NewConfig,
		scan.NewScan,
		func(cfg *Config, s *scan.Scan, l *zap.Logger) (*api.Api, error) {
			return api.NewApi(cfg.Listen, s, l)
		},
	}

	for _, constructor := range constructors {
		if err := c.Provide(constructor); err != nil {
			fmt.Println(err)
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := c.Invoke(func(a *api.Api) error {
			return a.Run()
		}); err != nil {
			fmt.Println(err)
			cancel()
		}
	}()

	termc := make(chan os.Signal, 1)
	signal.Notify(termc, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-termc:
		cancel()
	case <-ctx.Done():
		cancel()
	}

	if err := c.Invoke(func(a *api.Api, s *scan.Scan, l *zap.Logger) error {
		l.Info("shutdown")

		if err := a.Stop(); err != nil {
			return err
		}

		s.Stop()

		return l.Sync()
	}); err != nil {
		fmt.Println(err)
		return
	}
}

func NewLogger(cfg *Config) (*zap.Logger, error) {
	if cfg.Debug {
		lcfg := zap.NewDevelopmentConfig()
		lcfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		return lcfg.Build()
	}
	return zap.NewProduction()
}
