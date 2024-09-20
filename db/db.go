package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"jobs/types"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// Database defines the interface for database operations
type Database interface {
	RecordSubscriber(ctx context.Context, input *types.SubscribeInput) (uuid.UUID, error)
	GetInternalJobs(ctx context.Context, input *types.JobsInput) ([]uuid.UUID, error)
	Close() error
}

type DBConnector struct {
	DB     *sqlx.DB
	Logger *zap.Logger
}

// RecordSubscriber records a new subscriber in the database.
//
// It takes a context and a SubscribeInput struct as parameters.
// The SubscribeInput struct contains the subscriber's name, email, job titles, and salary minimum.
// It returns the ID of the newly recorded subscriber and an error if any.
func (db *DBConnector) RecordSubscriber(ctx context.Context, input *types.SubscribeInput) (uuid.UUID, error) {
	now := time.Now().UTC()

	jobTitles := fmt.Sprintf(`{"%s"}`, strings.Join(input.JobTitles, `","`))
	if len(input.JobTitles) == 0 {
		jobTitles = "{}"
	}
	countries := fmt.Sprintf(`{"%s"}`, strings.Join(input.PreferredCountries, `","`))
	if len(input.PreferredCountries) == 0 {
		countries = "{}"
	}
	const query = `
        INSERT INTO subscribers (user_name, email, job_titles, salary_min, preferred_countries, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (email)
        DO UPDATE SET
            user_name = EXCLUDED.user_name,
            job_titles = EXCLUDED.job_titles,
			salary_min = EXCLUDED.salary_min,
            updated_at = EXCLUDED.updated_at,
			preferred_countries = EXCLUDED.preferred_countries
        RETURNING id;
    `
	var id uuid.UUID
	err := db.DB.QueryRowContext(ctx, query, input.Name, input.Email, jobTitles, input.SalaryMin, countries, now, now).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error upserting subscriber: %w", err)
	}

	return id, nil
}

func (db *DBConnector) GetInternalJobs(ctx context.Context, input *types.JobsInput) ([]uuid.UUID, error) {
	originalJobTitles := input.JobTitles
	originalCountries := input.PreferredCountries

	if err := getUserInfo(ctx, db.DB, input); err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}
	batchSize := 20
	if originalJobTitles != nil {
		input.JobTitles = originalJobTitles
	}
	if originalCountries != nil {
		input.PreferredCountries = originalCountries
	}
	db.Logger.Sugar().Infof("input %v", input)
	i, err := getInternalJobs(ctx, db.DB, input, batchSize)
	if err != nil {
		return nil, fmt.Errorf("error getting internal jobs: %w", err)
	}
	db.Logger.Sugar().Infof("Retrieved %v internal jobs", i)
	return i, nil
}

// getUserInfo retrieves user information from the subscribers table based on the provided input ID.
//
// It takes a context, a database connection, and a JobsInput struct as parameters.
// Returns an error if the user information cannot be retrieved.
func getUserInfo(ctx context.Context, db *sqlx.DB, input *types.JobsInput) error {
	jobTitles := []string{}
	countries := []string{}
	const query = `
		SELECT 
			COALESCE(job_titles, '{}'),
			COALESCE(preferred_countries, '{}')
		FROM 
			subscribers
		WHERE id = $1
	`
	err := db.QueryRowxContext(ctx, query, input.UserID).Scan(pq.Array(&jobTitles), pq.Array(&countries))
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with ID %v not found", input.UserID)
		}
		return fmt.Errorf("error getting user info: %w", err)
	}

	input.JobTitles = jobTitles
	input.PreferredCountries = countries
	return nil
}

func getInternalJobs(ctx context.Context, db *sqlx.DB, input *types.JobsInput, batchSize int) ([]uuid.UUID, error) {
	var allJobIDs []uuid.UUID
	offset := 0
	const query = `
        SELECT
            id
        FROM
            jobs
        WHERE
            salary_min >= $1
            AND title = ANY($3)
			AND country = ANY($4)
		ORDER BY posted_date >= $2
		LIMIT $5 
		OFFSET $6
    `

	for {
		rows, err := db.QueryxContext(ctx, query, input.SalaryMin, input.PostedDate, pq.Array(input.JobTitles), pq.Array(input.PreferredCountries), batchSize, offset)
		if err != nil {
			return nil, fmt.Errorf("error executing query: %w", err)
		}

		var batch []uuid.UUID
		for rows.Next() {
			var jobID uuid.UUID
			if err := rows.Scan(&jobID); err != nil {
				return nil, fmt.Errorf("error scanning row: %w", err)
			}
			batch = append(batch, jobID)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over rows: %w", err)
		}

		if len(batch) == 0 {
			break
		}

		allJobIDs = append(allJobIDs, batch...)
		offset += batchSize
	}

	return allJobIDs, nil
}

func (db *DBConnector) Close() error {
	return db.DB.Close()
}
