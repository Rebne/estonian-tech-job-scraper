package errors

import "errors"

var ErrNoJobsFound = errors.New("no jobs found")
var ErrPlaywrightTimeout = errors.New("playwright timeout occurred")
