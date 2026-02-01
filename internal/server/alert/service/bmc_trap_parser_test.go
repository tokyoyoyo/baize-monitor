package service

import (
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	storage "baize-monitor/pkg/storage/postgres"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestRepo() repository.BMCTrapParserRepository {
	testConf, err := config.LoadTestMockServerConfig()
	if err != nil {
		panic("")
	}
	db, err := storage.NewClient(testConf.PostGresConfig)
	if err != nil {
		panic("")
	}

	db.DB.Migrator().DropTable(&models.BMCTrapParser{})
	db.AutoMigrate(&models.BMCTrapParser{})

	return repository.NewBMCTrapParserRepository(db.DB)
}

func TestBMCTrapParserServiceImp_validateOID(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	tests := []struct {
		name      string
		oid       string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid OID",
			oid:       "1.3.6.1.4.1.2.3",
			fieldName: "test OID",
			wantErr:   false,
		},
		{
			name:      "empty OID",
			oid:       "",
			fieldName: "test OID",
			wantErr:   true,
		},
		{
			name:      "invalid OID format",
			oid:       "1.3.6.invalid",
			fieldName: "test OID",
			wantErr:   true,
		},
		{
			name: "OID too long",
			oid: "1.3.6.1.4.1.2.3.4.5.6.7.8.9.10.11.12.13.14.15.16.17.18.19.20.21.22.23.24.25.26.27.28.29.30.31." +
				"32.33.34.35.36.37.38.39.40.41.42.43.44.45.46.47.48.49.50.51.52.53.54.55.56.57.58.59.60.61.62.63.64.65." +
				"32.33.34.35.36.37.38.39.40.41.42.43.44.45.46.47.48.49.50.51.52.53.54.55.56.57.58.59.60.61.62.63.64.65." +
				"32.33.34.35.36.37.38.39.40.41.42.43.44.45.46.47.48.49.50.51.52.53.54.55.56.57.58.59.60.61.62.63.64.65." +
				"32.33.34.35.36.37.38.39.40.41.42.43.44.45.46.47.48.49.50.51.52.53.54.55.56.57.58.59.60.61.62.63.64.65." +
				"32.33.34.35.36.37.38.39.40.41.42.43.44.45.46.47.48.49.50.51.52.53.54.55.56.57.58.59.60.61.62.63.64.65." +
				"32.33.34.35.36.37.38.39.40.41.42.43.44.45.46.47.48.49.50.51.52.53.54.55.56.57.58.59.60.61.62.63.64.65." +
				"66.67.68.69.70.71.72.73.74.75.76.77.78.79.80.81.82.83.84.85.86.87.88.89.90.91.92.93.94.95.96.97.98.99.100",
			fieldName: "test OID",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateOID(tt.oid, tt.fieldName)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestBMCTrapParserServiceImp_validateMappings(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	t.Run("valid mappings", func(t *testing.T) {
		req := &request.BMCTrapParserCreate{
			VendorName:        "TestVendor",
			AlertLevelOID:     "1.3.6.1.4.1.2.1",
			AlertContentOID:   "1.3.6.1.4.1.2.2",
			AlertTimeOID:      "1.3.6.1.4.1.2.3",
			AlertComponentOID: "1.3.6.1.4.1.2.4",
			TimeFormat:        "2006-01-02 15:04:05",
			LevelMappings: map[string]string{
				"1": "critical",
				"2": "warning",
			},
			ComponentMappings: map[string][]string{
				"CPU": {"cpu1", "cpu2"},
			},
		}

		err := service.validateMappings(req)
		assert.NoError(t, err)
	})

	t.Run("empty level mappings", func(t *testing.T) {
		req := &request.BMCTrapParserCreate{
			VendorName:        "TestVendor",
			AlertLevelOID:     "1.3.6.1.4.1.2.1",
			AlertContentOID:   "1.3.6.1.4.1.2.2",
			AlertTimeOID:      "1.3.6.1.4.1.2.3",
			AlertComponentOID: "1.3.6.1.4.1.2.4",
			TimeFormat:        "2006-01-02 15:04:05",
			LevelMappings:     map[string]string{},
			ComponentMappings: map[string][]string{
				"CPU": {"cpu1", "cpu2"},
			},
		}

		err := service.validateMappings(req)
		assert.Error(t, err)
	})

	t.Run("invalid level mapping value", func(t *testing.T) {
		req := &request.BMCTrapParserCreate{
			VendorName:        "TestVendor",
			AlertLevelOID:     "1.3.6.1.4.1.2.1",
			AlertContentOID:   "1.3.6.1.4.1.2.2",
			AlertTimeOID:      "1.3.6.1.4.1.2.3",
			AlertComponentOID: "1.3.6.1.4.1.2.4",
			TimeFormat:        "2006-01-02 15:04:05",
			LevelMappings: map[string]string{
				"1": "invalid-level",
			},
			ComponentMappings: map[string][]string{
				"CPU": {"cpu1", "cpu2"},
			},
		}

		err := service.validateMappings(req)
		assert.Error(t, err)
	})

	t.Run("enable auto close with valid status mappings", func(t *testing.T) {
		req := &request.BMCTrapParserCreate{
			VendorName:                         "TestVendor",
			AlertLevelOID:                      "1.3.6.1.4.1.2.1",
			AlertContentOID:                    "1.3.6.1.4.1.2.2",
			AlertTimeOID:                       "1.3.6.1.4.1.2.3",
			AlertComponentOID:                  "1.3.6.1.4.1.2.4",
			AlertIndexOID:                      "1.3.6.1.4.1.2.5",
			AlertStatusOID:                     "1.3.6.1.4.1.2.6",
			TimeFormat:                         "2006-01-02 15:04:05",
			EnableAutoClose:                    true,
			LevelMappings:                      map[string]string{"1": "critical"},
			ComponentMappings:                  map[string][]string{"CPU": {"cpu1"}},
			StatusMappings:                     map[string]string{"1": string(models.TrapStatusAsserted), "0": string(models.TrapStatusDeasserted)},
			ContactInterComponentIdentifierOID: "1.3.6.1.4.1.2.7",
		}

		err := service.validateMappings(req)
		if err != nil {
			t.Errorf("validateMappings() error = %v, wantErr false", err)
		}
	})

	t.Run("enable auto close without status mappings", func(t *testing.T) {
		req := &request.BMCTrapParserCreate{
			VendorName:        "TestVendor",
			AlertLevelOID:     "1.3.6.1.4.1.2.1",
			AlertContentOID:   "1.3.6.1.4.1.2.2",
			AlertTimeOID:      "1.3.6.1.4.1.2.3",
			AlertComponentOID: "1.3.6.1.4.1.2.4",
			AlertIndexOID:     "1.3.6.1.4.1.2.5",
			AlertStatusOID:    "1.3.6.1.4.1.2.6",
			TimeFormat:        "2006-01-02 15:04:05",
			EnableAutoClose:   true,
			LevelMappings:     map[string]string{"1": "critical"},
			ComponentMappings: map[string][]string{"CPU": {"cpu1"}},
			StatusMappings:    nil,
		}

		err := service.validateMappings(req)
		if err == nil {
			t.Error("validateMappings() expected error for missing status mappings when auto close enabled")
		}
	})
}

func TestBMCTrapParserServiceImp_Create(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	t.Run("valid create request", func(t *testing.T) {
		req := &request.BMCTrapParserCreate{
			ParserName:        "TestParser",
			VendorName:        "TestVendor",
			AlertLevelOID:     "1.3.6.1.4.1.2.1",
			AlertContentOID:   "1.3.6.1.4.1.2.2",
			AlertTimeOID:      "1.3.6.1.4.1.2.3",
			AlertComponentOID: "1.3.6.1.4.1.2.4",
			TimeFormat:        "2006-01-02 15:04:05",
			LevelMappings: map[string]string{
				"1": "critical",
				"2": "warning",
			},
			ComponentMappings: map[string][]string{
				"CPU": {"cpu1", "cpu2"},
			},
		}

		_, err := service.Create(req)
		assert.NoError(t, err)
	})

	t.Run("invalid OID in create request", func(t *testing.T) {
		req := &request.BMCTrapParserCreate{
			VendorName:        "TestVendor2",
			AlertLevelOID:     "invalid-oid",
			AlertContentOID:   "1.3.6.1.4.1.2.2",
			AlertTimeOID:      "1.3.6.1.4.1.2.3",
			AlertComponentOID: "1.3.6.1.4.1.2.4",
			TimeFormat:        "2006-01-02 15:04:05",
			LevelMappings: map[string]string{
				"1": "critical",
			},
			ComponentMappings: map[string][]string{
				"CPU": {"cpu1"},
			},
		}

		_, err := service.Create(req)
		assert.Error(t, err)
	})
}

func TestBMCTrapParserServiceImp_Update(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	// First create a parser for update test
	req := &request.BMCTrapParserCreate{
		ParserName:        "TestParserForUpdate",
		VendorName:        "TestVendor",
		AlertLevelOID:     "1.3.6.1.4.1.2.1",
		AlertContentOID:   "1.3.6.1.4.1.2.2",
		AlertTimeOID:      "1.3.6.1.4.1.2.3",
		AlertComponentOID: "1.3.6.1.4.1.2.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
			"2": "warning",
		},
		ComponentMappings: map[string][]string{
			"CPU": {"cpu1", "cpu2"},
		},
	}

	// Create parser
	status, err := service.Create(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)

	// Test update successfully
	t.Run("update existing parser", func(t *testing.T) {
		// Get the parser ID just created
		parserName := "TestParserForUpdate"
		parsers, _, err := service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parsers.List))

		parserID := parsers.List[0].ID

		// Prepare update request
		parserNameUpdate := "TestParserForUpdate2"
		updateReq := &request.BMCTrapParserUpdate{
			ID:         parserID,
			ParserName: &parserNameUpdate,
		}

		status, err := service.Update(updateReq)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, status)
	})

	// Test update non-existing parser
	t.Run("update non-existing parser", func(t *testing.T) {
		description := "test"
		updateReq := &request.BMCTrapParserUpdate{
			ID:          999, // Non-existing ID
			Description: &description,
		}

		status, err := service.Update(updateReq)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	// Test update with wrong OID format
	t.Run("update with invalid OID", func(t *testing.T) {
		// Get the parser ID just created
		parserName := "TestParserForUpdate"
		parsers, _, err := service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parsers.List))

		parserID := parsers.List[0].ID

		invalid := "invalid-oid"
		updateReq := &request.BMCTrapParserUpdate{
			ID:            parserID,
			AlertLevelOID: &invalid,
		}

		status, err := service.Update(updateReq)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, status)
	})
}

