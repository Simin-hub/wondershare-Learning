package pkg

const (
	TASK_PENDING    = "pending"
	TASK_PROCESSING = "processing"
	TASK_COMPLETED  = "completed"
	TASK_FAILED     = "failed"
	TASK_CANCELLED  = "cancelled"
	TASK_TIMEDOUT   = "timed out"
	TASK_QUEUQ      = "task_queue"
)

type Task struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Process string `json:"process"`
}
