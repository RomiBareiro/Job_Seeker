DO $$ 
DECLARE
    job_titles job_title[] := ARRAY[
        'SSr Java Developer',
        'Sr Java Developer',
        'Frontend Developer',
        'Backend Developer',
        'Full Stack Developer'
    ];
    countries country[] := ARRAY[
        'Argentina',
        'Australia',
        'USA',
        'UK',
        'ALL'
    ];
BEGIN
    FOR i IN 1..50 LOOP
        INSERT INTO jobs (title, description, location, salary_min, country, posted_date, updated_at)
        VALUES (
            job_titles[ceil(random() * array_length(job_titles, 1))], 
            'Job description for ' || i, 
            'Location ' || i, 
            (random() * 50000 + 50000)::INTEGER, 
            countries[ceil(random() * array_length(countries, 1))],
            CURRENT_TIMESTAMP - (random() * 365) * '1 day'::INTERVAL, 
            CURRENT_TIMESTAMP
        );
    END LOOP;
END $$;

