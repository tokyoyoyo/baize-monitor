package service

import (
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/pkg/constants"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/dto/response"
	"baize-monitor/pkg/models"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type BMCTrapParserService interface {
	Create(*request.BMCTrapParserCreate) (int, error)
	CheckParserNameExist(string, bool, int64) (bool, error)
	Update(*request.BMCTrapParserUpdate) (int, error)
	Delete(id int64) (int, error)
	Activate(id int64) (int, error)
	Deactivate(id int64) (int, error)
	List(*request.BMCTrapParserFilter) (response.BMCTrapParserListResult, int, error)
}

type BMCTrapParserServiceImp struct {
	repo repository.BMCTrapParserRepository
}

func NewBMCTrapParserServiceImp(repo repository.BMCTrapParserRepository) *BMCTrapParserServiceImp {
	return &BMCTrapParserServiceImp{repo: repo}
}

func (s *BMCTrapParserServiceImp) validateOID(oid, fieldName string) error {
	if strings.TrimSpace(oid) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if len(oid) > 500 {
		return fmt.Errorf("%s length cannot exceed 500 characters", fieldName)
	}

	if !constants.OidFormatRegex.MatchString(oid) {
		return fmt.Errorf("%s format is invalid, must be numeric dot-separated format", fieldName)
	}

	return nil
}

// validateRequiredOIDs validates OIDs if they exist. First check if empty to skip, then trim spaces to avoid missing multiple spaces
func (s *BMCTrapParserServiceImp) validateRequiredOIDs(req *request.BMCTrapParserCreate) error {
	// Validate four required OIDs
	requiredOIDs := []struct {
		oid       string
		fieldName string
	}{
		{strings.TrimSpace(req.AlertLevelOID), "Alert Level OID"},
		{strings.TrimSpace(req.AlertContentOID), "Alert Content OID"},
		{strings.TrimSpace(req.AlertTimeOID), "Alert Time OID"},
		{strings.TrimSpace(req.AlertComponentOID), "Alert Component OID"},
	}

	for _, item := range requiredOIDs {
		if item.oid == "" {
			continue
		}
		if err := s.validateOID(item.oid, item.fieldName); err != nil {
			return err
		}
	}

	// Validate optional OIDs (if auto-close is enabled)
	if req.EnableAutoClose {
		requiredOIDs = []struct {
			oid       string
			fieldName string
		}{
			{strings.TrimSpace(req.AlertIndexOID), "Alert Index OID"},
			{strings.TrimSpace(req.AlertStatusOID), "Alert Status OID"},
		}

		for _, item := range requiredOIDs {
			if item.oid == "" {
				continue
			}
			if err := s.validateOID(item.oid, item.fieldName); err != nil {
				return err
			}
		}
	}

	// Validate inter-component correlation OID
	if req.EnableContactInterComponentAlerts {
		if req.ContactInterComponentIdentifierOID != "" {
			if err := s.validateOID(strings.TrimSpace(req.ContactInterComponentIdentifierOID), "Inter-component Correlation OID"); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *BMCTrapParserServiceImp) validateMappings(req *request.BMCTrapParserCreate) error {
	// Validate level mappings
	if len(req.LevelMappings) == 0 {
		return errors.New("alert level mappings cannot be empty")
	}

	for key, value := range req.LevelMappings {
		if strings.TrimSpace(key) == "" {
			return errors.New("level mapping key cannot be empty")
		}
		// Validate if value is a valid alert level
		if !isValidAlertLevel(value) {
			return fmt.Errorf("invalid alert level: %s (key: %s)", value, key)
		}
	}

	// If auto-close is enabled, validate status mappings
	if req.EnableAutoClose {
		if len(req.StatusMappings) != 2 {
			return errors.New("when auto-close is enabled, alert status mappings cannot be empty")
		}

		for key, value := range req.StatusMappings {
			if strings.TrimSpace(key) == "" {
				return errors.New("status mapping key cannot be empty")
			}
			if !isValidAlertStatus(value) {
				return fmt.Errorf("invalid alert status: %s (key: %s)", value, key)
			}
		}
	}

	// Validate component mappings
	if len(req.ComponentMappings) != 0 {
		for component, identifiers := range req.ComponentMappings {
			if !isValidAlertComponent(component) {
				return fmt.Errorf("component '%s' is invalid", component)
			}
			if len(identifiers) == 0 {
				return fmt.Errorf("identifier list for component '%s' cannot be empty", component)
			}
			for _, identifier := range identifiers {
				if strings.TrimSpace(identifier) == "" {
					return fmt.Errorf("identifier for component '%s' cannot be empty", component)
				}
			}
		}
	}
	return nil
}

func (s *BMCTrapParserServiceImp) Create(req *request.BMCTrapParserCreate) (int, error) {
	trimmedParserName := strings.TrimSpace(req.ParserName)
	if trimmedParserName == "" {
		return http.StatusBadRequest, fmt.Errorf("parserName cannot be empty")
	}
	exist, err := s.CheckParserNameExist(trimmedParserName, true, 0)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("parserName already exists")
	}
	if exist {
		return http.StatusBadRequest, fmt.Errorf("parserName already exists")
	}

	if err := s.validateRequiredOIDs(req); err != nil {
		return http.StatusBadRequest, err
	}
	if err := s.validateMappings(req); err != nil {
		return http.StatusBadRequest, err
	}

	parser, err := req.ToParserRecord()
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Save via Repository
	err = s.repo.Create(parser)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateParserName) {
			return http.StatusBadRequest, fmt.Errorf("parserName already exists")
		}
		return http.StatusInternalServerError, fmt.Errorf("failed to save parser: %v", err)
	}
	return http.StatusCreated, nil
}

