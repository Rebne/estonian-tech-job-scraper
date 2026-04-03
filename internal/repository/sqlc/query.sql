-- name: GetAllJobs :many
SELECT *
FROM
    jobs
WHERE
    deleted <> TRUE;

-- name: InsertJob :exec
INSERT INTO jobs (job_hash, page, title)
VALUES($1, $2, $3);

-- name: DeleteJob :exec
UPDATE jobs
SET deleted = TRUE,
    deleted_at = NOW()
WHERE
    job_hash = $1 AND
    deleted = FALSE;
