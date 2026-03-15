ALTER TABLE jobs
DROP COLUMN deleted,
DROP COLUMN deleted_at;

ALTER TABLE jobs
DROP CONSTRAINT jobs_job_hash_deleted_unique;

ALTER TABLE jobs
ADD CONSTRAINT jobs_job_hash_unique
UNIQUE(job_hash)
