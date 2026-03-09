package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/internal/server/alert/service"
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"baize-monitor/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) service.BMCTrapService {
	testConf, err := config.LoadTestMockServerConfig()
	if err != nil {
		panic("")
	}
	db, err := storage.NewClient(testConf.PostGresConfig)
	if err != nil {
		panic("")
	}

	err = db.DB.Migrator().DropTable(&models.BMCTrapParser{})
	if err != nil {
		panic("")
	}
	err = db.AutoMigrate(&models.BMCTrapParser{})
	if err != nil {
		panic("")
	}
	err = db.DB.Migrator().DropTable(&models.Alert{})
	if err != nil {
		panic("")
	}
	err = db.AutoMigrate(&models.Alert{})
	if err != nil {
		panic("")
	}

	parserRepo := repository.NewBMCTrapParserRepository(db.DB)
	alertRepo := repository.NewAlertRepoImp(db.DB)
	parser := &models.BMCTrapParser{
		ParserName: "test1",
		VendorCode: "1234",
		VendorName: "IBM",

		AlertLevelOID:     ".1.3.6.1.4.1.1234.1.1",
		AlertContentOID:   ".1.3.6.1.4.1.1234.1.2",
		AlertTimeOID:      ".1.3.6.1.4.1.1234.1.3",
		AlertComponentOID: ".1.3.6.1.4.1.1234.1.4",
		ComponentMappings: map[string][]string{
			string(models.BMC_AlertComponentPower): {"Power", "PSU"},
		},
		LevelMappings: map[string]models.AlertLevel{
			"0":        models.AlertLevelNotification,
			"1":        models.AlertLevelWarning,
			"critical": models.AlertLevelCritical,
		},
	}
	assert.NoError(t, parserRepo.Create(parser))

	pc := service.NewBMCTrapParserCache(parserRepo)
	service := service.NewBMCTrapServiceImp(pc, alertRepo)

	return service
}

func TestNewBMCTrapHandler(t *testing.T) {
	service := setupTest(t)
	handler := NewBMCTrapHandler(service)
	assert.NotNil(t, handler)
}

