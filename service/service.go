package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	d "jobs/db"
	e "jobs/external"
	"jobs/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service interface for the service methods
type Service interface {
	Subscribe(ctx context.Context, input types.SubscribeInput) (types.SubscribeOutput, error)
	GetJobs(ctx context.Context, input types.JobsInput) (types.JobsOutput, error)
}

// JobsService implements the Service interface
type JobsService struct {
	DB          d.Database
	Logger      *zap.Logger
	JobsFetcher e.ExternalJobsFetcher
}

// NewJobsService creates a new instance of JobsService
func NewJobsService(logger *zap.Logger, conn d.Database, externalJobsFetcher e.ExternalJobsFetcher) *JobsService {
	return &JobsService{Logger: logger, DB: conn, JobsFetcher: externalJobsFetcher}
}

// Subscribe method for JobsService
func (s *JobsService) Subscribe(ctx context.Context, input types.SubscribeInput) (types.SubscribeOutput, error) {
	id, err := s.DB.RecordSubscriber(ctx, &input)
	if err != nil {
		return types.SubscribeOutput{}, fmt.Errorf("could not upsert user to subscriber table: %v", err)
	}

	output := types.SubscribeOutput{
		UserID:    id,
		Name:      input.Name,
		TimeStamp: time.Now(),
		Message:   "User successfully subscribed",
	}
	return output, nil
}
func (s *JobsService) GetJobs(ctx context.Context, input types.JobsInput) (types.JobsOutput, error) {
	var (
		internalJobs []uuid.UUID
		externalJobs []types.Job
		internalErr  error
		externalErr  error
		wg           sync.WaitGroup
		errChan      = make(chan error, 2)
		jobsChan     = make(chan struct {
			internal []uuid.UUID
			external []types.Job
		}, 1)
	)

	s.Logger.Sugar().Info("Starting to fetch internal jobs...")

	// Fetch internal jobs concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		internalJobs, internalErr = s.DB.GetInternalJobs(ctx, &input)
		if internalErr != nil {
			errChan <- fmt.Errorf("could not get internal jobs: %w", internalErr)
		}
		s.Logger.Sugar().Infof("Got internal jobs: %v", len(internalJobs))
	}()

	s.Logger.Sugar().Info("Starting to fetch external jobs...")

	// Fetch external jobs concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		externalJobs, externalErr = s.fetchAllExtJobs(&input)
		if externalErr != nil {
			errChan <- fmt.Errorf("could not get external jobs: %w", externalErr)
		} else {
			s.Logger.Sugar().Infof("Got external jobs: %v", len(externalJobs))
		}
	}()

	// Collect results once all goroutines have completed
	go func() {
		wg.Wait()
		close(errChan)
		jobsChan <- struct {
			internal []uuid.UUID
			external []types.Job
		}{
			internal: internalJobs,
			external: externalJobs,
		}
	}()

	var err error
	// Process any errors encountered
	for e := range errChan {
		if e != nil {
			if err == nil {
				err = e // Set the first error encountered
			}
			s.Logger.Sugar().Errorf("Error fetching jobs: %v", e)
		}
	}

	// Retrieve the job results
	jobs := <-jobsChan

	output := types.JobsOutput{
		InternalJobs: jobs.internal,
		ExternalJobs: jobs.external,
	}

	s.Logger.Sugar().Infof("Got jobs: internal: %v, external: %v", len(output.InternalJobs), len(output.ExternalJobs))
	return output, err
}

func (s *JobsService) fetchAllExtJobs(in *types.JobsInput) ([]types.Job, error) {
	var allJobs []types.Job

	s.Logger.Sugar().Info("Starting to fetch all external jobs...")

	if in.PreferredCountries == nil {
		in.PreferredCountries = []string{"Argentina"}
	}
	for _, title := range in.JobTitles {
		for _, country := range in.PreferredCountries {
			s.Logger.Sugar().Infof("Fetching external jobs for title: %v, country: %v", title, country)
			jobs, err := s.JobsFetcher.FetchExternalJobs(title, in.SalaryMin, 0, country)
			if err != nil {
				return nil, fmt.Errorf("could not fetch external jobs %v/%v: %v", title, country, err)
			}
			allJobs = append(allJobs, jobs...)
		}
	}
	s.Logger.Sugar().Info("Finished fetching all external jobs")
	return allJobs, nil
}
