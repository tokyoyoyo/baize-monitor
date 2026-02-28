package constants

// ExecutionStatus 执行状态类型
type ExecutionStatus string

const (
	StatusPending  ExecutionStatus = "pending"
	StatusRunning  ExecutionStatus = "running"
	StatusSuccess  ExecutionStatus = "success"
	StatusFailed   ExecutionStatus = "failed"
	StatusDisabled ExecutionStatus = "disabled"
)
