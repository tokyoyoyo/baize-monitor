package repository

import (
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"baize-monitor/pkg/storage"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AlertRepositoryTestSuite struct {
	suite.Suite
	db   *storage.Client
	repo AlertRepo
}

// SetupSuite initializes test database and repository
func (suite *AlertRepositoryTestSuite) SetupSuite() {
	db, err := getAlertTestDB()
	if err != nil {
		suite.T().Fatalf("Failed to get test DB: %v", err)
	}

	// Clean and migrate Alert table
	db.DB.Migrator().DropTable(&models.Alert{})
	if err := db.AutoMigrate(&models.Alert{}); err != nil {
		suite.T().Fatalf("Failed to migrate Alert: %v", err)
	}

	suite.db = db
	suite.repo = NewAlertRepoImp(db.DB)
}

// SetupTest cleans alerts table before each test
func (suite *AlertRepositoryTestSuite) SetupTest() {
	suite.db.DB.Exec("DELETE FROM alerts")
}

// TearDownTest cleans alerts table after each test
func (suite *AlertRepositoryTestSuite) TearDownTest() {
	suite.db.DB.Exec("DELETE FROM alerts")
}

// TearDownSuite closes database connection
func (suite *AlertRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, err := suite.db.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// Helper to get test database client
func getAlertTestDB() (*storage.Client, error) {
	testConf, err := config.LoadTestMockServerConfig()
	if err != nil {
		return nil, err
	}
	return storage.NewClient(testConf.PostGresConfig)
}

// ==================== TEST CASES ====================

func (suite *AlertRepositoryTestSuite) TestCreateAlert_Success() {
	// Arrange
	alert := &models.Alert{
		ParserID:    1001,
		AlertStatus: models.AlertStatusActive,
		TrapOID:     "1.3.6.1.4.1.12345.1.1",
		SourceIP:    "192.168.1.100",
		VendorCode:  "DELL",
		VendorName:  "Dell Inc.",
		AlertLevel:  "CRITICAL",
		AlertTime:   time.Now(),
		Component:   "CPU",
		Content:     "CPU temperature critical",
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",                    // AlertLevelOID
			".1.3.6.1.4.1.1234.1.2": "System temperature critical", // AlertContentOID
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",        // AlertTimeOID
			".1.3.6.1.4.1.1234.1.4": "processor Unit",              // AlertComponentOID
		},
	}

	// Act
	err := suite.repo.Create(alert)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), alert.ID)
	assert.WithinDuration(suite.T(), time.Now(), alert.CreatedAt, 2*time.Second)
}

func (suite *AlertRepositoryTestSuite) TestUpdateAlert_Success() {
	// Arrange
	original := &models.Alert{
		ParserID:    1002,
		AlertStatus: models.AlertStatusActive,
		TrapOID:     "1.3.6.1.4.1.12345.1.2",
		SourceIP:    "10.0.0.50",
		VendorCode:  "HP",
		AlertLevel:  models.AlertLevelWarning,
		AlertTime:   time.Now().Add(-1 * time.Hour),
		Content:     "Disk warning",
	}
	suite.repo.Create(original)

	// Act: Update status and content
	original.AlertStatus = models.AlertStatusCleared
	original.Content = "Disk issue resolved"
	err := suite.repo.Update(original)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify update persisted
	updated, _ := suite.repo.FindByID(original.ID)
	assert.Equal(suite.T(), models.AlertStatusCleared, updated.AlertStatus)
	assert.Equal(suite.T(), "Disk issue resolved", updated.Content)
	assert.WithinDuration(suite.T(), time.Now(), updated.UpdatedAt, 2*time.Second)
}

func (suite *AlertRepositoryTestSuite) TestFindByID_Success() {
	// Arrange
	alert := &models.Alert{
		ParserID:    1003,
		AlertStatus: models.AlertStatusActive,
		TrapOID:     "1.3.6.1.4.1.12345.1.3",
		SourceIP:    "172.16.0.10",
		VendorCode:  "IBM",
		AlertLevel:  models.AlertLevelInfo,
		AlertTime:   time.Now(),
		Content:     "System boot completed",
	}
	suite.repo.Create(alert)

	// Act
	result, err := suite.repo.FindByID(alert.ID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), alert.ID, result.ID)
	assert.Equal(suite.T(), "System boot completed", result.Content)
}

func (suite *AlertRepositoryTestSuite) TestFindByID_NotFound() {
	// Act
	result, err := suite.repo.FindByID(999999999)

	// Assert
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
	assert.NotNil(suite.T(), result) // Note: Returns zero-value pointer per implementation
}

func (suite *AlertRepositoryTestSuite) TestList_FilterByStatus() {
	// Arrange
	now := time.Now()
	suite.createTestAlerts([]models.Alert{
		{ParserID: 1, AlertStatus: models.AlertStatusActive, SourceIP: "1.1.1.1", AlertTime: now},
		{ParserID: 2, AlertStatus: models.AlertStatusCleared, SourceIP: "2.2.2.2", AlertTime: now},
		{ParserID: 3, AlertStatus: models.AlertStatusActive, SourceIP: "3.3.3.3", AlertTime: now},
	})

	status := string(models.AlertStatusActive)
	filter := &request.AlertFilter{
		Status:   &status,
		Page:     1,
		PageSize: 10,
	}

	// Act
	result, err := suite.repo.List(filter)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), result.Total)
	assert.Len(suite.T(), result.List, 2)
	for _, a := range result.List {
		assert.Equal(suite.T(), models.AlertStatusActive, a.AlertStatus)
	}
}

