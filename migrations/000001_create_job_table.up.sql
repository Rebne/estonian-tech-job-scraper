CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    job_hash BYTEA NOT NULL,
    page VARCHAR(32) NOT NULL,
    title VARCHAR(255) NOT NULL,
    CONSTRAINT jobs_job_hash_unique UNIQUE(job_hash)
);
