package external

import (
	"io"

	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockTransport struct {
	Response   string
	StatusCode int
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.StatusCode,
		Body:       io.NopCloser(strings.NewReader(m.Response)),
		Header:     make(http.Header),
	}, nil
}

func TestFetchExternalJobs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockResponse := `{
			"USA": [
				[
					"Cloud Engineer",
					65000,
					"<skills><skill>AWS</skill><skill>Azure</skill><skill>Docker</skill></skills>"
				]
			],
			"Spain": [
				[
					"Machine Learning Engineer",
					75000,
					"<skills><skill>Python</skill><skill>TensorFlow</skill><skill>Deep Learning</skill></skills>"
				]
			]
		}`

		mockClient := &http.Client{
			Transport: &mockTransport{
				Response:   mockResponse,
				StatusCode: http.StatusOK,
			},
		}

		logger, _ := zap.NewProduction()
		externalJobs := NewExternalJobs(mockClient, logger)

		jobs, err := externalJobs.FetchExternalJobs("Cloud Engineer", 50000, 70000, "USA")
		assert.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, "Cloud Engineer", jobs[0].Title)
		assert.Equal(t, 65000, jobs[0].Salary)
		assert.Len(t, jobs[0].Skills.Skills, 3)
		assert.Contains(t, []string{"AWS", "Azure", "Docker"}, jobs[0].Skills.Skills[0].Name)
	})

	t.Run("Failure", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: &mockTransport{
				Response:   "",
				StatusCode: http.StatusInternalServerError,
			},
		}

		logger, _ := zap.NewProduction()
		externalJobs := NewExternalJobs(mockClient, logger)

		jobs, err := externalJobs.FetchExternalJobs("Cloud Engineer", 50000, 70000, "USA")
		assert.Error(t, err)
		assert.Nil(t, jobs)
		assert.Equal(t, "unexpected status code: 500", err.Error())
	})
}
