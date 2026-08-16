package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"garment-ppc/internal/config"
	"garment-ppc/internal/db"
	httpapi "garment-ppc/internal/http"
	"garment-ppc/internal/sim"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env", "../../.env")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}

	ticker := sim.New(pool, time.Duration(cfg.TickSeconds)*time.Second)
	ticker.Start()
	defer ticker.Stop()

	srv := httpapi.New(pool, cfg.WebOrigin, ticker)
	go func() {
		log.Printf("loom ppc api on %s  origin=%s  tick=%ds", cfg.ListenAddr, cfg.WebOrigin, cfg.TickSeconds)
		if err := srv.App().Listen(cfg.ListenAddr); err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	_ = srv.App().Shutdown()
}
