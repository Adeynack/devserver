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
	engine.Use(gin.Recovery())
	engine.GET("/persons", getPersonList)
	engine.GET("/persons/:personID", getPersonById)

	srv := &http.Server{
		Addr:    "localhost:3000",
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	log.Printf("Shutting down server (received signal %q)", <-quit)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
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
