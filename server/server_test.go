package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"jobs/setup"
	types "jobs/types"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJobsService struct {
	mock.Mock
}

func (m *MockJobsService) Subscribe(ctx context.Context, input types.SubscribeInput) (types.SubscribeOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(types.SubscribeOutput), args.Error(1)
}

func (m *MockJobsService) GetJobs(ctx context.Context, input types.JobsInput) (types.JobsOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(types.JobsOutput), args.Error(1)
}
func TestSubscribeHandler(t *testing.T) {
	logger, _ := setup.SetupLogger()

	// Initialize mock JobsService
	svc := new(MockJobsService)

	// Create a new server instance
	server := NewServer(context.Background(), svc, logger)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		expectedBody   string
		setupMock      func()
	}{
		{
			name:   "Valid Request",
			method: http.MethodPost,
			body: map[string]interface{}{
				"name":       "Romina Bareiro",
				"email":      "bareiro.romina@gmail.com",
				"job_titles": []string{"SSr Java Developer", "Sr Java Developer"},
				"country":    []string{"USA", "Canada"}, // Key should match the JSON tag
				"salary_min": 10000,
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":"00000000-0000-0000-0000-000000000000", "name":"Romina Bareiro", "timestamp":"2023-11-23T16:42:23Z"}`,
			setupMock: func() {
				fixedTime, _ := time.Parse(time.RFC3339, "2023-11-23T16:42:23Z")
				validInput := types.SubscribeInput{
					Name:               "Romina Bareiro",
					Email:              "bareiro.romina@gmail.com",
					JobTitles:          []string{"SSr Java Developer", "Sr Java Developer"},
					PreferredCountries: []string{"USA", "Canada"},
					SalaryMin:          10000,
				}
				validOutput := types.SubscribeOutput{
					UserID:    uuid.Nil,
					Name:      "Romina Bareiro",
					TimeStamp: fixedTime,
				}
				svc.On("Subscribe", mock.Anything, validInput).Return(validOutput, nil)
			},
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `{"code":405, "message":"Method Not Allowed"}`,
			setupMock:      func() {}, // No mock setup needed for this case
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           make(chan int), // Invalid JSON
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"code":422, "message":"Invalid JSON format"}`,
			setupMock:      func() {}, // No mock setup needed for this case
		},
		{
			name:   "Invalid email",
			method: http.MethodPost,
			body: map[string]interface{}{
				"name":       "Romina Bareiro",
				"email":      "bareiro.rominagmail.com",
				"job_titles": []string{"SSr Java Developer", "Sr Java Developer"},
				"country":    []string{"USA", "Canada"},
				"salary_min": 10000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"code":422, "message":"Validation error: validation failed: Key: 'SubscribeInput.Email' Error:Field validation for 'Email' failed on the 'email' tag"}`,
			setupMock: func() {
				svc.On("Subscribe", mock.Anything, mock.AnythingOfType("types.SubscribeInput")).Return(types.SubscribeOutput{}, errors.New("Validation error: Subscription failed"))
			},
		},
		{
			name:   "Invalid preferred country",
			method: http.MethodPost,
			body: map[string]interface{}{
				"name":       "Romina Bareiro",
				"email":      "bareiro.romina@gmail.com",
				"job_titles": []string{"SSr Java Developer", "Sr Java Developer"},
				"salary_min": 10000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"code":422, "message":"Validation error: validation failed: Key: 'SubscribeInput.PreferredCountries' Error:Field validation for 'PreferredCountries' failed on the 'required' tag"}`,
			setupMock: func() {
				svc.On("Subscribe", mock.Anything, mock.AnythingOfType("types.SubscribeInput")).Return(types.SubscribeOutput{}, errors.New("Validation error: Subscription failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock according to test case
			tt.setupMock()

			var req *http.Request
			if tt.body != nil {
				body, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, "/subscribe", bytes.NewReader(body))
			} else {
				req = httptest.NewRequest(tt.method, "/subscribe", nil)
			}

			w := httptest.NewRecorder()

			server.SubscribeHandler(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			responseBody := w.Body.String()
			assert.JSONEq(t, tt.expectedBody, responseBody)

			// Verify mock was called with expected parameters
			svc.AssertExpectations(t)
		})
	}
}

func TestJobsHandler(t *testing.T) {
	logger, _ := setup.SetupLogger()

	// Initialize mock JobsService
	svc := new(MockJobsService)
	server := NewServer(context.Background(), svc, logger)

	tests := []struct {
		name           string
		method         string
		queryParams    map[string]string
		expectedStatus int
		expectedBody   string
		setupMock      func()
	}{
		{
			name:   "Valid Request",
			method: http.MethodGet,
			queryParams: map[string]string{
				"id":          "b2b20e8a-8702-4a44-9ede-3dc9a53e5aa6",
				"posted_date": "2023-11-23T16:42:23Z",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"internal_jobs":["00000000-0000-0000-0000-000000000001","00000000-0000-0000-0000-000000000002"],"external_jobs":[]}`,
			setupMock: func() {
				validInput := types.JobsInput{
					UserID:     uuid.MustParse("b2b20e8a-8702-4a44-9ede-3dc9a53e5aa6"),
					PostedDate: time.Date(2023, time.November, 23, 16, 42, 23, 0, time.UTC),
				}
				validOutput := types.JobsOutput{
					InternalJobs: []uuid.UUID{
						uuid.MustParse("00000000-0000-0000-0000-000000000001"),
						uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					},
					ExternalJobs: []types.Job{},
				}
				svc.On("GetJobs", mock.Anything, validInput).Return(validOutput, nil)
			},
		},
		{
			name:           "Invalid Method",
			method:         http.MethodPost,
			queryParams:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `{"code":405, "message":"Method Not Allowed"}`,
			setupMock:      func() {}, // No mock setup needed for this case
		},
		{
			name:   "Invalid Query Parameters",
			method: http.MethodGet,
			queryParams: map[string]string{
				"id":          "invalid-uuid",
				"posted_date": "invalid-date-format",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"code":422, "message":"Invalid User ID format"}`,
			setupMock: func() {
				// No mocking required, as the handler should handle validation errors
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock according to test case
			tt.setupMock()

			req := httptest.NewRequest(tt.method, "/V1/jobs", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			server.JobsHandler(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			responseBody := w.Body.String()
			assert.JSONEq(t, tt.expectedBody, responseBody)

			// Verify mock was called with expected parameters
			svc.AssertExpectations(t)
		})
	}
}

type MockExternalJobAPI struct{}

func (m *MockExternalJobAPI) FetchJobs(ctx context.Context, input types.JobsInput) (*types.JobsOutput, error) {
	return &types.JobsOutput{
		InternalJobs: []uuid.UUID{uuid.New()},
		ExternalJobs: []types.Job{
			{Title: "Software Engineer", Salary: 120000},
		},
	}, nil
}
func TestIntegration_UserSubscriptionAndJobSearch(t *testing.T) {
	// Setup the mock service
	logger, _ := setup.SetupLogger()
	jobOutput := types.JobsOutput{
		InternalJobs: []uuid.UUID{uuid.New()},
		ExternalJobs: []types.Job{
			{Title: "Software Engineer", Salary: 100000},
		},
	}
	subscribeOutput := types.SubscribeOutput{
		UserID: uuid.New(),
		Name:   "John Doe",
	}
	validInput := types.JobsInput{
		UserID:     uuid.MustParse("b2b20e8a-8702-4a44-9ede-3dc9a53e5aa6"),
		PostedDate: time.Date(2023, time.November, 23, 16, 42, 23, 0, time.UTC),
	}
	mockService := new(MockJobsService)
	mockService.On("Subscribe", mock.AnythingOfType("types.SubscribeInput")).Return(subscribeOutput, nil)
	mockService.On("GetJobs", mock.Anything, validInput).Return(jobOutput, nil)

	// Create the server without actually starting the HTTP server
	server := NewServer(context.Background(), mockService, logger) // Mock the service
	server.Router = mux.NewRouter()                                // Assign router manually
	protectedRoutes := server.Router.PathPrefix("/V1").Subrouter()
	protectedRoutes.HandleFunc("/subscribe", server.SubscribeHandler).Methods("POST")
	protectedRoutes.HandleFunc("/jobs", server.JobsHandler).Methods("GET")

	// Create a mock request and response recorder for testing JobsHandler
	req, err := http.NewRequest("GET", "/V1/jobs?id=b2b20e8a-8702-4a44-9ede-3dc9a53e5aa6&posted_date=2023-11-23T16:42:23Z", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// Call the handler directly
	server.JobsHandler(rr, req)

	// Check for the expected result
	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, rr.Code)
	}

	var response types.JobsOutput
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal("Error decoding response:", err)
	}

	// Validate the response
	if response.Message != "" {
		t.Error("Expected some jobs in response")
	}
}
