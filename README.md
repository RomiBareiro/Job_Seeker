# Job Subscription API

This project provides an API for subscribing to job notifications and querying job listings based on user preferences.
- **Swagger/OpenAPI Documentation**: Documentation for the API endpoints.

## Requirements

- Docker
- Docker Compose
- Go 1.22
- Postgresql

# Setup

## Create a .env file:

If you need to configure specific environment variables, create a .env file in the root directory with the following variables 
(See .env_example)


# Usage

To bring up your application containers, use the following command:

```bash

docker compose up --build
```

This command builds the Docker images according to the Dockerfile and docker-compose.yml files in the project and starts the necessary containers.

## API Endpoints

## Subscribe

        Method: POST

        Path: /V1/subscribe

        Description: Allows a user to subscribe to job notifications.

### Request Body:

```json

{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "job_titles": ["Full Stack Developer"],
  "country": ["Argentina"],
  "salary_min": 50000
}
```
### Successful Response:

```json

{
  "id": "uuid",
  "name": "John Doe",
  "timestamp": "2024-08-27T12:00:00Z",
  "message": "Subscription created successfully"
}
```

### Errors:
        400 Bad Request
        422 Validation Error
        500 Internal Server Error

## Jobs

    Method: GET

    Path: /V1/jobs

    Description: Retrieves job listings based on user preferences and query parameters.

### Query Parameters:
        id (optional): User ID.
        posted_date (optional): Job posted date.
        job_titles (optional): List of job titles.
        country (optional): List of preferred countries.

### Successful Response:

```json

        {
          "internal_jobs": ["uuid1", "uuid2"],
          "external_jobs": [
            {
              "title": "Full Stack Developer",
              "salary": 58000,
              "skills": {
                "skills": ["JavaScript", "Node.js", "React"]
              }
            }
          ]
        }
```

        Errors:
            400 Bad Request
            422 Validation Error
            500 Internal Server Error

# API Documentation

Check openpi.yml file


### Test the API:

Use tools like curl, Postman, or Swagger UI to send requests to the endpoints and verify the API functionality.

### Logs

You can view the logs of the containers with:

```bash
docker compose logs -f
```
