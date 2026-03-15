-- name: GetAllJobs :many
SELECT *
FROM
    jobs
WHERE
    deleted <> FALSE;

-- name: InsertJob :exec
INSERT INTO jobs (job_hash, page, title)
VALUES($1, $2, $3);

