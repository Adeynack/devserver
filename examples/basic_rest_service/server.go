package basic_rest_service

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func StartServer() {
	engine := gin.New()
	srv := &http.Server{
		Addr:    "localhost:3000",
		Handler: engine,
	}

	engine.Use(gin.Recovery())

	engine.POST("/shutdown", shutdown(srv))
	engine.GET("/persons", getPersonList)
	engine.GET("/persons/:personID", getPersonById)

	go shutdownOnInterrupt(srv)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe: %s\n", err)
	}
	log.Print("Server is now terminated")
}

func shutdownOnInterrupt(srv *http.Server) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	log.Printf("Shutting down server (received signal %q)", <-quit)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}

func shutdown(srv *http.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		go func() {
			shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := srv.Shutdown(shutdownContext)
			if err != nil {
				log.Printf("Error shutting down server: %v", err)
			}
		}()
		log.Print("Shutting down server (received POST to /shutdown)")
		c.String(http.StatusAccepted, "Shutdown request registered.")
	}
}

func getPersonList(c *gin.Context) {
	c.JSON(http.StatusOK, people)
}

func getPersonById(c *gin.Context) {
	rawPersonId := c.Param("personID")
	if rawPersonId == "" {
		c.String(http.StatusBadRequest, "Path parameter \"personID\" is required.")
		return
	}

	var person *Person
	personID, err := strconv.ParseInt(rawPersonId, 10, 64)
	if err == nil {
		person = findPersonByID(personID)
	}

	if person == nil {
		c.String(http.StatusNotFound,
			fmt.Sprintf("No person with ID %q was found.", rawPersonId))
		return
	}

	c.JSON(http.StatusOK, person)
}