func TestBMCTrapParserServiceImp_Delete(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	// First create a parser for delete test
	req := &request.BMCTrapParserCreate{
		ParserName:        "TestParserForDelete",
		VendorName:        "TestVendor",
		AlertLevelOID:     "1.3.6.1.4.1.2.1",
		AlertContentOID:   "1.3.6.1.4.1.2.2",
		AlertTimeOID:      "1.3.6.1.4.1.2.3",
		AlertComponentOID: "1.3.6.1.4.1.2.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
			"2": "warning",
		},
		ComponentMappings: map[string][]string{
			"CPU": {"cpu1", "cpu2"},
		},
	}

	// Create parser
	status, err := service.Create(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)

	t.Run("delete existing parser", func(t *testing.T) {
		// Get the parser ID just created
		parserName := "TestParserForDelete"
		parsers, _, err := service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parsers.List))
		parserID := parsers.List[0].ID

		// Delete parser
		status, err := service.Delete(parserID)
		assert.NoError(t, err)
		if status != http.StatusOK {
			t.Errorf("Delete() status = %v, want %v", status, http.StatusOK)
		}

		// Confirm parser is deleted (soft delete)
		parsers, _, err = service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, 0, len(parsers.List))
	})

	t.Run("delete non-existing parser", func(t *testing.T) {
		status, err := service.Delete(999) // Non-existing ID
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, status)

	})
}

