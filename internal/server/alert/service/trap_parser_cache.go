package service

import (
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/pkg/models"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Parser interface for trap message parsing
type Parser interface {
	Parse(trapData *models.TrapMessage) (*models.Alert, error)
	Match(trapData *models.TrapMessage) bool
}

type bmcTrapParser struct {
	ID         int64
	ParserName string
	VendorCode string
	VendorName string

	// Core OID fields
	AlertLevelOID     string
	AlertContentOID   string
	AlertTimeOID      string
	AlertComponentOID string

	EnableAutoClose bool
	AlertIndexOID   string
	AlertStatusOID  string

	EnableContactInterComponentAlerts  bool   // Whether to enable inter-component alert correlation
	ContactInterComponentIdentifierOID string // Inter-component identifier OID

	// Mapping configurations - serialized to database as JSON
	LevelMappings         map[string]models.AlertLevel
	StatusMappings        map[string]models.TrapStatus
	EnableProductNameList []string
	EnableHostNameList    []string

	ComponentMappings                   map[string][]string
	ComponentMappingIndex               map[string]string
	ComponentMappingKeyWordLengthSorted []string
}

// Parse implements the Parser interface
func (p *bmcTrapParser) Parse(tm *models.TrapMessage) (*models.Alert, error) {
	alert := &models.Alert{
		TrapOID:    tm.SnmpTrapOID,
		SourceIP:   string(tm.SourceIP.String()),
		VendorCode: p.VendorCode,
		VendorName: p.VendorName,
		ParserID:   p.ID,
		RawData:    tm.RawData,
	}

	// Parse four core alert fields, which must succeed
	if err := p.parseLevel(alert, tm); err != nil {
		return nil, fmt.Errorf("parse bmc alert level failed: %w", err)
	}

	if err := p.processTime(alert, tm); err != nil {
		return nil, fmt.Errorf("parse bmc alert time failed: %w", err)
	}

	if err := p.parseComponent(alert, tm); err != nil {
		return nil, fmt.Errorf("parse bmc alert component failed: %w", err)
	}

	if err := p.parseContent(alert, tm); err != nil {
		return nil, fmt.Errorf("parse bmc alert content failed: %w", err)
	}

	// Additional features are optional, skip if parsing fails
	// Parse alert status field
	p.processAutoClose(alert, tm)
	// Parse inter-component identifier field
	p.processTheIdentifierOfTheSameComponent(alert, tm)

	return alert, nil
}

func (p *bmcTrapParser) Match(trapData *models.TrapMessage) bool {
	// Check if core OIDs exist in trap varbinds
	if _, exists := trapData.VariableMap[p.AlertLevelOID]; !exists {
		return false
	}
	if _, exists := trapData.VariableMap[p.AlertTimeOID]; !exists {
		return false
	}
	if _, exists := trapData.VariableMap[p.AlertContentOID]; !exists {
		return false
	}
	if _, exists := trapData.VariableMap[p.AlertComponentOID]; !exists {
		return false
	}

	return true
}

func (p *bmcTrapParser) parseLevel(alert *models.Alert, tm *models.TrapMessage) error {
	if p.AlertLevelOID == "" {
		return fmt.Errorf("please make sure bmc trap parser configuration field of alert level exists")
	}

	rawLevel, exists := tm.VariableMap[p.AlertLevelOID]
	if !exists {
		return fmt.Errorf("level OID %s not found", p.AlertLevelOID)
	}

	levelStr := fmt.Sprintf("%v", rawLevel)
	if mappedLevel, exists := p.LevelMappings[levelStr]; exists {
		alert.AlertLevel = mappedLevel
	} else {
		return fmt.Errorf("unknown level: %s", levelStr)
	}

	return nil
}

func (p *bmcTrapParser) processTime(alert *models.Alert, tm *models.TrapMessage) error {
	if p.AlertTimeOID == "" {
		return fmt.Errorf("please make sure bmc trap parser configuration field of alert time exists")
	}

	rawTime, exists := tm.VariableMap[p.AlertTimeOID]
	if !exists {
		return fmt.Errorf("time OID %s not found", p.AlertTimeOID)
	}

	alert.TrapRawTime = rawTime
	alert.AlertTime = time.Now()

	return nil
}

func (p *bmcTrapParser) parseContent(alert *models.Alert, tm *models.TrapMessage) error {
	if p.AlertContentOID == "" {
		return fmt.Errorf("please make sure bmc trap parser configuration field of alert content exists")
	}
	rawContent, exists := tm.VariableMap[p.AlertContentOID]
	if !exists {
		return fmt.Errorf("content OID %s not found", p.AlertContentOID)
	}
	alert.Content = fmt.Sprintf("%v", rawContent)
	return nil
}

func (p *bmcTrapParser) parseComponent(alert *models.Alert, tm *models.TrapMessage) error {
	if p.AlertComponentOID == "" {
		return fmt.Errorf("please make sure bmc trap parser configuration field of alert component exists")
	}
	rawComponent, exists := tm.VariableMap[p.AlertComponentOID]
	if !exists {
		return fmt.Errorf("component OID %s not found", p.AlertComponentOID)
	}

	rawLower := strings.ToLower(strings.TrimSpace(rawComponent))

	alert.Component = string(models.AlertComponentUnknown)

	for _, keyWord := range p.ComponentMappingKeyWordLengthSorted {
		if strings.Contains(rawLower, keyWord) {
			alert.Component = p.ComponentMappingIndex[keyWord]
		}
	}

	return nil
}

func (p *bmcTrapParser) processAutoClose(alert *models.Alert, tm *models.TrapMessage) {
	if !p.EnableAutoClose {
		alert.EnableAutoClose = false
		return
	}

	if p.AlertStatusOID == "" || p.AlertIndexOID == "" {
		alert.EnableAutoClose = false
		return
	}

	rawStatus, exists := tm.VariableMap[p.AlertStatusOID]
	if !exists {
		alert.EnableAutoClose = false
		return
	}

	rawIndex, exists := tm.VariableMap[p.AlertIndexOID]
	if !exists {
		alert.EnableAutoClose = false
		return
	}

	statusStr := fmt.Sprintf("%v", rawStatus)
	if mappedStatus, exists := p.StatusMappings[statusStr]; exists {
		alert.TrapStatus = mappedStatus
	} else {
		alert.EnableAutoClose = false
		return
	}
	alert.EnableAutoClose = true
	alert.TrapIndex = fmt.Sprintf("%v", rawIndex)
}

func (p *bmcTrapParser) processTheIdentifierOfTheSameComponent(alert *models.Alert, tm *models.TrapMessage) {
	if !p.EnableContactInterComponentAlerts {
		alert.EnableContactInterComponentAlerts = false
		return
	}

	if p.ContactInterComponentIdentifierOID == "" {
		alert.EnableContactInterComponentAlerts = false
		return
	}

	rawIdentifier, exists := tm.VariableMap[p.ContactInterComponentIdentifierOID]
	if !exists {
		alert.EnableContactInterComponentAlerts = false
		return
	}
	alert.IdentifierOfTheSameComponent = fmt.Sprintf("%v", rawIdentifier)
}

type BMCTrapParserCache struct {
	// Core inverted index
	vendorIndex map[string][]*bmcTrapParser // vendor_code -> parser list

	// Auxiliary indexes
	allParsers      map[int64]*bmcTrapParser // parser_id -> parser
	parserChecksums map[int64]string         // parser_id -> config checksum

	bmcTrapParserRepo repository.BMCTrapParserRepository
}

func NewBMCTrapParserCache(btpr repository.BMCTrapParserRepository) *BMCTrapParserCache {
	cache := &BMCTrapParserCache{
		vendorIndex:       make(map[string][]*bmcTrapParser),
		allParsers:        make(map[int64]*bmcTrapParser),
		parserChecksums:   make(map[int64]string),
		bmcTrapParserRepo: btpr,
	}

	cache.initload()

	go cache.autoReload()

	return cache
}

// FindParser uses inverted index to quickly find a matching parser
func (c *BMCTrapParserCache) FindParser(trap *models.TrapMessage) (*bmcTrapParser, error) {
	trapVendorCode, err := trap.ParseVendorCode()
	if err != nil {
		return nil, fmt.Errorf("failed to parse vendor code: %w", err)
	}

	// Get enterprise code first, then match against all parsers for that vendor
	parsers := c.getParsersByVendor(trapVendorCode)
	for _, parser := range parsers {
		if parser.Match(trap) {
			return parser, nil
		}
	}

	return nil, fmt.Errorf("failed to find parser")
}

// Reload rebuilds the inverted index
func (c *BMCTrapParserCache) initload() {
	parserRecords, err := c.bmcTrapParserRepo.FindAll()
	if err != nil {
		panic(fmt.Errorf("failed load bmc trap parser, err:%v", err))
	}

	// Clear all indexes
	c.vendorIndex = make(map[string][]*bmcTrapParser)
	c.allParsers = make(map[int64]*bmcTrapParser)

	for i := range parserRecords {
		pr := parserRecords[i]
		bmcTrapParserInstance := ConvertToParser(pr)
		// Calculate config checksum for change detection
		checksum := c.calculateChecksum(bmcTrapParserInstance)
		c.parserChecksums[bmcTrapParserInstance.ID] = checksum

		// Add to vendor index
		if c.vendorIndex[bmcTrapParserInstance.VendorCode] == nil {
			c.vendorIndex[bmcTrapParserInstance.VendorCode] = []*bmcTrapParser{}
		}
		c.vendorIndex[bmcTrapParserInstance.VendorCode] = append(c.vendorIndex[bmcTrapParserInstance.VendorCode], bmcTrapParserInstance)
		c.allParsers[bmcTrapParserInstance.ID] = bmcTrapParserInstance
	}
}

// calculateChecksum calculates the checksum of parser configuration
func (c *BMCTrapParserCache) calculateChecksum(parser *bmcTrapParser) string {
	bytes, _ := json.Marshal(parser)
	return fmt.Sprintf("%x", md5.Sum(bytes))
}

// GetParsersByVendor gets all parsers for a specific vendor
func (c *BMCTrapParserCache) getParsersByVendor(vendorCode string) []*bmcTrapParser {
	if parsers, exists := c.vendorIndex[vendorCode]; exists {
		return parsers
	}
	return []*bmcTrapParser{}
}

// autoReload automatically reloads parsers periodically
func (c *BMCTrapParserCache) autoReload() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bmcTrapParserList, err := c.bmcTrapParserRepo.FindAll()
		if err != nil {
			continue
		}

		parserIDSet := map[int64]bool{}
		for _, parser := range bmcTrapParserList {
			parserInstance := ConvertToParser(parser)
			checksum := c.calculateChecksum(parserInstance)
			parserIDSet[parserInstance.ID] = true
			if checksum != c.parserChecksums[parser.ID] {
				// Configuration has changed
				c.updateParser(parser.ID, parser, checksum)
			}
		}

		for oldParserID, oldParser := range c.allParsers {
			_, exists := parserIDSet[oldParserID]
			if !exists {
				// This parser has been soft-deleted
				c.removeFromVendorIndex(oldParser)
				delete(c.parserChecksums, oldParserID)
				delete(c.allParsers, oldParserID)
			}
		}
	}
}

