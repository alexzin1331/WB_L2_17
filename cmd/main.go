package main

import (
	_ "WB_L2_17/docs"
	"WB_L2_17/internal/service"
	"WB_L2_17/internal/storage"
	"WB_L2_17/models"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	configPath = "config.yaml"
)

// @title Calendar Events API
// @version 1.0
// @description HTTP API для управления событиями календаря
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	// set logs
	logFile, err := os.OpenFile("/var/log/calendar/calendar.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	// write logs to file and stdout
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	//init config, service, handler
	cfg := models.MustLoad(configPath)
	calendarService := storage.NewManager()
	handler := service.NewRouter(calendarService)

	log.Printf("Starting server on %s", cfg.Host)
	log.Fatal(http.ListenAndServe(cfg.Host, handler))
}
