package main

import (
	"context"
	"gitlab.com/jsiebens/gke-resizer/pkg/gkeresizer"
	"google.golang.org/api/container/v1"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Disable timestamps in go logs because stackdriver has them already.
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := "0.0.0.0:" + port

	service, err := container.NewService(context.Background())
	if err != nil {
		log.Fatalf("failed to create resizer: %s", err)
	}
	resizer, err := gkeresizer.NewResizer(*service)
	if err != nil {
		log.Fatalf("failed to create resizer: %s", err)
	}

	resizerServer, err := gkeresizer.NewServer(resizer)
	if err != nil {
		log.Fatalf("failed to create server: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/http", resizerServer.HTTPHandler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("server is listening on %s\n", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server exited: %s", err)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh

	log.Printf("received stop, shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server: %s", err)
	}
}
