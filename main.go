package main

import (
	"context"
	"jobs/external"
	"jobs/server"
	"jobs/service"
	s "jobs/setup"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/grafana/pyroscope-go"
	_ "github.com/lib/pq"
)

// main is the entry point of the application.
//
// It sets up the database connection, creates a new notification service, and starts the server.
// No parameters.
// No return values.
func main() {
	ctx := context.Background()

	// Setup the database connection
	db, err := s.Setup(ctx)
	logger := db.Logger
	if err != nil {
		logger.Sugar().Fatalf("could not configure db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Sugar().Fatalf("error closing database: %v", err)
		}
	}()

	_, err = pyroscope.Start(pyroscope.Config{
		ApplicationName: "your.app.name",
		ServerAddress:   "http://pyroscope:4040", // URL del servidor Pyroscope
	})

	if err != nil {
		logger.Sugar().Warnf("could not start Pyroscope: %v", err)
	}
	// Run pprof in a separate goroutine
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			logger.Sugar().Warnf("pprof server error: %v", err)
		}
	}()
	client := &http.Client{Timeout: 10 * time.Second}
	jobsFetcher := external.NewExternalJobs(client, logger)
	jobsService := service.NewJobsService(logger, db, jobsFetcher)
	port := ":8080"
	server.ServerSetup(jobsService, port, logger)
}