func (c *BMCTrapParserCache) updateParser(parserID int64, parser *models.BMCTrapParser, newChecksum string) {
	parserInstance := ConvertToParser(parser)

	// Remove from old vendor index
	oldParser, exists := c.allParsers[parserID]
	if exists {
		c.removeFromVendorIndex(oldParser)
	}

	// Update checksum
	c.parserChecksums[parserID] = newChecksum

	// Add to new vendor index
	if c.vendorIndex[parserInstance.VendorCode] == nil {
		c.vendorIndex[parserInstance.VendorCode] = []*bmcTrapParser{}
	}
	c.vendorIndex[parserInstance.VendorCode] = append(c.vendorIndex[parserInstance.VendorCode], parserInstance)
	c.allParsers[parserID] = parserInstance
}

func (c *BMCTrapParserCache) removeFromVendorIndex(parser *bmcTrapParser) {
	parsers := c.vendorIndex[parser.VendorCode]
	for i, p := range parsers {
		if p.ID == parser.ID {
			// Remove element from slice
			c.vendorIndex[parser.VendorCode] = append(parsers[:i], parsers[i+1:]...)
			break
		}
	}
}

// ConvertToParser converts DAO model to parser instance
func ConvertToParser(daoModel *models.BMCTrapParser) *bmcTrapParser {
	if daoModel == nil {
		return nil
	}

	p := &bmcTrapParser{
		ID:         daoModel.ID,
		ParserName: daoModel.ParserName,
		VendorCode: daoModel.VendorCode,
		VendorName: daoModel.VendorName,

		AlertLevelOID:     daoModel.AlertLevelOID,
		AlertContentOID:   daoModel.AlertContentOID,
		AlertTimeOID:      daoModel.AlertTimeOID,
		AlertComponentOID: daoModel.AlertComponentOID,

		EnableAutoClose: daoModel.EnableAutoClose,
		AlertIndexOID:   daoModel.AlertIndexOID,
		AlertStatusOID:  daoModel.AlertStatusOID,

		EnableContactInterComponentAlerts:  daoModel.EnableContactInterComponentAlerts,
		ContactInterComponentIdentifierOID: daoModel.ContactInterComponentIdentifierOID,

		LevelMappings:         daoModel.LevelMappings,
		StatusMappings:        daoModel.StatusMappings,
		EnableProductNameList: daoModel.EnableProductNameList,
		EnableHostNameList:    daoModel.EnableHostNameList,
		ComponentMappings:     daoModel.ComponentMappings,
	}

	keyWords := []string{}

	p.ComponentMappingIndex = make(map[string]string)
	for component, mappingKeyWordList := range p.ComponentMappings {
		for _, mappingKeyWord := range mappingKeyWordList {
			mappingKeyWord = strings.ToLower(mappingKeyWord)
			p.ComponentMappingIndex[mappingKeyWord] = component
			keyWords = append(keyWords, mappingKeyWord)
		}
	}

	sort.Slice(keyWords, func(i, j int) bool {
		return len(keyWords[i]) > len(keyWords[j])
	})

	p.ComponentMappingKeyWordLengthSorted = keyWords

	return p
}