func TestBMCTrapParserServiceImp_Activate(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	// First create a parser for activate test
	req := &request.BMCTrapParserCreate{
		ParserName:        "TestParserForActivate",
		VendorName:        "TestVendor",
		AlertLevelOID:     "1.3.6.1.4.1.2.1",
		AlertContentOID:   "1.3.6.1.4.1.2.2",
		AlertTimeOID:      "1.3.6.1.4.1.2.3",
		AlertComponentOID: "1.3.6.1.4.1.2.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
			"2": "warning",
		},
		ComponentMappings: map[string][]string{
			"CPU": {"cpu1", "cpu2"},
		},
	}

	// Create parser
	status, err := service.Create(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)

	t.Run("activate deactivated parser", func(t *testing.T) {
		// Get the parser ID just created
		parserName := "TestParserForActivate"
		parsers, _, err := service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parsers.List))
		assert.Equal(t, true, parsers.List[0].IsActive)
		parserID := parsers.List[0].ID

		// Deactivate parser first
		status, err := service.Deactivate(parserID)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, status)
		parsers, _, err = service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, false, parsers.List[0].IsActive)

		// Activate parser
		status, err = service.Activate(parserID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)

		// Confirm parser is activated
		parsers, _, err = service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parsers.List))
		assert.Equal(t, true, parsers.List[0].IsActive)
	})

	t.Run("activate already active parser", func(t *testing.T) {
		// Create another parser, it is active by default
		parserName := "TestParserForActivate2"
		req2 := &request.BMCTrapParserCreate{
			ParserName:        parserName,
			VendorName:        "TestVendor2",
			AlertLevelOID:     "1.3.6.1.4.1.2.5",
			AlertContentOID:   "1.3.6.1.4.1.2.6",
			AlertTimeOID:      "1.3.6.1.4.1.2.7",
			AlertComponentOID: "1.3.6.1.4.1.2.8",
			TimeFormat:        "2006-01-02 15:04:05",
			LevelMappings: map[string]string{
				"1": "critical",
			},
			ComponentMappings: map[string][]string{
				"CPU": {"cpu1"},
			},
		}

		status, err := service.Create(req2)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, status)

		// Get the parser ID just created
		parsers, _, err := service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parsers.List))
		parserID := parsers.List[0].ID

		// Try to activate an already active parser
		status, err = service.Activate(parserID)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("activate non-existing parser", func(t *testing.T) {
		status, err := service.Activate(999) // Non-existing ID
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, status)
	})
}

