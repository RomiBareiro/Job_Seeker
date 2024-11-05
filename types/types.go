package types

import (
	"encoding/xml"
	"time"

	"github.com/google/uuid"
)

type SubscribeInput struct {
	Name               string   `json:"name" validate:"required"`
	Email              string   `json:"email" validate:"required,email"`
	JobTitles          []string `json:"job_titles" validate:"required,min=1,dive,required"`
	PreferredCountries []string `json:"country" validate:"required,min=1,dive,required"`
	SalaryMin          int64    `json:"salary_min" validate:"required,min=0"`
}

type SubscribeOutput struct {
	UserID    uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	TimeStamp time.Time `json:"timestamp,omitempty"`
	Message   string    `json:"message,omitempty"`
}
type JobsInput struct {
	UserID             uuid.UUID `json:"id"`
	JobTitles          []string  `json:"job_titles,omitempty"`
	SalaryMin          int64     `json:"salary_min,omitempty"`
	PostedDate         time.Time `json:"posted_date" validate:"required"`
	PreferredCountries []string  `json:"country,omitempty"`
}

type JobsOutput struct {
	InternalJobs []uuid.UUID `json:"internal_jobs" db:"id"`
	ExternalJobs []Job       `json:"external_jobs"`
	Message      string      `json:"message,omitempty"`
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Skill struct {
	Name string `xml:",chardata"`
}

type Skills struct {
	XMLName xml.Name `xml:"skills"`
	Skills  []Skill  `xml:"skill"`
}

type Job struct {
	Title  string `xml:"title"`
	Salary int    `xml:"salary"`
	Skills Skills `xml:"skills"`
}

type CountryJobs struct {
	Jobs []Job `xml:"job"`
}

type Response struct {
	XMLName   xml.Name               `xml:"root"`
	Countries map[string]CountryJobs `xml:",any"`
}
