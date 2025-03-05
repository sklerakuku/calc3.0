package main

import (
	"log"

	"github.com/sklerakuku/calc3.0/internal/agent"
	"github.com/sklerakuku/calc3.0/internal/config"
)

func main() {
	cfg := config.Load()
	a := agent.New(cfg)

	log.Printf("Agent starting with %d computing power", cfg.ComputingPower)
	a.Start()
}
