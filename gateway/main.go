package main

import (
	"gateway/internal/router"

	"go.uber.org/zap"
)

func main() {
	log := zap.NewDevelopment()

	var srv router.HTTPServer
	srv.Init(log)
	if err := srv.Start(); err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}
}