func (s *BMCTrapParserServiceImp) CheckParserNameExist(vendorName string, cteateModel bool, id int64) (bool, error) {
	return s.repo.IsParserNameExist(vendorName, cteateModel, id)
}

// Update updates a Trap parser
func (s *BMCTrapParserServiceImp) Update(req *request.BMCTrapParserUpdate) (int, error) {
	// Check if record exists
	parser, err := s.repo.FindByID(req.ID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get parser: %v", err)
	}
	if parser == nil {
		return http.StatusBadRequest, errors.New("parser does not exist")
	}

	// Validate request (convert to Create request for validation)
	req.UpdateParserRecord(parser)

	rc := new(request.BMCTrapParserCreate)
	rc.FromBMCTrapParser(parser)

	trimmedParserName := strings.TrimSpace(parser.ParserName)
	if trimmedParserName == "" {
		return http.StatusBadRequest, fmt.Errorf("parserName cannot be empty")
	}
	exist, err := s.CheckParserNameExist(trimmedParserName, true, 0)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("update failed: %v", err)
	}
	if exist {
		return http.StatusBadRequest, fmt.Errorf("parserName already exists")
	}
	if err := s.validateRequiredOIDs(rc); err != nil {
		return http.StatusBadRequest, err
	}
	if err := s.validateMappings(rc); err != nil {
		return http.StatusBadRequest, err
	}

	// Update
	if err := s.repo.Update(parser); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("update failed: %v", err)
	}

	return http.StatusOK, nil
}

// Delete deletes a Trap parser
func (s *BMCTrapParserServiceImp) Delete(id int64) (int, error) {
	// Check if record exists
	parser, err := s.repo.FindByID(id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get parser: %v", err)
	}
	if parser == nil {
		return http.StatusBadRequest, errors.New("parser does not exist")
	}

	// Perform soft delete
	if err := s.repo.SoftDelete(id); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete Trap parser: %v", err)
	}
	return http.StatusOK, nil
}

func (s *BMCTrapParserServiceImp) Activate(id int64) (int, error) {
	// Check if record exists
	parser, err := s.repo.FindByID(id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get parser: %v", err)
	}
	if parser == nil {
		return http.StatusBadRequest, errors.New("parser does not exist")
	}
	if parser.IsActive {
		return http.StatusBadRequest, errors.New("parser is already active")
	}
	if err := s.repo.Activate(id); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to activate Trap parser: %v", err)
	}
	return http.StatusOK, nil
}

func (s *BMCTrapParserServiceImp) Deactivate(id int64) (int, error) {
	// Check if record exists
	parser, err := s.repo.FindByID(id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get parser: %v", err)
	}
	if parser == nil {
		return http.StatusBadRequest, errors.New("parser does not exist")
	}
	if !parser.IsActive {
		return http.StatusBadRequest, errors.New("parser is already inactive")
	}
	if err := s.repo.Deactivate(id); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to deactivate Trap parser: %v", err)
	}
	return http.StatusOK, nil
}

func (s *BMCTrapParserServiceImp) List(filter *request.BMCTrapParserFilter) (response.BMCTrapParserListResult, int, error) {
	var resp response.BMCTrapParserListResult
	resp, err := s.repo.List(filter)
	if err != nil {
		return response.BMCTrapParserListResult{}, http.StatusInternalServerError, fmt.Errorf("failed to get list: %v", err)
	}

	return resp, http.StatusOK, nil
}

// ============ Helper functions ============

// isValidAlertLevel validates if alert level is valid
func isValidAlertLevel(level string) bool {
	switch models.AlertLevel(level) {
	case models.AlertLevelCritical,
		models.AlertLevelWarning,
		models.AlertLevelInfo,
		models.AlertLevelNotification:
		return true
	default:
		return false
	}
}

// isValidAlertStatus validates if alert status is valid
func isValidAlertStatus(status string) bool {
	switch models.TrapStatus(status) {
	case models.TrapStatusAsserted,
		models.TrapStatusDeasserted:
		return true
	default:
		return false
	}
}

var componentSet = getComponentSet()

func getComponentSet() map[models.AlertComponent]struct{} {
	cs := make(map[models.AlertComponent]struct{})
	for _, bmd := range models.BMCAlertComponentDescriptions {
		cs[bmd.Component] = struct{}{}
		for _, sub := range bmd.SubComponentsDescription {
			cs[sub.Component] = struct{}{}
		}
	}
	return cs
}

func isValidAlertComponent(component string) bool {
	c := models.AlertComponent(component)
	_, exists := componentSet[c]
	return exists
}
