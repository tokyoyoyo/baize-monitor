package service

import (
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"
	"baize-monitor/pkg/storage"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestAlertRepo() repository.BMCTrapParserRepository {
	testConf, err := config.LoadTestMockServerConfig()
	if err != nil {
		panic(err.Error())
	}
	db, err := storage.NewClient(testConf.PostGresConfig)
	if err != nil {
		panic(err.Error())
	}

	err = db.DB.Migrator().DropTable(&models.BMCTrapParser{})
	if err != nil {
		panic(err.Error())
	}

	err = db.DB.Migrator().DropTable(&models.Alert{})
	if err != nil {
		panic(err.Error())
	}
	db.AutoMigrate(&models.Alert{}, &models.BMCTrapParser{})

	return repository.NewBMCTrapParserRepository(db.DB)
}

func TestConvertToParser(t *testing.T) {
	daoModel := &models.BMCTrapParser{
		ID:         1,
		ParserName: "test-parser",
		VendorCode: "vendor123",
		VendorName: "Test Vendor",

		AlertLevelOID:     ".1.3.6.1.4.1.1234.1.1",
		AlertContentOID:   ".1.3.6.1.4.1.1234.1.2",
		AlertTimeOID:      ".1.3.6.1.4.1.1234.1.3",
		AlertComponentOID: ".1.3.6.1.4.1.1234.1.4",

		EnableAutoClose: true,
		AlertIndexOID:   ".1.3.6.1.4.1.1234.1.5",
		AlertStatusOID:  ".1.3.6.1.4.1.1234.1.6",

		EnableContactInterComponentAlerts:  true,
		ContactInterComponentIdentifierOID: ".1.3.6.1.4.1.1234.1.7",

		LevelMappings: map[string]models.AlertLevel{
			"critical": models.AlertLevelCritical,
			"warning":  models.AlertLevelWarning,
		},
		StatusMappings: map[string]models.TrapStatus{
			"active":   models.TrapStatusAsserted,
			"inactive": models.TrapStatusDeasserted,
		},
		EnableProductNameList: []string{"product1", "product2"},
		EnableHostNameList:    []string{"host1", "host2"},
		ComponentMappings: map[string][]string{
			"cpu":    {"processor", "cpu_a"},
			"memory": {"rams", "mem"},
		},
	}

	parser := ConvertToParser(daoModel)

	assert.Equal(t, daoModel.ID, parser.ID)
	assert.Equal(t, daoModel.VendorCode, parser.VendorCode)
	assert.Equal(t, daoModel.AlertLevelOID, parser.AlertLevelOID)
	assert.Equal(t, daoModel.EnableAutoClose, parser.EnableAutoClose)
	assert.Equal(t, daoModel.EnableContactInterComponentAlerts, parser.EnableContactInterComponentAlerts)

	// Verify ComponentMappingIndex construction
	expectedIndex := map[string]string{
		"processor": "cpu",
		"cpu_a":     "cpu",
		"rams":      "memory",
		"mem":       "memory",
	}
	assert.Equal(t, expectedIndex, parser.ComponentMappingIndex)

	// Verify ComponentMappingKeyWordLengthSorted sorting
	// Should be sorted by length in descending order
	expectedSorted := []string{"processor", "cpu_a", "rams", "mem"}
	assert.Equal(t, expectedSorted, parser.ComponentMappingKeyWordLengthSorted)
}

func TestBmcTrapParser_Parse(t *testing.T) {
	parser := &bmcTrapParser{
		ID:         1,
		ParserName: "test-parser",
		VendorCode: "vendor123",
		VendorName: "Test Vendor",

		AlertLevelOID:     ".1.3.6.1.4.1.1234.1.1",
		AlertContentOID:   ".1.3.6.1.4.1.1234.1.2",
		AlertTimeOID:      ".1.3.6.1.4.1.1234.1.3",
		AlertComponentOID: ".1.3.6.1.4.1.1234.1.4",

		LevelMappings: map[string]models.AlertLevel{
			"critical": models.AlertLevelCritical,
		},
		ComponentMappings: map[string][]string{
			string(models.BMC_AlertComponentCPU): {"Processor"},
		},
		ComponentMappingKeyWordLengthSorted: []string{"processor"},
		ComponentMappingIndex: map[string]string{
			"processor": string(models.BMC_AlertComponentCPU),
		},
	}

	trapData := &models.TrapMessage{
		SnmpTrapOID: ".1.99999.1",
		SourceIP:    net.IP{4, 3, 2, 1},
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",                    // AlertLevelOID
			".1.3.6.1.4.1.1234.1.2": "System temperature critical", // AlertContentOID
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",        // AlertTimeOID
			".1.3.6.1.4.1.1234.1.4": "processor Unit",              // AlertComponentOID
		},
		RawData: []byte("raw data"),
	}

	alert, err := parser.Parse(trapData)

	assert.NoError(t, err)
	assert.Equal(t, ".1.99999.1", alert.TrapOID)
	assert.Equal(t, "4.3.2.1", alert.SourceIP)
	assert.Equal(t, "vendor123", alert.VendorCode)
	assert.Equal(t, "Test Vendor", alert.VendorName)
	assert.Equal(t, int64(1), alert.ParserID)
	assert.Equal(t, []byte("raw data"), alert.RawData)
	assert.Equal(t, models.AlertLevelCritical, alert.AlertLevel)
	assert.Equal(t, "System temperature critical", alert.Content)
	assert.Equal(t, string(models.BMC_AlertComponentCPU), alert.Component)
}

