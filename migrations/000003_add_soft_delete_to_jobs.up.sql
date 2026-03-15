ALTER TABLE jobs
ADD COLUMN deleted BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN deleted_at TIMESTAMP DEFAULT NULL;

ALTER TABLE jobs
DROP CONSTRAINT jobs_job_hash_unique;

ALTER TABLE jobs
ADD CONSTRAINT jobs_job_hash_deleted_unique
UNIQUE (job_hash, deleted, deleted_at);
