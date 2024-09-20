package external

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"jobs/types"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type ExternalJobsFetcher interface {
	FetchExternalJobs(name string, minSalary, maxSalary int64, country string) ([]types.Job, error)
}

type ExternalJobs struct {
	Client *http.Client
	Log    *zap.Logger
}

func NewExternalJobs(client *http.Client, log *zap.Logger) *ExternalJobs {
	return &ExternalJobs{Client: client, Log: log}
}

func (e *ExternalJobs) FetchExternalJobs(name string, minSalary, maxSalary int64, country string) ([]types.Job, error) {
	apiURL := buildAPIURL(name, minSalary, maxSalary, country)
	e.Log.Sugar().Infof("Fetching jobs from API: %s", apiURL)

	resp, err := e.Client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching jobs from API (%s): %w", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var jobsResponse map[string][][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jobsResponse); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}

	var jobs []types.Job
	jobList, ok := jobsResponse[country]
	if !ok {
		return nil, fmt.Errorf("no jobs found for country: %s", country)
	}

	for _, jobData := range jobList {
		if len(jobData) < 3 {
			continue // Skip if there isn't enough data
		}

		// Parse the title, salary, and skills from jobData
		title, ok1 := jobData[0].(string)
		salary, ok2 := jobData[1].(float64) // JSON numbers are decoded as float64
		skillsXML, ok3 := jobData[2].(string)

		if !ok1 || !ok2 || !ok3 {
			continue // Skip if the type assertion fails
		}

		// Unmarshal the XML into the Skills structure
		var skills types.Skills
		if err := xml.Unmarshal([]byte(skillsXML), &skills); err != nil {
			return nil, fmt.Errorf("could not unmarshal skills XML: %w", err)
		}

		// Append the job to the jobs slice
		jobs = append(jobs, types.Job{
			Title:  title,
			Salary: int(salary),
			Skills: skills,
		})
	}

	return jobs, nil
}

func buildAPIURL(name string, minSalary, maxSalary int64, country string) string {
	params := url.Values{}

	params.Add("name", name)
	if minSalary > 0 {
		params.Add("salary_min", fmt.Sprint(minSalary))
	}
	if maxSalary > 0 {
		params.Add("salary_max", fmt.Sprint(maxSalary))
	}
	if country != "" {
		params.Add("country", country)
	}

	apiURL := fmt.Sprintf("http://localhost:8081/jobs?%s", params.Encode())
	return apiURL
}