func TestBmcTrapParser_Match(t *testing.T) {
	parser := &bmcTrapParser{
		AlertLevelOID:     ".1.3.6.1.4.1.1234.1.1",
		AlertContentOID:   ".1.3.6.1.4.1.1234.1.2",
		AlertTimeOID:      ".1.3.6.1.4.1.1234.1.3",
		AlertComponentOID: ".1.3.6.1.4.1.1234.1.4",
	}

	// Successful match scenario
	trapDataWithAllFields := &models.TrapMessage{
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",
			".1.3.6.1.4.1.1234.1.2": "System temperature critical",
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",
			".1.3.6.1.4.1.1234.1.4": "Processor Unit",
		},
	}

	assert.True(t, parser.Match(trapDataWithAllFields))

	// Failed match scenario - missing one field
	trapDataMissingField := &models.TrapMessage{
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",
			".1.3.6.1.4.1.1234.1.2": "System temperature critical",
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",
			// Missing AlertComponentOID
		},
	}

	assert.False(t, parser.Match(trapDataMissingField))
}

func TestParserCache_FindParser(t *testing.T) {
	repo := getTestAlertRepo()

	// Create test data
	testParser := &models.BMCTrapParser{
		ParserName: "test-parser",
		VendorCode: "1234",
		VendorName: "Test Vendor",

		AlertLevelOID:     ".1.3.6.1.4.1.1234.1.1",
		AlertContentOID:   ".1.3.6.1.4.1.1234.1.2",
		AlertTimeOID:      ".1.3.6.1.4.1.1234.1.3",
		AlertComponentOID: ".1.3.6.1.4.1.1234.1.4",

		LevelMappings: map[string]models.AlertLevel{
			"critical": models.AlertLevelCritical,
		},
		ComponentMappings: map[string][]string{
			"cpu": {"processor"},
		},
	}

	err := repo.Create(testParser)
	assert.NoError(t, err)

	// Create ParserCache instance
	cache := NewBMCTrapParserCache(repo)

	// Create matching TrapMessage
	trapData := &models.TrapMessage{
		SourceIP: net.IP{192, 168, 1, 1},
		VariableMap: map[string]string{
			".1.3.6.1.4.1.1234.1.1": "critical",
			".1.3.6.1.4.1.1234.1.2": "System temperature critical",
			".1.3.6.1.4.1.1234.1.3": "2023-08-01T12:00:00Z",
			".1.3.6.1.4.1.1234.1.4": "Processor Unit",
		},
	}

	// Find parser
	foundParser, err := cache.FindParser(trapData)
	assert.NoError(t, err)
	assert.NotNil(t, foundParser)
	assert.Equal(t, "1234", foundParser.VendorCode)

	// Test scenario where parser is not found
	invalidTrapData := &models.TrapMessage{
		SourceIP: net.IP{192, 168, 1, 1},
		VariableMap: map[string]string{
			".1.3.6.1.4.1.99999.1.1": "critical", // Non-existent OID
		},
	}

	_, err = cache.FindParser(invalidTrapData)
	assert.Error(t, err)
}

func TestParserCache_InitLoad(t *testing.T) {
	repo := getTestAlertRepo()

	// Insert test data
	testParser := &models.BMCTrapParser{
		ParserName: "init-test-parser",
		VendorCode: "5678",
		VendorName: "Init Test Vendor",

		AlertLevelOID:     ".1.3.6.1.4.1.5678.1.1",
		AlertContentOID:   ".1.3.6.1.4.1.5678.1.2",
		AlertTimeOID:      ".1.3.6.1.4.1.5678.1.3",
		AlertComponentOID: ".1.3.6.1.4.1.5678.1.4",

		LevelMappings: map[string]models.AlertLevel{
			"warning": models.AlertLevelWarning,
		},
		ComponentMappings: map[string][]string{
			"memory": {"RAM"},
		},
	}

	err := repo.Create(testParser)
	assert.NoError(t, err)

	// Create ParserCache instance
	cache := &BMCTrapParserCache{
		vendorIndex:       make(map[string][]*bmcTrapParser),
		allParsers:        make(map[int64]*bmcTrapParser),
		parserChecksums:   make(map[int64]string),
		bmcTrapParserRepo: repo,
	}

	// Call initload method
	cache.initload()

	// Verify data has been loaded
	assert.Len(t, cache.allParsers, 1)
	assert.Contains(t, cache.allParsers, int64(1))
	assert.Contains(t, cache.vendorIndex, "5678")
	assert.Len(t, cache.vendorIndex["5678"], 1)
	assert.Equal(t, "init-test-parser", cache.allParsers[1].ParserName)
}

func TestParserCache_CalculateChecksum(t *testing.T) {
	cache := &BMCTrapParserCache{}

	parser1 := &bmcTrapParser{
		ID:            1,
		ParserName:    "parser1",
		VendorCode:    "vendor1",
		AlertLevelOID: ".1.1.1.1",
	}

	parser2 := &bmcTrapParser{
		ID:            1,
		ParserName:    "parser1",
		VendorCode:    "vendor1",
		AlertLevelOID: ".1.1.1.1",
	}

	parser3 := &bmcTrapParser{
		ID:            1,
		ParserName:    "parser1",
		VendorCode:    "vendor1",
		AlertLevelOID: ".1.1.1.2", // Different OID
	}

	checksum1 := cache.calculateChecksum(parser1)
	checksum2 := cache.calculateChecksum(parser2)
	checksum3 := cache.calculateChecksum(parser3)

	assert.Equal(t, checksum1, checksum2)    // Same configuration should have same checksum
	assert.NotEqual(t, checksum1, checksum3) // Different configurations should have different checksums
}
