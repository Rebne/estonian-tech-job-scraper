-- name: GetAllJobs :many
SELECT * FROM jobs;

-- name: GetAllTitles :many
SELECT title FROM jobs;

-- name: InsertJob :exec
INSERT INTO jobs (job_hash, page, title) VALUES($1, $2, $3);

-- name: DeleteJob :exec
DELETE FROM jobs WHERE job_hash = $1 AND page = $2;

-- name: DeleteJobByID :exec
DELETE FROM jobs WHERE id = $1;
