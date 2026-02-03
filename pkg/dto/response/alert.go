package response

import "baize-monitor/pkg/models"

type AlertListResult struct {
	List     []*models.Alert `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}
