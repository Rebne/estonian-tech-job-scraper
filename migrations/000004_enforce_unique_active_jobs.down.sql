DROP INDEX jobs_job_hash_active_unique;

UPDATE jobs
SET deleted_at = COALESCE(deleted_at, created_at)
                 + (id * INTERVAL '1 microsecond')
WHERE deleted = TRUE;

ALTER TABLE jobs
ADD CONSTRAINT jobs_job_hash_deleted_unique
UNIQUE (job_hash, deleted, deleted_at);
