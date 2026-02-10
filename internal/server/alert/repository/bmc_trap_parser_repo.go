package repository

import (
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/dto/response"
	"baize-monitor/pkg/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type BMCTrapParserRepository interface {
	List(filter *request.BMCTrapParserFilter) (response.BMCTrapParserListResult, error)
	FindAll() ([]*models.BMCTrapParser, error)
	FindByID(id int64) (*models.BMCTrapParser, error)
	FindByVendorCode(vendorCode string) ([]*models.BMCTrapParser, error)
	IsParserNameExist(parserName string, createMode bool, id int64) (bool, error)
	Create(parser *models.BMCTrapParser) error
	Update(parser *models.BMCTrapParser) error
	Activate(id int64) error
	Deactivate(id int64) error
	SoftDelete(id int64) error
}

type bmcTrapParserRepoImpl struct {
	db *gorm.DB
}

func NewBMCTrapParserRepository(db *gorm.DB) BMCTrapParserRepository {
	return &bmcTrapParserRepoImpl{db: db}
}

func (r *bmcTrapParserRepoImpl) List(filter *request.BMCTrapParserFilter) (response.BMCTrapParserListResult, error) {
	db := r.db.Model(&models.BMCTrapParser{}).Where("is_deleted = ?", false)

	// ParserName fuzzy matching
	if filter.ParserName != nil && *filter.ParserName != "" {
		db = db.Where("parser_name ILIKE ?", "%"+*filter.ParserName+"%")
	}

	// VendorCode exact matching
	if filter.VendorCode != nil && *filter.VendorCode != "" {
		db = db.Where("vendor_code = ?", *filter.VendorCode)
	}

	// VendorName fuzzy matching
	if filter.VendorName != nil && *filter.VendorName != "" {
		db = db.Where("vendor_name ILIKE ?", "%"+*filter.VendorName+"%")
	}

	// IsActive exact matching
	if filter.IsActive != nil {
		db = db.Where("is_active = ?", *filter.IsActive)
	}

	// Description fuzzy matching
	if filter.Description != nil && *filter.Description != "" {
		db = db.Where("description ILIKE ?", "%"+*filter.Description+"%")
	}

	// Get total record count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return response.BMCTrapParserListResult{}, fmt.Errorf("failed to count records: %w", err)
	}

	// Pagination parameters with default and limit
	page := filter.Page
	pageSize := filter.PageSize
	offset := (page - 1) * pageSize

	// Query data
	var parsers []*models.BMCTrapParser
	err := db.
		Limit(pageSize).
		Offset(offset).
		Find(&parsers).Error
	if err != nil {
		return response.BMCTrapParserListResult{}, fmt.Errorf("failed to fetch records: %w", err)
	}

	respList := make([]*response.BMCTrapParserResponse, len(parsers))
	for i, parser := range parsers {
		respList[i] = new(response.BMCTrapParserResponse).FromParserDao(parser)
	}

	result := response.BMCTrapParserListResult{
		List:     respList,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	return result, nil
}

func (r *bmcTrapParserRepoImpl) FindAll() ([]*models.BMCTrapParser, error) {
	var parsers []*models.BMCTrapParser
	err := r.db.Where("is_deleted = ?", false).Find(&parsers).Error
	return parsers, err
}

func (r *bmcTrapParserRepoImpl) FindByID(id int64) (*models.BMCTrapParser, error) {
	var parser models.BMCTrapParser
	err := r.db.First(&parser, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || parser.IsDeleted {
			return nil, nil
		}
		return nil, err
	}
	return &parser, nil
}

func (r *bmcTrapParserRepoImpl) FindByVendorCode(vendorCode string) ([]*models.BMCTrapParser, error) {
	var parsers []*models.BMCTrapParser
	err := r.db.Where("is_deleted = ?", false).Where("vendor_code = ?", vendorCode).Find(&parsers).Error
	return parsers, err
}

func (r *bmcTrapParserRepoImpl) IsParserNameExist(parserName string, createMode bool, id int64) (bool, error) {
	var parser models.BMCTrapParser
	query := r.db.Where("is_deleted = ?", false).Where("parser_name = ?", parserName)

	if !createMode {
		// Update mode, exclude current record
		query = query.Where("id != ?", id)
	}

	err := query.First(&parser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		// If "record not found", it means no record matches the criteria
		return false, err
	}

	if !createMode && parser.ID == id {
		// If it's update mode and the current record's ID matches the query result, return false
		return false, nil
	}

	return true, nil
}

func (r *bmcTrapParserRepoImpl) Create(parser *models.BMCTrapParser) error {
	return r.db.Create(parser).Error
}

func (r *bmcTrapParserRepoImpl) Update(parser *models.BMCTrapParser) error {
	return r.db.Model(&models.BMCTrapParser{}).Where("id = ?", parser.ID).Updates(parser).Error
}

func (r *bmcTrapParserRepoImpl) Activate(id int64) error {
	return r.db.Model(&models.BMCTrapParser{}).
		Where("id = ?", id).
		Update("is_active", true).Error
}

func (r *bmcTrapParserRepoImpl) Deactivate(id int64) error {
	return r.db.Model(&models.BMCTrapParser{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *bmcTrapParserRepoImpl) SoftDelete(id int64) error {
	return r.db.Model(&models.BMCTrapParser{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
		}).Error
}
