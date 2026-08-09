package main

import (
	"butterfly-monitor/internal/starter"

	"github.com/pwh19920920/butterfly/pkg/server"
)

func main() {
	_, _ = starter.InitButterflyAdmin()
	server.StartHttpServer()
}
