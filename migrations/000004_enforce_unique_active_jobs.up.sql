ALTER TABLE jobs
DROP CONSTRAINT jobs_job_hash_deleted_unique;

WITH ranked_active_jobs AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY job_hash
            ORDER BY created_at ASC, id ASC
        ) AS active_rank
    FROM jobs
    WHERE deleted = FALSE
)
UPDATE jobs
SET
    deleted = TRUE,
    deleted_at = transaction_timestamp() + (jobs.id * INTERVAL '1 microsecond')
FROM ranked_active_jobs
WHERE jobs.id = ranked_active_jobs.id
  AND ranked_active_jobs.active_rank > 1;

CREATE UNIQUE INDEX jobs_job_hash_active_unique
ON jobs (job_hash)
WHERE deleted = FALSE;
