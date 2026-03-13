CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    job_hash BYTEA NOT NULL,
    page VARCHAR(32) NOT NULL,
    title VARCHAR(255) NOT NULL,
    UNIQUE(job_hash, page)
);
