-- name: GetAllJobs :many
SELECT * FROM jobs;

-- name: InsertJob :exec
INSERT INTO jobs (job_hash, page, title) VALUES($1, $2, $3);
