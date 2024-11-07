package service

import (
	"context"
	"fmt"
	"jobs/setup"
	"jobs/types"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetInternalJobs(ctx context.Context, input *types.JobsInput) ([]uuid.UUID, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockDB) RecordSubscriber(ctx context.Context, input *types.SubscribeInput) (uuid.UUID, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(uuid.UUID), args.Error(1)
}
func (m *MockDB) Close() error {
	return nil
}

// Mock function
func MockFetchExternalJobs(title string, salaryMin int, minExperience int, country string) ([]types.Job, error) {
	return nil, nil
}

type MockExternalJobsFetcher struct {
	mock.Mock
}

func (m *MockExternalJobsFetcher) FetchExternalJobs(name string, minSalary, maxSalary int64, country string) ([]types.Job, error) {
	args := m.Called(name, minSalary, maxSalary, country)
	return args.Get(0).([]types.Job), args.Error(1)
}

// Test cases for JobsService
func TestGetJobs(t *testing.T) {
	tests := []struct {
		name             string
		internalJobs     []uuid.UUID
		internalJobsErr  error
		externalJobs     []types.Job
		externalJobsErr  error
		expectedOutput   types.JobsOutput
		expectedErrorMsg string
	}{
		{
			name:            "Success - Internal and External jobs fetched successfully",
			internalJobs:    []uuid.UUID{uuid.New()},
			internalJobsErr: nil,
			externalJobs:    []types.Job{{Title: "Backend Developer"}},
			externalJobsErr: nil,
			expectedOutput: types.JobsOutput{
				InternalJobs: []uuid.UUID{uuid.New()},
				ExternalJobs: []types.Job{{Title: "Backend Developer"}},
				Message:      "",
			},
			expectedErrorMsg: "",
		},
		{
			name:            "Success - Only internal jobs available",
			internalJobs:    []uuid.UUID{uuid.New()},
			internalJobsErr: nil,
			externalJobs:    nil,
			externalJobsErr: fmt.Errorf("external service error"),
			expectedOutput: types.JobsOutput{
				InternalJobs: []uuid.UUID{uuid.New()},
				ExternalJobs: nil,
				Message:      "Warning: failed to fetch external jobs",
			},
			expectedErrorMsg: "",
		},
		{
			name:            "Error - Internal jobs fetching fails",
			internalJobs:    nil,
			internalJobsErr: fmt.Errorf("database error"),
			externalJobs:    []types.Job{{Title: "Backend Developer"}},
			externalJobsErr: nil,
			expectedOutput: types.JobsOutput{
				InternalJobs: nil,
				ExternalJobs: []types.Job{{Title: "Backend Developer"}},
				Message:      "",
			},
			expectedErrorMsg: "could not get internal jobs: database error",
		},
		{
			name:            "Error - External jobs fetching fails, internal jobs available",
			internalJobs:    []uuid.UUID{uuid.New()},
			internalJobsErr: nil,
			externalJobs:    nil,
			externalJobsErr: fmt.Errorf("external service error"),
			expectedOutput: types.JobsOutput{
				InternalJobs: []uuid.UUID{uuid.New()},
				ExternalJobs: nil,
				Message:      "Warning: failed to fetch external jobs",
			},
			expectedErrorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := setup.SetupLogger()
			mockDB := new(MockDB)
			mockFetcher := new(MockExternalJobsFetcher)

			service := &JobsService{
				DB:          mockDB,
				Logger:      l,
				JobsFetcher: mockFetcher,
			}

			mockDB.On("GetInternalJobs", mock.Anything, mock.Anything).Return(tt.internalJobs, tt.internalJobsErr)
			mockFetcher.On("FetchExternalJobs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.externalJobs, tt.externalJobsErr)

			// Call the GetJobs method
			output, err := service.GetJobs(context.Background(), types.JobsInput{JobTitles: []string{"Backend Developer"}})

			// Assertions
			if tt.expectedErrorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}

			// Check that the length of internal and external jobs is as expected
			assert.Equal(t, len(tt.expectedOutput.InternalJobs), len(output.InternalJobs))
			assert.Equal(t, len(tt.expectedOutput.ExternalJobs), len(output.ExternalJobs))

			// Verify the message
			assert.Equal(t, tt.expectedOutput.Message, output.Message)

			// Assert that the expectations were met
			mockDB.AssertExpectations(t)
			mockFetcher.AssertExpectations(t)
		})
	}
}
func BenchmarkGetJobs(b *testing.B) {
	l, _ := setup.SetupLogger()

	// Mocks
	mockDB := new(MockDB)
	mockFetcher := new(MockExternalJobsFetcher)
	service := &JobsService{
		DB:          mockDB,
		Logger:      l,
		JobsFetcher: mockFetcher,
	}

	// Inputs
	jobUUID := uuid.New()
	mockInternalJobs := []uuid.UUID{jobUUID}
	mockExternalJobs := []types.Job{{Title: "Backend Developer"}}

	// Simulation
	mockDB.On("GetInternalJobs", mock.Anything, mock.Anything).Return(mockInternalJobs, nil)
	mockFetcher.On("FetchExternalJobs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockExternalJobs, nil)

	// Run benchmark
	for i := 0; i < b.N; i++ {
		startTime := time.Now()
		_, err := service.GetJobs(context.Background(), types.JobsInput{JobTitles: []string{"Backend Developer"}})
		if err != nil {
			b.Fatalf("Error en GetJobs: %v", err)
		}
		duration := time.Since(startTime)
		b.Logf("Execution took %s", duration)
	}

	mockDB.AssertExpectations(b)
	mockFetcher.AssertExpectations(b)
}
