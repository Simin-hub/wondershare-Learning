package service

type Task struct {
	ID string `json:"id"`
}

type TaskStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
