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

	client := &http.Client{Timeout: 10 * time.Second}
	jobsFetcher := external.NewExternalJobs(client, logger)
	jobsService := service.NewJobsService(logger, db, jobsFetcher)
	port := ":8080"
	server.ServerSetup(jobsService, port, logger)
}
