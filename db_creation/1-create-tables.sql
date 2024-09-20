-- Create extension pgcrypto if it does not already exist
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create job_title type if it does not already exist
DO $$
BEGIN
   IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_title') THEN
      CREATE TYPE job_title AS ENUM (
         'SSr Java Developer',
         'Sr Java Developer',
         'Frontend Developer',
         'Backend Developer',
         'Full Stack Developer',
         'ALL' -- all job titles
      );
   END IF;
END
$$;

-- Create country type if it does not already exist
DO $$
BEGIN
   IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'country') THEN
      CREATE TYPE country AS ENUM (
         'Argentina',
         'Australia',
         'USA',
         'UK',
         'ALL' -- all countries
      );
   END IF;
END
$$;

-- Create subscribers table if it does not already exist
CREATE TABLE IF NOT EXISTS subscribers (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    job_titles job_title[],
    salary_min INTEGER,
    preferred_countries country[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create jobs table if it does not already exist
CREATE TABLE IF NOT EXISTS jobs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    title job_title NOT NULL,
    description TEXT,
    location VARCHAR(255),
    salary_min INTEGER,
    country country NOT NULL,
    posted_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
