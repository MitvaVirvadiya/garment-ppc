package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	ListenAddr  string
	WebOrigin   string
	TickSeconds int
}

func Load() (Config, error) {
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8014"
	}
	origin := os.Getenv("WEB_ORIGIN")
	if origin == "" {
		origin = "http://localhost:3014"
	}
	tick := 4
	if raw := os.Getenv("TICK_SECONDS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 && n <= 60 {
			tick = n
		}
	}
	return Config{
		DatabaseURL: url,
		ListenAddr:  addr,
		WebOrigin:   origin,
		TickSeconds: tick,
	}, nil
}