func TestBMCTrapParserServiceImp_Deactivate(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	// First create a parser for deactivate test
	req := &request.BMCTrapParserCreate{
		ParserName:        "TestParserForDeactivate",
		VendorName:        "TestVendor",
		AlertLevelOID:     "1.3.6.1.4.1.2.1",
		AlertContentOID:   "1.3.6.1.4.1.2.2",
		AlertTimeOID:      "1.3.6.1.4.1.2.3",
		AlertComponentOID: "1.3.6.1.4.1.2.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
			"2": "warning",
		},
		ComponentMappings: map[string][]string{
			"CPU": {"cpu1", "cpu2"},
		},
	}

	// Create parser
	status, err := service.Create(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)

	t.Run("deactivate active parser", func(t *testing.T) {
		// Get the parser ID just created
		parserName := "TestParserForDeactivate"
		parsers, _, err := service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parsers.List))
		parserID := parsers.List[0].ID

		// Deactivate parser
		status, err := service.Deactivate(parserID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)

		// Confirm parser is deactivated
		parsers, _, err = service.List(&request.BMCTrapParserFilter{
			ParserName: &parserName,
			Page:       1,
			PageSize:   10,
		})
		assert.Nil(t, err)
		assert.Equal(t, false, parsers.List[0].IsActive)
	})

	t.Run("deactivate already inactive parser", func(t *testing.T) {
		// Use the parser ID of the parser we already deactivated
		parserID := int64(1) // We know the first parser is now deactivated

		// Try to deactivate an already deactivated parser
		status, err := service.Deactivate(parserID)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("deactivate non-existing parser", func(t *testing.T) {
		status, err := service.Deactivate(999) // Non-existing ID
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, status)
	})
}

