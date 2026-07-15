package main
//main class that go uses to execute application. Everything branches off from here
import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
)

func main() {

    cfg := config.Load()

	httpServer := server.NewHTTPServer(cfg)

	log.Println("===================================")
	log.Println("Village API Starting")
	log.Println("Environment :", cfg.Environment)
	log.Println("Port        :", cfg.Port)
	log.Println("===================================")

	go func() {

		if err := httpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Println("Server started successfully.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutdown signal received.")
    log.Println("Total upTime: ", server.Uptime())


	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)

	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server shutdown complete.")
}
