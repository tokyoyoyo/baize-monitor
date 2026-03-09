package repository

import (
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/dto/response"
	"baize-monitor/pkg/models"

	"gorm.io/gorm"
)

type AlertRepo interface {
	Create(alert *models.Alert) error
	Update(tm *models.Alert) error
	List(filter *request.AlertFilter) (response.AlertListResult, error)
	FindByID(id int64) (*models.Alert, error)
}

type AlertRepoImp struct {
	db *gorm.DB
}

func NewAlertRepoImp(db *gorm.DB) AlertRepo {
	return &AlertRepoImp{db: db}
}

func (r *AlertRepoImp) Create(alert *models.Alert) error {
	db := r.db.Model(alert)
	return db.Create(alert).Error
}

func (r *AlertRepoImp) Update(alert *models.Alert) error {
	db := r.db.Model(alert)
	return db.Updates(alert).Error
}

func (r *AlertRepoImp) List(filter *request.AlertFilter) (response.AlertListResult, error) {
	var alerts []*models.Alert
	var total int64

	db := r.db.Model(&models.Alert{})

	// 1. Dynamically build query conditions (unchanged)
	if filter.ParserID != nil {
		db = db.Where("parser_id = ?", *filter.ParserID)
	}
	if filter.Status != nil {
		db = db.Where("alert_status = ?", *filter.Status)
	}
	if filter.TrapOID != nil {
		db = db.Where("trap_oid = ?", *filter.TrapOID)
	}
	if filter.SourceIP != nil {
		db = db.Where("source_ip = ?", *filter.SourceIP)
	}
	if filter.VendorCode != nil {
		db = db.Where("vendor_code = ?", *filter.VendorCode)
	}
	if filter.VendorName != nil {
		db = db.Where("vendor_name LIKE ?", "%"+*filter.VendorName+"%")
	}
	if filter.AlertLevel != nil {
		db = db.Where("alert_level = ?", *filter.AlertLevel)
	}
	if filter.Component != nil {
		db = db.Where("component = ?", *filter.Component)
	}
	if filter.Content != nil {
		db = db.Where("content LIKE ?", "%"+*filter.Content+"%")
	}

	// Time range handling
	if filter.AlertTimeRangeStart != nil {
		db = db.Where("alert_time >= ?", filter.AlertTimeRangeStart)
	}
	if filter.AlertTimeRangeEnd != nil {
		db = db.Where("alert_time <= ?", filter.AlertTimeRangeEnd)
	}

	// 2. Get total count (this section can be removed if total count display is not needed to improve performance)
	if err := db.Count(&total).Error; err != nil {
		return response.AlertListResult{}, err
	}

	// 3. Pagination and sorting
	page := filter.Page
	pageSize := filter.PageSize
	offset := (page - 1) * pageSize

	// 4. Query data
	if err := db.Limit(pageSize).Offset(offset).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return response.AlertListResult{}, err
	}

	// 5. Construct return result
	result := response.AlertListResult{
		List:     alerts,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	// Return only the result and error, no longer return int total
	return result, nil
}

func (r *AlertRepoImp) FindByID(id int64) (*models.Alert, error) {
	var alert models.Alert
	db := r.db.Model(&models.Alert{}).Where("id = ?", id).First(&alert)
	return &alert, db.Error
}