func TestBMCTrapParserServiceImp_List(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	// First create several parsers for list test
	req1 := &request.BMCTrapParserCreate{
		ParserName:        "TestParserList1",
		VendorName:        "TestVendorA",
		AlertLevelOID:     "1.3.6.1.4.1.2.1",
		AlertContentOID:   "1.3.6.1.4.1.2.2",
		AlertTimeOID:      "1.3.6.1.4.1.2.3",
		AlertComponentOID: "1.3.6.1.4.1.2.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
		},
		ComponentMappings: map[string][]string{
			"CPU": {"cpu1"},
		},
	}

	req2 := &request.BMCTrapParserCreate{
		ParserName:        "TestParserList2",
		VendorName:        "TestVendorB",
		AlertLevelOID:     "1.3.6.1.4.1.2.5",
		AlertContentOID:   "1.3.6.1.4.1.2.6",
		AlertTimeOID:      "1.3.6.1.4.1.2.7",
		AlertComponentOID: "1.3.6.1.4.1.2.8",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "warning",
		},
	}

	// Create parsers
	_, err := service.Create(req1)
	assert.Nil(t, err)

	_, err = service.Create(req2)
	assert.Nil(t, err)

	t.Run("list all parsers", func(t *testing.T) {
		filter := &request.BMCTrapParserFilter{
			Page:     1,
			PageSize: 10,
		}

		result, status, err := service.List(filter)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, 2, len(result.List))
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.PageSize)
	})

	t.Run("list with parser name filter", func(t *testing.T) {
		var parserName = "TestParserList1"
		filter := &request.BMCTrapParserFilter{
			Page:       1,
			PageSize:   10,
			ParserName: &parserName,
		}

		result, status, err := service.List(filter)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, 1, len(result.List))

		assert.Equal(t, "TestParserList1", result.List[0].ParserName)
	})

	t.Run("list with vendor name filter", func(t *testing.T) {
		var vendorName = "TestVendorA"
		filter := &request.BMCTrapParserFilter{
			Page:       1,
			PageSize:   10,
			VendorName: &vendorName,
		}

		result, status, err := service.List(filter)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, 1, len(result.List))
		assert.Equal(t, "TestParserList1", result.List[0].ParserName)
		assert.Equal(t, "TestVendorA", result.List[0].VendorName)
		assert.Equal(t, "1.3.6.1.4.1.2.1", result.List[0].AlertLevelOID)
	})

	t.Run("list with activation status filter", func(t *testing.T) {
		var isActive = true
		filter := &request.BMCTrapParserFilter{
			Page:     1,
			PageSize: 10,
			IsActive: &isActive,
		}

		result, status, err := service.List(filter)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, status)

		// Verify the returned results
		for _, parser := range result.List {
			assert.True(t, parser.IsActive)
		}
	})
}

func TestBMCTrapParserServiceImp_CheckParserNameExist(t *testing.T) {
	repo := getTestRepo()
	service := NewBMCTrapParserServiceImp(repo)

	// Test non-existing parser name
	exists, err := service.CheckParserNameExist("NonExistentParser", true, 0)
	assert.Nil(t, err)
	assert.False(t, exists)

	// Create a parser, then test if it exists
	req := &request.BMCTrapParserCreate{
		ParserName:        "TestParserForExistCheck",
		VendorName:        "TestVendor",
		AlertLevelOID:     "1.3.6.1.4.1.2.1",
		AlertContentOID:   "1.3.6.1.4.1.2.2",
		AlertTimeOID:      "1.3.6.1.4.1.2.3",
		AlertComponentOID: "1.3.6.1.4.1.2.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
			"2": "warning",
		},
		ComponentMappings: map[string][]string{
			"CPU": {"cpu1", "cpu2"},
		},
	}

	// Create a parser first
	code, err := service.Create(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, code)

	// Check if the parser name just created exists (should exist in create mode)
	exists, err = service.CheckParserNameExist("TestParserForExistCheck", true, 1)
	assert.Nil(t, err)
	assert.True(t, exists)

	// Test parser name check in update mode
	exists, err = service.CheckParserNameExist("TestParserForExistCheck", false, 1)
	assert.Nil(t, err)
	assert.False(t, exists)
}
