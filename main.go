package main

import (
	"fmt"
	"github.com/supperdoggy/webproject/backend/internal/handlers"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	logger, _ := zap.NewDevelopment()
	// configure the songs directory name and port
	const port = 8080

	h := handlers.NewHandlers(logger)
	h.InitHandlers()

	fmt.Printf("Starting server on %v\n", port)

	// serve and log errors
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
