package pkg

import "errors"

var (
	ERROR_NON_EXIS  = errors.New("task does not exist")
	ERROR_CANCELLED = errors.New("task Canceled")
	ERROR_STATUS    = errors.New("task status error")
)
