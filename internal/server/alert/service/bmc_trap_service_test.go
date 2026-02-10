package service

import (
	"net"
	"net/http"
	"testing"
	"time"

	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"baize-monitor/pkg/storage"

	"github.com/stretchr/testify/assert"
)

func setupTest(_ *testing.T) (BMCTrapService, *BMCTrapParserCache, repository.BMCTrapParserRepository) {
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
	pc := NewBMCTrapParserCache(parserRepo)
	service := NewBMCTrapServiceImp(pc, alertRepo)

	return service, pc, parserRepo
}

func TestBMCTrapServiceImp_Process_NoParser(t *testing.T) {
	service, parserCache, _ := setupTest(t)

	tm := &models.TrapMessage{
		SourceIP:    net.ParseIP("192.168.1.100"),
		SnmpTrapOID: ".1.3.6.1.4.1.1234.1.1.0",
		SourceType:  models.TrapSourceTypeBMC,
		ReceivedAt:  time.Now(),
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",                    // AlertLevelOID
			".1.3.6.1.4.1.1234.1.2": "System temperature critical", // AlertContentOID
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",        // AlertTimeOID
			".1.3.6.1.4.1.1234.1.4": "processor Unit",              // AlertComponentOID
		},
		RawData: []byte("test"),
	}
	parserCache.initload()

	status, err := service.Process(tm)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, err)

	ip := tm.SourceIP.String()

	// 添加短暂延时确保数据已写入数据库
	time.Sleep(10 * time.Millisecond)

	filter := &request.AlertFilter{SourceIP: &ip, Page: 1, PageSize: 10}
	res, _, err := service.List(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res.Total)
	assert.Equal(t, string(models.AlertComponentUnknown), res.List[0].Component)
}

func TestBMCTrapServiceImp_Process_NewAlert(t *testing.T) {
	service, parserCache, parserRepo := setupTest(t)

	tm := &models.TrapMessage{
		SourceIP:   net.IP{192, 168, 1, 100},
		SourceType: models.TrapSourceTypeBMC,
		ReceivedAt: time.Now(),
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",
			".1.3.6.1.4.1.1234.1.2": "System temperature critical",
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",
			".1.3.6.1.4.1.1234.1.4": "Power Supply",
		},
	}

	status, err := service.Process(tm)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, err)
	ip := tm.SourceIP.String()

	filter := &request.AlertFilter{
		Page:     1,
		PageSize: 10,
		SourceIP: &ip}
	res, _, err := service.List(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res.Total)
	assert.Equal(t, string(models.AlertComponentUnknown), res.List[0].Component)
	assert.Equal(t, models.AlertStatusActive, res.List[0].AlertStatus)

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
	parserCache.initload()

	status, err = service.Process(tm)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, err)

	res, _, err = service.List(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)
	assert.Equal(t, string(models.BMC_AlertComponentPower), res.List[0].Component)
	assert.Equal(t, models.AlertStatusActive, res.List[0].AlertStatus)
}

func TestBMCTrapServiceImp_Process_AutoCloseAlert(t *testing.T) {
	service, parserCache, parserRepo := setupTest(t)

	parser := &models.BMCTrapParser{
		ParserName: "test1",
		VendorCode: "1234",
		VendorName: "IBM",

		AlertLevelOID:     ".1.3.6.1.4.1.1234.1.1",
		AlertContentOID:   ".1.3.6.1.4.1.1234.1.2",
		AlertTimeOID:      ".1.3.6.1.4.1.1234.1.3",
		AlertComponentOID: ".1.3.6.1.4.1.1234.1.4",
		AlertStatusOID:    ".1.3.6.1.4.1.1234.1.5",
		AlertIndexOID:     ".1.3.6.1.4.1.1234.1.6",

		ComponentMappings: map[string][]string{
			string(models.BMC_AlertComponentPower): {"Power", "PSU"},
		},
		LevelMappings: map[string]models.AlertLevel{
			"critical": models.AlertLevelCritical,
		},
		EnableAutoClose: true,
		StatusMappings: map[string]models.TrapStatus{
			"1": models.TrapStatusAsserted,
			"0": models.TrapStatusDeasserted,
		},
	}

	assert.NoError(t, parserRepo.Create(parser))
	parserCache.initload()

	netIP := net.IP{192, 168, 1, 100}

	tmActive := &models.TrapMessage{
		SourceIP:   netIP,
		SourceType: models.TrapSourceTypeBMC,
		ReceivedAt: time.Now(),
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",
			".1.3.6.1.4.1.1234.1.2": "System temperature critical",
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",
			".1.3.6.1.4.1.1234.1.4": "Power Supply",
			".1.3.6.1.4.1.1234.1.5": "1",
			".1.3.6.1.4.1.1234.1.6": "101",
		},
	}

	status, err := service.Process(tmActive)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, err)
	ip := tmActive.SourceIP.String()
	filter := &request.AlertFilter{
		Page:     1,
		PageSize: 10,
		SourceIP: &ip}
	res, _, err := service.List(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res.Total)
	assert.Equal(t, string(models.BMC_AlertComponentPower), res.List[0].Component)
	assert.Equal(t, models.AlertStatusActive, res.List[0].AlertStatus)

	tmDeactive := &models.TrapMessage{
		SourceIP:   netIP,
		SourceType: models.TrapSourceTypeBMC,
		ReceivedAt: time.Now(),
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",
			".1.3.6.1.4.1.1234.1.2": "System temperature critical",
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",
			".1.3.6.1.4.1.1234.1.4": "Power Supply",
			".1.3.6.1.4.1.1234.1.5": "0",
			".1.3.6.1.4.1.1234.1.6": "101",
		},
	}
	status, err = service.Process(tmDeactive)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, err)
	res, _, err = service.List(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res.Total)
	assert.Equal(t, models.AlertStatusCleared, res.List[0].AlertStatus)
	assert.Equal(t, string(models.BMC_AlertComponentPower), res.List[0].Component)
	assert.Equal(t, "101", res.List[0].TrapIndex)
	assert.Equal(t, string(models.BMC_AlertComponentPower), res.List[0].Component)

}

func TestBMCTrapServiceImp_Update_ExistingAlert(t *testing.T) {
	service, _, _ := setupTest(t)

	alert := &models.Alert{
		SourceIP:    "192.168.1.100",
		SourceType:  "BMC",
		Component:   "Power",
		AlertStatus: models.AlertStatusActive,
		AlertLevel:  models.AlertLevelCritical,
		AlertTime:   time.Now(),
	}
	assert.NoError(t, service.(*BMCTrapServiceImp).repo.Create(alert))

	req := request.AlertUpdate{
		ID:     alert.ID,
		Status: models.AlertStatusCleared,
	}
	status, err := service.Update(req)

	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, err)

	updated, err := service.(*BMCTrapServiceImp).repo.FindByID(alert.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.AlertStatusCleared, updated.AlertStatus)
}

func TestBMCTrapServiceImp_Update_NonExistingAlert(t *testing.T) {
	service, _, _ := setupTest(t)

	req := request.AlertUpdate{
		ID:     99999,
		Status: models.AlertStatusCleared,
	}
	status, err := service.Update(req)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alert not found")
}
