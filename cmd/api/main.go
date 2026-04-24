package main

import (
	"context"
	"log"
	"portal-system/internal/app"
	"portal-system/internal/composer"
	"portal-system/internal/platform/logger"
)

// Run with: go run github.com/air-verse/air@latest
func main() {
	logger.InitLogger()
	application, err := composer.Composer()
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := application.NewGRPCServer()
	gatewayHandler, err := application.NewGatewayMux(context.Background(), ":"+application.Config.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, 2)

	go func() {
		log.Printf("grpc listening on :%s", application.Config.GRPCPort)
		errCh <- app.RunGRPCServer(":"+application.Config.GRPCPort, grpcServer)
	}()

	go func() {
		log.Printf("gateway listening on :%s", application.Config.HTTPPort)
		errCh <- app.RunGatewayServer(":"+application.Config.HTTPPort, gatewayHandler)
	}()

	log.Fatal(<-errCh)
}
