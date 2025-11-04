package main

import (
	"github.com/DiDinar5/mini-dex-aggregator/config"
	"github.com/DiDinar5/mini-dex-aggregator/internal/app"
)

func main() {
	cfg := config.Load()
	app.Run(*cfg)
}
