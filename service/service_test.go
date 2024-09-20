package service

import (
	"context"
	"fmt"
	"jobs/setup"
	"jobs/types"
	"testing"

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
			internalJobs:    []uuid.UUID{},
			internalJobsErr: nil,
			externalJobs: []types.Job{
				{
					Title:  "Backend Developer",
					Salary: 34000,
					Skills: types.Skills{
						Skills: []types.Skill{
							{Name: "Java"},
							{Name: "OOP"},
							{Name: "Design Patterns"},
						},
					},
				},
				{
					Title:  "Backend Developer",
					Salary: 44000,
					Skills: types.Skills{
						Skills: []types.Skill{
							{Name: "Java"},
							{Name: "OOP"},
							{Name: "Design Patterns"},
						},
					},
				},
			},
			externalJobsErr: nil,
			expectedOutput: types.JobsOutput{
				InternalJobs: []uuid.UUID{},
				ExternalJobs: []types.Job{
					{
						Title:  "Backend Developer",
						Salary: 34000,
						Skills: types.Skills{
							Skills: []types.Skill{
								{Name: "Java"},
								{Name: "OOP"},
								{Name: "Design Patterns"},
							},
						},
					},
					{
						Title:  "Backend Developer",
						Salary: 44000,
						Skills: types.Skills{
							Skills: []types.Skill{
								{Name: "Java"},
								{Name: "OOP"},
								{Name: "Design Patterns"},
							},
						},
					},
				},
			},
			expectedErrorMsg: "",
		},
		{
			name:             "Error - Internal jobs fetching fails",
			internalJobs:     []uuid.UUID{},
			internalJobsErr:  fmt.Errorf("database error"),
			externalJobs:     []types.Job{},
			externalJobsErr:  nil,
			expectedOutput:   types.JobsOutput{},
			expectedErrorMsg: "database error",
		},
		{
			name:             "Error - External jobs fetching fails",
			internalJobs:     []uuid.UUID{},
			internalJobsErr:  nil,
			externalJobs:     []types.Job{},
			externalJobsErr:  fmt.Errorf("external service error"),
			expectedOutput:   types.JobsOutput{InternalJobs: []uuid.UUID{}},
			expectedErrorMsg: "external service error",
		},
		{
			name:             "Error - Both internal and external jobs fetching fails",
			internalJobs:     []uuid.UUID{},
			internalJobsErr:  fmt.Errorf("database error"),
			externalJobs:     []types.Job{},
			externalJobsErr:  nil,
			expectedOutput:   types.JobsOutput{},
			expectedErrorMsg: "database error",
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
			mockFetcher.On("FetchExternalJobs", "Backend Developer", int64(0), int64(0), "Argentina").Return(tt.externalJobs, tt.externalJobsErr)

			input := types.JobsInput{
				JobTitles:          []string{"Backend Developer"},
				SalaryMin:          0,
				PreferredCountries: []string{"Argentina"},
			}

			output, err := service.GetJobs(context.Background(), input)
			if tt.expectedErrorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}

			mockDB.AssertExpectations(t)
			mockFetcher.AssertExpectations(t)
		})
	}
}
