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
		wg           sync.WaitGroup
		errChan      = make(chan error, 2)
	)

	s.Logger.Sugar().Info("Starting to fetch jobs...")

	// Fetch internal and external jobs concurrently
	wg.Add(2)
	go s.fetchInternalJobs(ctx, &input, &internalJobs, &wg, errChan)
	go s.fetchExternalJobs(&input, &externalJobs, &wg, errChan)

	// Wait for goroutines to finish and then close errChan
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Process errors
	var err error
	for e := range errChan {
		if e != nil {
			if err == nil {
				err = e
			}
			s.Logger.Sugar().Errorf("Error fetching jobs: %v", e)
		}
	}

	// Set message if external jobs failed
	message := ""
	if err != nil && len(externalJobs) == 0 && len(internalJobs) > 0 {
		message = "Warning: failed to fetch external jobs"
		err = nil // clear error since we have internal jobs
	}

	output := types.JobsOutput{
		InternalJobs: internalJobs,
		ExternalJobs: externalJobs,
		Message:      message,
	}

	s.Logger.Sugar().Infof("Fetched jobs: internal: %v, external: %v, error: %v", len(output.InternalJobs), len(output.ExternalJobs), err)
	return output, err
}

// fetchInternalJobs retrieves internal jobs and sends any error to errChan
func (s *JobsService) fetchInternalJobs(ctx context.Context, input *types.JobsInput, jobs *[]uuid.UUID, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()
	internalJobs, err := s.DB.GetInternalJobs(ctx, input)
	if err != nil {
		s.Logger.Sugar().Errorf("Could not get internal jobs: %v", err)
		errChan <- fmt.Errorf("could not get internal jobs: %w", err)
		return
	}
	*jobs = internalJobs
	s.Logger.Sugar().Infof("Fetched internal jobs: %v", len(internalJobs))
}

// fetchExternalJobs retrieves external jobs and sends any error to errChan
func (s *JobsService) fetchExternalJobs(input *types.JobsInput, jobs *[]types.Job, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()
	externalJobs, err := s.fetchAllExtJobs(input)
	if err != nil {
		s.Logger.Sugar().Errorf("Could not fetch external jobs: %v", err)
		errChan <- fmt.Errorf("could not get external jobs: %w", err)
		return
	}
	*jobs = externalJobs
	s.Logger.Sugar().Infof("Fetched external jobs: %v", len(externalJobs))
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
