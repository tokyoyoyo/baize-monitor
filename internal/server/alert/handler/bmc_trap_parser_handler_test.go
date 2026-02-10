package handler

import (
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/internal/server/alert/service"
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"baize-monitor/pkg/storage"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func getTestSer() *service.BMCTrapParserServiceImp {
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

	repo := repository.NewBMCTrapParserRepository(db.DB)

	return service.NewBMCTrapParserServiceImp(repo)
}

// TestBMCTrapParserHandler_Create tests various scenarios for creating a BMC Trap parser
func TestBMCTrapParserHandler_Create(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Get test service instance
	testService := getTestSer()
	handler := NewBMCTrapParserHandler(testService)

	tests := []struct {
		name           string
		requestBody    request.BMCTrapParserCreate
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "Valid request",
			requestBody: request.BMCTrapParserCreate{
				VendorName: "Dell",
				ParserName: "Dell Parser v1",

				AlertLevelOID:     ".1.3.6.1.4.1.674.10892.5.4.200.10.1",
				AlertContentOID:   ".1.3.6.1.4.1.674.10892.5.4.200.10.2",
				AlertTimeOID:      ".1.3.6.1.4.1.674.10892.5.4.200.10.3",
				AlertComponentOID: ".1.3.6.1.4.1.674.10892.5.4.200.10.4",

				EnableAutoClose: false, // Explicitly set
				AlertIndexOID:   "a",   // Can be empty, because EnableAutoClose=false
				AlertStatusOID:  "a",   // Same as above

				EnableContactInterComponentAlerts:  false, // Explicitly set
				ContactInterComponentIdentifierOID: "a",   // Can be empty

				TimeFormat: "2006-01-02 15:04:05",

				LevelMappings: map[string]string{
					"1": "critical",
					"2": "warning",
					"4": "info",
				},
				StatusMappings: map[string]string{"a": "b"}, // Or empty map, because EnableAutoClose=false

				EnableProductNameList: []string{"a"}, // Required, can be empty
				EnableHostNameList:    []string{"a"}, // Required, can be empty

				ComponentMappings: map[string][]string{
					"CPU": {"CPU1", "CPU2"},
				},

				Description: "Dell BMC Trap Parser for server monitoring",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid OID format",
			requestBody: request.BMCTrapParserCreate{
				VendorName:        "HP vendor",
				ParserName:        "HP Parser v1",
				AlertLevelOID:     "invalid.oid.format",
				AlertContentOID:   ".1.3.6.1.4.1.232.10.2",
				AlertTimeOID:      ".1.3.6.1.4.1.232.10.3",
				AlertComponentOID: ".1.3.6.1.4.1.232.10.4",
				TimeFormat:        "2006-01-02 15:04:05",
				LevelMappings: map[string]string{
					"1": "critical",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty vendor name",
			requestBody: request.BMCTrapParserCreate{
				VendorName:        "", // Empty vendor name
				AlertLevelOID:     ".1.3.6.1.4.1.674.10892.5.4.200.10.1",
				AlertContentOID:   ".1.3.6.1.4.1.674.10892.5.4.200.10.2",
				AlertTimeOID:      ".1.3.6.1.4.1.674.10892.5.4.200.10.3",
				AlertComponentOID: ".1.3.6.1.4.1.674.10892.5.4.200.10.4",
				TimeFormat:        "2006-01-02 15:04:05",
				LevelMappings: map[string]string{
					"1": "critical",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert request body to JSON
			jsonValue, _ := json.Marshal(tt.requestBody)

			// Create HTTP request and response recorder
			req, _ := http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer(jsonValue))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create gin context and call handler function
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.Create(c)

			// Assert response status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestBMCTrapParserHandler_Create_InvalidJSON tests invalid JSON input
func TestBMCTrapParserHandler_Create_InvalidJSON(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Get test service instance
	testService := getTestSer()
	handler := NewBMCTrapParserHandler(testService)

	// Create request with invalid JSON
	req, _ := http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer([]byte("{invalid json}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create gin context and call handler function
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Create(c)

	// Assert response status code is 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check if there is an error field
	assert.Contains(t, response, "error")
}

// TestBMCTrapParserHandler_Create_DuplicateVendor tests duplicate vendor names
func TestBMCTrapParserHandler_Create_DuplicateVendor(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Get test service instance
	testService := getTestSer()
	handler := NewBMCTrapParserHandler(testService)

	// First create a valid parser
	validRequest := request.BMCTrapParserCreate{
		VendorName:        "UniqueVendorTest",
		ParserName:        "test",
		AlertLevelOID:     ".1.3.6.1.4.1.674.10892.5.4.200.10.1",
		AlertContentOID:   ".1.3.6.1.4.1.674.10892.5.4.200.10.2",
		AlertTimeOID:      ".1.3.6.1.4.1.674.10892.5.4.200.10.3",
		AlertComponentOID: ".1.3.6.1.4.1.674.10892.5.4.200.10.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
		},
	}

	// Create first parser
	jsonValue, _ := json.Marshal(validRequest)
	req1, _ := http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer(jsonValue))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()

	c1, _ := gin.CreateTestContext(w1)
	c1.Request = req1

	handler.Create(c1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	// Try to create again with the same parser name, should fail
	req2, _ := http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer(jsonValue))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req2

	handler.Create(c2)

	// Assert second creation fails because vendor name already exists
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, w2.Code, http.StatusBadRequest)

}

// TestBMCTrapParserHandler_Update tests updating a BMC Trap parser
func TestBMCTrapParserHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testService := getTestSer()
	handler := NewBMCTrapParserHandler(testService)

	// First create a parser for update test
	createReq := request.BMCTrapParserCreate{
		VendorName:        "UpdateTestVendor",
		ParserName:        "Update Test Parser",
		AlertLevelOID:     ".1.3.6.1.4.1.674.10892.5.4.200.10.1",
		AlertContentOID:   ".1.3.6.1.4.1.674.10892.5.4.200.10.2",
		AlertTimeOID:      ".1.3.6.1.4.1.674.10892.5.4.200.10.3",
		AlertComponentOID: ".1.3.6.1.4.1.674.10892.5.4.200.10.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
		},
	}

	jsonValue, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler.Create(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get ID from response (assuming ID is in data field)
	var id int64
	list, code, err := testService.List(&request.BMCTrapParserFilter{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NotEmpty(t, list.List)
	assert.Equal(t, len(list.List), 1)
	id = list.List[0].ID

	// Prepare update request
	parserName := "Updated Parser Name"
	alertOid := ".1.3.6.1.4.1.674.10892.5.4.200.10.5"
	updateReq := request.BMCTrapParserUpdate{
		ID:            id,
		ParserName:    &parserName,
		AlertLevelOID: &alertOid, // Modify OID
	}

	jsonValue, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/bmc-trap-parser/%d", id), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.AddParam("id", fmt.Sprintf("%d", id))

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var updateResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
	assert.NoError(t, err)
	assert.Contains(t, updateResponse["message"], "Update successful")
}

// TestBMCTrapParserHandler_Update_InvalidID tests updating with invalid ID
func TestBMCTrapParserHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testService := getTestSer()
	handler := NewBMCTrapParserHandler(testService)

	parserName := "Updated Parser Name"
	alertOid := ".1.3.6.1.4.1.674.10892.5.4.200.10.5"
	updateReq := request.BMCTrapParserUpdate{
		ID:            99999, // Non-existent ID
		ParserName:    &parserName,
		AlertLevelOID: &alertOid,
	}

	jsonValue, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/bmc-trap-parser/99999", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.AddParam("id", "99999")

	handler.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBMCTrapParserHandler_Delete tests deleting a BMC Trap parser
func TestBMCTrapParserHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testService := getTestSer()
	handler := NewBMCTrapParserHandler(testService)

	// First create a parser for delete test
	createReq := request.BMCTrapParserCreate{
		VendorName:        "DeleteTestVendor",
		ParserName:        "Delete Test Parser",
		AlertLevelOID:     ".1.3.6.1.4.1.674.10892.5.4.200.10.1",
		AlertContentOID:   ".1.3.6.1.4.1.674.10892.5.4.200.10.2",
		AlertTimeOID:      ".1.3.6.1.4.1.674.10892.5.4.200.10.3",
		AlertComponentOID: ".1.3.6.1.4.1.674.10892.5.4.200.10.4",
		TimeFormat:        "2006-01-02 15:04:05",
		LevelMappings: map[string]string{
			"1": "critical",
		},
	}

	jsonValue, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler.Create(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get created parser ID
	var id int64
	list, code, err := testService.List(&request.BMCTrapParserFilter{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NotEmpty(t, list.List)
	assert.Equal(t, len(list.List), 1)
	id = list.List[0].ID

	// Delete this parser
	updateReq := request.BMCTrapParserUpdate{
		ID: id,
	}
	jsonValue, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer(jsonValue))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var deleteResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &deleteResponse)
	assert.NoError(t, err)
}

// TestBMCTrapParserHandler_Delete_NotFound tests deleting a non-existent parser
func TestBMCTrapParserHandler_Delete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testService := getTestSer()
	handler := NewBMCTrapParserHandler(testService)

	updateReq := request.BMCTrapParserUpdate{
		ID: int64(9999),
	}
	jsonValue, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code) // Even if not found it's 200 status code
}

// TestBMCTrapParserHandler_List tests getting the list of BMC Trap parsers
func TestBMCTrapParserHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testService := getTestSer()
	handler := NewBMCTrapParserHandler(testService)

	// Create multiple parsers for list test
	vendors := []string{"ListTestVendor1", "ListTestVendor2", "ListTestVendor3"}
	for i, vendor := range vendors {
		createReq := request.BMCTrapParserCreate{
			VendorName:        vendor,
			ParserName:        fmt.Sprintf("List Test Parser %d", i+1),
			AlertLevelOID:     fmt.Sprintf(".1.3.6.1.4.1.674.10892.5.4.200.10.%d", i+1),
			AlertContentOID:   fmt.Sprintf(".1.3.6.1.4.1.674.10892.5.4.200.10.%d", i+4),
			AlertTimeOID:      fmt.Sprintf(".1.3.6.1.4.1.674.10892.5.4.200.10.%d", i+7),
			AlertComponentOID: fmt.Sprintf(".1.3.6.1.4.1.674.10892.5.4.200.10.%d", i+10),
			TimeFormat:        "2006-01-02 15:04:05",
			LevelMappings: map[string]string{
				"1": "critical",
			},
		}

		jsonValue, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/bmc-trap-parser", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.Create(c)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Test case 1: Successful list retrieval
	t.Run("Successful List Retrieval", func(t *testing.T) {
		filter := request.BMCTrapParserFilter{
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
		filter := request.BMCTrapParserFilter{
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
	t.Run("Zero Request Parameters", func(t *testing.T) {
		filter := request.BMCTrapParserFilter{
			Page:     1,
			PageSize: 0, // Invalid page size
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

	// Test case 5: PageSize exceeds maximum limit
	t.Run("PageSize Exceeds Maximum", func(t *testing.T) {
		filter := request.BMCTrapParserFilter{
			Page:     1,
			PageSize: 101, // Exceeds maximum value of 100
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

	// Test case 6: Empty filter conditions (all pointer fields are nil)
	t.Run("Empty Filter Conditions", func(t *testing.T) {
		filter := request.BMCTrapParserFilter{
			Page:     1,
			PageSize: 10,
			// All optional filter fields are nil
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
		
		// Verify returned data structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "data")
		
		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "list")
		assert.Contains(t, data, "total")
		assert.Contains(t, data, "page")
		assert.Contains(t, data, "page_size")
		
		list := data["list"].([]interface{})
		assert.True(t, len(list) >= 3, "Should have at least 3 items in list")
	})
}