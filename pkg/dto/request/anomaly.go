package request

import "time"

type AnomalyResultRequest struct {
	Content []AnomalyResult `json:"content"`
}

type AnomalyResult struct {
	CheckType string `json:"check_type"`
	CheckItem string `json:"check_item"`
	Success   bool   `json:"success"`
	Level     string `json:"level"`
	Message   string `json:"message"`

	ExtraData map[string]interface{} `json:"extra_data"`

	CheckedAt time.Time `json:"checked_at"`
}