func (suite *AlertRepositoryTestSuite) TestList_FilterByTimeRange() {
	// Arrange
	baseTime := time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC)
	suite.createTestAlerts([]models.Alert{
		{AlertTime: baseTime.Add(-24 * time.Hour), SourceIP: "past.example.com"},
		{AlertTime: baseTime, SourceIP: "current.example.com"},
		{AlertTime: baseTime.Add(24 * time.Hour), SourceIP: "future.example.com"},
	})

	start := baseTime.Add(-1 * time.Hour)
	end := baseTime.Add(1 * time.Hour)
	filter := &request.AlertFilter{
		AlertTimeRangeStart: &start,
		AlertTimeRangeEnd:   &end,
		Page:                1,
		PageSize:            10,
	}

	// Act
	result, err := suite.repo.List(filter)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.List, 1)
	assert.Equal(suite.T(), "current.example.com", result.List[0].SourceIP)
}

func (suite *AlertRepositoryTestSuite) TestList_PaginationAndSorting() {
	// Arrange: Create 15 alerts with sequential timestamps
	alerts := make([]models.Alert, 0)
	for i := 1; i <= 15; i++ {
		alerts = append(alerts, models.Alert{
			ParserID:    int64(i),
			AlertStatus: models.AlertStatusActive,
			SourceIP:    "192.168.1." + string(rune(i+48)),
			AlertTime:   time.Now().Add(time.Duration(-i) * time.Hour),
			Content:     "Alert " + string(rune(i+48)),
		})
	}
	suite.createTestAlerts(alerts)

	// Page 1
	filter1 := &request.AlertFilter{Page: 1, PageSize: 5}
	result1, _ := suite.repo.List(filter1)
	assert.Len(suite.T(), result1.List, 5)
	assert.Equal(suite.T(), int64(15), result1.Total)
	// Verify descending order by created_at
	for i := 0; i < 4; i++ {
		assert.True(suite.T(), result1.List[i].CreatedAt.After(result1.List[i+1].CreatedAt))
	}

	// Page 2
	filter2 := &request.AlertFilter{Page: 2, PageSize: 5}
	result2, _ := suite.repo.List(filter2)
	assert.Len(suite.T(), result2.List, 5)
	assert.NotEqual(suite.T(), result1.List[0].ID, result2.List[0].ID)

	// Page 3 (last page)
	filter3 := &request.AlertFilter{Page: 3, PageSize: 5}
	result3, _ := suite.repo.List(filter3)
	assert.Len(suite.T(), result3.List, 5)

	// Page 4 (beyond total)
	filter4 := &request.AlertFilter{Page: 4, PageSize: 5}
	result4, _ := suite.repo.List(filter4)
	assert.Len(suite.T(), result4.List, 0)
}

func (suite *AlertRepositoryTestSuite) TestList_FuzzySearch() {
	// Arrange
	suite.createTestAlerts([]models.Alert{
		{VendorName: "SuperMicro Computer", Content: "Memory error on DIMM_A1"},
		{VendorName: "Dell Technologies", Content: "Fan failure detected"},
		{VendorName: "HPE", Content: "Power supply warning"},
	})

	contentQuery := "Memory"
	filter := &request.AlertFilter{
		Content:  &contentQuery,
		Page:     1,
		PageSize: 10,
	}

	// Act
	result, err := suite.repo.List(filter)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Contains(suite.T(), result.List[0].Content, "Memory")
}

func (suite *AlertRepositoryTestSuite) TestList_EmptyResult() {
	// Arrange: No data in DB
	filter := &request.AlertFilter{
		VendorCode: strPtr("NONEXISTENT"),
		Page:       1,
		PageSize:   10,
	}

	// Act
	result, err := suite.repo.List(filter)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), result.Total)
	assert.Empty(suite.T(), result.List)
	assert.Equal(suite.T(), 1, result.Page)
	assert.Equal(suite.T(), 10, result.PageSize)
}

// Helper to create multiple test alerts
func (suite *AlertRepositoryTestSuite) createTestAlerts(alerts []models.Alert) {
	for i := range alerts {
		// Set minimal required fields if not provided
		if alerts[i].SourceIP == "" {
			alerts[i].SourceIP = "127.0.0.1"
		}
		if alerts[i].AlertTime.IsZero() {
			alerts[i].AlertTime = time.Now()
		}
		if alerts[i].AlertStatus == "" {
			alerts[i].AlertStatus = models.AlertStatusActive
		}
		suite.repo.Create(&alerts[i])
	}
}

// Helper to create string pointer
func strPtr(s string) *string { return &s }

// Run test suite
func TestAlertRepositorySuite(t *testing.T) {
	suite.Run(t, new(AlertRepositoryTestSuite))
}
