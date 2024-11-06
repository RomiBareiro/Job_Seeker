package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"jobs/service"
	t "jobs/types"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Server structure
type Server struct {
	Logger   *zap.Logger
	Svc      service.Service // Use the Service interface
	ctx      context.Context
	validate *validator.Validate
	Router   *mux.Router
}

var errorResponse = "could not send response"

func NewServer(ctx context.Context, svc service.Service, logger *zap.Logger) *Server {
	v := validator.New()
	return &Server{
		Svc:      svc,
		Logger:   logger,
		ctx:      ctx,
		validate: v,
	}
}

// ValidateRequestBody validates any input structure
func (s *Server) validateRequestBody(body interface{}) error {
	err := s.validate.Struct(body)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

// SubscribeHandler handles subscription requests
func (s *Server) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	var reqBody t.SubscribeInput
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		sendErrorResponse(w, http.StatusUnprocessableEntity, "Invalid JSON format")
		return
	}

	if err := s.validateRequestBody(reqBody); err != nil {
		sendErrorResponse(w, http.StatusUnprocessableEntity, fmt.Sprintf("Validation error: %s", err))
		return
	}

	resp, err := s.Svc.Subscribe(s.ctx, reqBody)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to marshal response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(respBytes); err != nil {
		s.Logger.Error(errorResponse, zap.Error(err))
	}
}
func (s *Server) JobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	queryParams := r.URL.Query()

	idStr := queryParams.Get("id")
	postedDateStr := queryParams.Get("posted_date")
	var input t.JobsInput
	if idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			sendErrorResponse(w, http.StatusUnprocessableEntity, "Invalid User ID format")
			return
		}
		input.UserID = id
	}

	if postedDateStr != "" {
		postedDate, err := time.Parse(time.RFC3339, postedDateStr)
		if err != nil {
			sendErrorResponse(w, http.StatusUnprocessableEntity, "Invalid posted_date format")
			return
		}
		input.PostedDate = postedDate
	}
	input.JobTitles = queryParams["job_titles"]

	output, err := s.Svc.GetJobs(s.ctx, input)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	respBytes, err := json.Marshal(output)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to marshal response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(respBytes); err != nil {
		s.Logger.Error("Error writing response", zap.Error(err))
	}
}

// ServerSetup sets up the server and routes
func ServerSetup(svc service.Service, port string, logger *zap.Logger) *Server {
	s := NewServer(context.Background(), svc, logger)
	s.Router = mux.NewRouter()

	protectedRoutes := s.Router.PathPrefix("/V1").Subrouter()
	protectedRoutes.HandleFunc("/subscribe", s.SubscribeHandler).Methods("POST")
	protectedRoutes.HandleFunc("/jobs", s.JobsHandler).Methods("GET")

	s.Logger.Sugar().Infof("Listening port %s", port)
	s.Logger.Sugar().Fatal(http.ListenAndServe(port, s.Router))

	return s
}

// SendErrorResponse sends an error response in JSON format
func sendErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorResponse := t.ErrorResponse{
		Code:    code,
		Message: message,
	}
	response, err := json.Marshal(errorResponse)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(response); err != nil {
		fmt.Print(errorResponse)
	}
}
