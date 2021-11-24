package main

import (
	"fmt"
	"github.com/supperdoggy/spotify-web-project/spotify-back/internal/handlers"
	service2 "github.com/supperdoggy/spotify-web-project/spotify-back/internal/service"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	logger, _ := zap.NewDevelopment()
	// configure the songs directory name and port
	const port = 8080
	// test
	service := service2.NewService(logger)
	h := handlers.NewHandlers(logger, service)
	h.InitHandlers()

	fmt.Printf("Starting server on %v\n", port)

	// serve and log errors
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
