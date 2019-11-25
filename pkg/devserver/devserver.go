package devserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
)

type Configuration struct {
	ListenAddress         string
	HttpDevConfigurations []HttpDevConfiguration
}

type HttpDevConfiguration struct {
	DestinationAddress string
	MountURLPrefix     string
}

type serverInstance struct {
	configuration      *Configuration
	errorShutdownChan  chan error
	masterShutdownChan chan struct{}
	shutdownWg         *sync.WaitGroup
	lastRequestID      uint64
}

func Start(conf Configuration) {
	serv := serverInstance{
		configuration:      &conf,
		errorShutdownChan:  make(chan error),
		masterShutdownChan: make(chan struct{}),
		shutdownWg:         new(sync.WaitGroup),
	}

	// Start HTTP development server
	serv.shutdownWg.Add(1)
	go serv.startHTTPServer()

	// Monitoring for OS signal --or-- sub-system to fail.
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case sig := <-sigChan:
		log.Printf("Shutting down server (received message %q)...", sig)
	case err := <-serv.errorShutdownChan:
		log.Printf("Shutting down server because of an error: %v", err)
	}

	// Shutting development server down.
	close(serv.masterShutdownChan)
	serv.shutdownWg.Wait()
	log.Print("Server is off.")
}

func (serv *serverInstance) startHTTPServer() {
	defer serv.shutdownWg.Done()

	// Handle every different configuration path
	engine := http.NewServeMux()
	for _, hc := range serv.configuration.HttpDevConfigurations {
		mountURLPrefix := hc.MountURLPrefix
		if mountURLPrefix == "" {
			mountURLPrefix = "/"
		}
		engine.HandleFunc(mountURLPrefix, serv.generateHandler(&hc))
	}
	server := http.Server{
		Addr:    serv.configuration.ListenAddress,
		Handler: engine,
	}

	// Shut the HTTP server down if the dev-server shuts down.
	go func() {
		<-serv.masterShutdownChan
		if err := server.Shutdown(context.Background()); err != nil && err != http.ErrServerClosed {
			log.Printf("error shutting HTTP server down: %v", err)
		}
	}()

	// Start listening for HTTP traffic.
	log.Print("Starting Dev HTTP Server")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		serv.errorShutdownChan <- fmt.Errorf("starting HTTP server: %w", err)
	}
	log.Print("Dev HTTP Server is down")
}

func (serv *serverInstance) generateHandler(hc *HttpDevConfiguration) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		requestID := atomic.AddUint64(&serv.lastRequestID, 1)

		log.Printf("[request %d] request: %v", requestID, request)

		writer.Header().Add("x-devserver-request-id", fmt.Sprintf("%d", requestID))
		writer.WriteHeader(http.StatusOK)
	}
}
