package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/rate-limit/app/api"
	"github.com/rate-limit/config"
)

func main() {
	config.InializeConfig()
	ServeRequest(api.GetRoutes())
}

func ServeRequest(configuredRoutes http.Handler) {
	port := config.GetConfig().Server.Port
	fmt.Println("########## SERVER STARTED ########## ", port)
	server := &http.Server{
		Addr: port,
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{"OPTIONS", "GET", "POST", "PUT", "PATCH"}),
			handlers.MaxAge(600),
		)(configuredRoutes),
	}

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("server is shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		server.Shutdown(ctx)
		close(done)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("server shutdown with error: ", err.Error())
	}
	<-done
}