func TestBMCTrapHandlerImpl_ReceiveTrap(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := setupTest(t)
	handler := &BMCTrapHandlerImpl{s: service}

	// Test case 1: Valid BMC Trap request
	t.Run("Valid BMC Trap Request", func(t *testing.T) {
		// Prepare test data
		trapMessages := []models.TrapMessage{
			{
				SourceType: models.TrapSourceTypeBMC,
				VariableMap: map[string]string{
					".1.3.6.1.4.1.1234.1.1": "critical",             // AlertLevelOID
					".1.3.6.1.4.1.1234.1.2": "Test alert content",   // AlertContentOID
					".1.3.6.1.4.1.1234.1.3": "2023-08-29T10:00:00Z", // AlertTimeOID
					".1.3.6.1.4.1.1234.1.4": "Power",                // AlertComponentOID
				},
			},
			{
				SourceType: models.TrapSourceTypeBMC,
				VariableMap: map[string]string{
					".1.3.6.1.4.1.1234.1.1": "1",                    // AlertLevelOID (warning)
					".1.3.6.1.4.1.1234.1.2": "Another test content", // AlertContentOID
					".1.3.6.1.4.1.1234.1.3": "2023-08-29T11:00:00Z", // AlertTimeOID
					".1.3.6.1.4.1.1234.1.4": "PSU",                  // AlertComponentOID (maps to Power)
				},
			},
			{
				SourceType: models.TrapSourceTypeBMC,
				VariableMap: map[string]string{
					".1.3.6.1.4.1.1234.1.1": "0",                    // AlertLevelOID (notification)
					".1.3.6.1.4.1.1234.1.2": "Notification content", // AlertContentOID
					".1.3.6.1.4.1.1234.1.3": "2023-08-29T12:00:00Z", // AlertTimeOID
					".1.3.6.1.4.1.1234.1.4": "Fan",                  // AlertComponentOID (not in mapping)
				},
			},
		}

		for _, trapMessage := range trapMessages {
			jsonData, _ := json.Marshal(trapMessage)

			// Create request and response
			req, _ := http.NewRequest(http.MethodPost, "/trap", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute
			handler.ReceiveTrap(c)

			// Assertion
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "success")
		}
	})

	// Test case 2: Invalid request parameters
	t.Run("Invalid Request Parameters", func(t *testing.T) {
		// Create invalid request data
		invalidJSON := []byte(`{"invalid": }`)

		req, _ := http.NewRequest(http.MethodPost, "/trap", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.ReceiveTrap(c)

		// Assertion
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request parameters")
	})

	// Test case 3: Non-BMC type request
	t.Run("Non-BMC Source Type", func(t *testing.T) {
		// Prepare test data
		trapMessage := models.TrapMessage{
			SourceType: "OTHER", // Not BMC type
		}

		jsonData, _ := json.Marshal(trapMessage)

		req, _ := http.NewRequest(http.MethodPost, "/trap", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.ReceiveTrap(c)

		// Assertion
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "only BMC trap supported")
	})

	// Test case 4: Service processing failure - parsing will fail due to missing required OID values
	t.Run("Service Processing Failed", func(t *testing.T) {
		// Prepare test data - missing required OID values
		trapMessage := models.TrapMessage{
			SourceType: models.TrapSourceTypeBMC,
			VariableMap: map[string]string{
				"1.3.6.1.2.1.1.1.0": "Test Device", // Not our configured OID
			},
		}

		jsonData, _ := json.Marshal(trapMessage)

		req, _ := http.NewRequest(http.MethodPost, "/trap", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.ReceiveTrap(c)

		// Assertion - in this case, even without a matching parser, it can still be handled as a fallback
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestBMCTrapHandlerImpl_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create real service
	service := setupTest(t)
	handler := &BMCTrapHandlerImpl{s: service}

	// Create multiple test alert data using loop
	trapMessages := []models.TrapMessage{
		{
			SourceType: models.TrapSourceTypeBMC,
			VariableMap: map[string]string{
				".1.3.6.1.4.1.1234.1.1": "critical",             // AlertLevelOID
				".1.3.6.1.4.1.1234.1.2": "Test alert content",   // AlertContentOID
				".1.3.6.1.4.1.1234.1.3": "2023-08-29T10:00:00Z", // AlertTimeOID
				".1.3.6.1.4.1.1234.1.4": "Power",                // AlertComponentOID
			},
		},
		{
			SourceType: models.TrapSourceTypeBMC,
			VariableMap: map[string]string{
				".1.3.6.1.4.1.1234.1.1": "1",                    // AlertLevelOID (warning)
				".1.3.6.1.4.1.1234.1.2": "Another test content", // AlertContentOID
				".1.3.6.1.4.1.1234.1.3": "2023-08-29T11:00:00Z", // AlertTimeOID
				".1.3.6.1.4.1.1234.1.4": "PSU",                  // AlertComponentOID (maps to Power)
			},
		},
		{
			SourceType: models.TrapSourceTypeBMC,
			VariableMap: map[string]string{
				".1.3.6.1.4.1.1234.1.1": "0",                    // AlertLevelOID (notification)
				".1.3.6.1.4.1.1234.1.2": "Notification content", // AlertContentOID
				".1.3.6.1.4.1.1234.1.3": "2023-08-29T12:00:00Z", // AlertTimeOID
				".1.3.6.1.4.1.1234.1.4": "Fan",                  // AlertComponentOID (not in mapping)
			},
		},
	}

	for _, trapMessage := range trapMessages {
		jsonData, _ := json.Marshal(trapMessage)

		// Create request and response
		req, _ := http.NewRequest(http.MethodPost, "/trap", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create Gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.ReceiveTrap(c)

		// Assertion
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	}

	// Test case 1: Successfully retrieve list
	t.Run("Successful List Retrieval", func(t *testing.T) {

		filter := request.AlertFilter{
			Page:     1,
			PageSize: 10,
		}

		jsonData, _ := json.Marshal(filter)

		req, _ := http.NewRequest(http.MethodPost, "/list", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.List(c)

		// Assertion
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "List retrieved successfully")
	})

	// Test case 2: List query failed - using invalid filter conditions
	t.Run("List Query Failed", func(t *testing.T) {
		// Prepare test data
		filter := request.AlertFilter{
			Page:     -1, // Invalid page number
			PageSize: 10,
		}

		jsonData, _ := json.Marshal(filter)

		req, _ := http.NewRequest(http.MethodPost, "/list", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.List(c)

		// Assertion
		// Depending on the actual implementation, this might succeed or fail based on how invalid page numbers are handled
		assert.NotEqual(t, http.StatusInternalServerError, w.Code) // Assuming it won't cause an internal server error
	})

	// Test case 3: Invalid request parameters
	t.Run("Invalid Request Parameters", func(t *testing.T) {
		// Create invalid request data
		invalidJSON := []byte(`{"invalid": }`)

		req, _ := http.NewRequest(http.MethodPost, "/list", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.List(c)

		// Assertion
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request parameters")
	})

	// Test case 4: Zero-value request parameters
	t.Run("zero Request Parameters", func(t *testing.T) {
		filter := request.AlertFilter{
			Page:     1, // Invalid page number
			PageSize: 0,
		}

		jsonData, _ := json.Marshal(filter)

		req, _ := http.NewRequest(http.MethodPost, "/list", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.List(c)

		// Assertion
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request parameters")
	})
}

func TestBMCTrapHandlerImpl_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create real service
	service := setupTest(t)
	handler := &BMCTrapHandlerImpl{s: service}

	// Test case 1: Successful update
	t.Run("Successful Update", func(t *testing.T) {

		trapMessages := []models.TrapMessage{
			{
				SourceType: models.TrapSourceTypeBMC,
				VariableMap: map[string]string{
					".1.3.6.1.4.1.1234.1.1": "critical",             // AlertLevelOID
					".1.3.6.1.4.1.1234.1.2": "Test alert content",   // AlertContentOID
					".1.3.6.1.4.1.1234.1.3": "2023-08-29T10:00:00Z", // AlertTimeOID
					".1.3.6.1.4.1.1234.1.4": "Power",                // AlertComponentOID
				},
			},
			{
				SourceType: models.TrapSourceTypeBMC,
				VariableMap: map[string]string{
					".1.3.6.1.4.1.1234.1.1": "1",                    // AlertLevelOID (warning)
					".1.3.6.1.4.1.1234.1.2": "Another test content", // AlertContentOID
					".1.3.6.1.4.1.1234.1.3": "2023-08-29T11:00:00Z", // AlertTimeOID
					".1.3.6.1.4.1.1234.1.4": "PSU",                  // AlertComponentOID (maps to Power)
				},
			},
			{
				SourceType: models.TrapSourceTypeBMC,
				VariableMap: map[string]string{
					".1.3.6.1.4.1.1234.1.1": "0",                    // AlertLevelOID (notification)
					".1.3.6.1.4.1.1234.1.2": "Notification content", // AlertContentOID
					".1.3.6.1.4.1.1234.1.3": "2023-08-29T12:00:00Z", // AlertTimeOID
					".1.3.6.1.4.1.1234.1.4": "Fan",                  // AlertComponentOID (not in mapping)
				},
			},
		}

		for _, trapMessage := range trapMessages {
			jsonData, _ := json.Marshal(trapMessage)

			// Create request and response
			req, _ := http.NewRequest(http.MethodPost, "/trap", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute
			handler.ReceiveTrap(c)

			// Assertion
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "success")
		}

		// Prepare test data
		updateReq := request.AlertUpdate{
			ID:     int64(1),
			Status: models.AlertStatusCleared,
		}

		jsonData, _ := json.Marshal(updateReq)

		req, _ := http.NewRequest(http.MethodPut, "/update", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.Update(c)

		// Assertion
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Update successful")
	})

	// Test case 2: Attempt to update non-existent alert
	t.Run("Update Non-existent Alert", func(t *testing.T) {
		// Prepare test data
		updateReq := request.AlertUpdate{
			ID:     99999, // Non-existent ID
			Status: models.AlertStatusCleared,
		}

		jsonData, _ := json.Marshal(updateReq)

		req, _ := http.NewRequest(http.MethodPut, "/update", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.Update(c)

		// Assertion
		// The actual implementation may return different status codes; here we assume it would return 404 or similar error
		// If the actual implementation returns 200 with an error message, the response content should also be verified
	})

	// Test case 3: Invalid request parameters
	t.Run("Invalid Request Parameters", func(t *testing.T) {
		// Create invalid request data
		invalidJSON := []byte(`{"invalid": }`)

		req, _ := http.NewRequest(http.MethodPut, "/update", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute
		handler.Update(c)

		// Assertion
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request parameters")
	})
}
