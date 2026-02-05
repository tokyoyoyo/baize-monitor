//go:build wireinject
// +build wireinject

package wire

import (
	"baize-monitor/internal/server"
	"baize-monitor/internal/server/alert/handler"
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/internal/server/alert/service"
	"baize-monitor/internal/server/http/routes"
	snmp "baize-monitor/internal/server/snmp"
	"baize-monitor/pkg/config"
	pkg_snmp "baize-monitor/pkg/snmp"
	storage "baize-monitor/pkg/storage"
	postgres "baize-monitor/pkg/storage/postgres"

	"github.com/google/wire"
)

func InitializeServer(serverConfigPath string) (*server.BaiZeServer, error) {
	wire.Build(
		ProvideConfig,
		ProvideDatabase,

		ProvideBMCParserRepository,
		ProvideBMCParserService,
		ProvideBMCParserHandler,
		ProvideBMCParserRouter,
		ProvideAdminServer,

		ProviderAlterRepository,
		ProviderBMCTrapParserCache,
		ProvideBMCTrapService,
		ProvideBMCTrapHandler,
		ProvideAlertRouter,
		ProvideAlertService,

		ProviderDistributedLocker,
		ProvideResponseManager,
		ProvideSNMPServer,

		ProviderBaizieServer,
	)
	return nil, nil
}

// ProvideConfig 配置
func ProvideConfig(serverConfigPath string) (*config.ServerConfig, error) {
	return config.LoadServerConfig(serverConfigPath)
}

// ProvideDatabase 数据库
func ProvideDatabase(cfg *config.ServerConfig) (*postgres.Client, error) {
	return postgres.NewClient(cfg.PostGresConfig)
}

// ProvideBMCParserRepository BMC Repository
func ProvideBMCParserRepository(db *postgres.Client) repository.BMCTrapParserRepository {
	return repository.NewBMCTrapParserRepository(db.DB)
}

// ProvideBMCParserService BMC Service
func ProvideBMCParserService(repo repository.BMCTrapParserRepository) service.BMCTrapParserService {
	return service.NewBMCTrapParserServiceImp(repo)
}

// ProvideBMCParserHandler BMC Handler
func ProvideBMCParserHandler(s service.BMCTrapParserService) handler.BMCTrapParserHandler {
	return handler.NewBMCTrapParserHandler(s)
}

func ProvideBMCParserRouter(h handler.BMCTrapParserHandler) routes.AdminRouter {
	return routes.NewAdminRouter(h)
}

func ProvideAdminServer(cfg *config.ServerConfig, r routes.AdminRouter) *server.AdminServer {
	return server.NewAdminServer(cfg, r)
}

func ProviderAlterRepository(db *postgres.Client) repository.AlertRepo {
	return repository.NewAlertRepoImp(db.DB)
}

func ProviderBMCTrapParserCache(btpr repository.BMCTrapParserRepository) *service.BMCTrapParserCache {
	return service.NewBMCTrapParserCache(btpr)
}

func ProvideBMCTrapService(parserCache *service.BMCTrapParserCache, alertRepo repository.AlertRepo) service.BMCTrapService {
	return service.NewBMCTrapServiceImp(parserCache, alertRepo)
}

func ProvideBMCTrapHandler(
	bmcTrapService service.BMCTrapService,
) handler.BMCTrapHandler {
	return handler.NewBMCTrapHandler(bmcTrapService)
}

func ProvideAlertRouter(bmcTraprH handler.BMCTrapHandler) routes.AlertRouter {
	return routes.NewAlertRouter(bmcTraprH)
}

func ProvideAlertService(cfg *config.ServerConfig, r routes.AlertRouter) (*server.AlertServer, error) {
	return server.NewAlertServer(cfg, r)
}

func ProviderDistributedLocker(cfg *config.ServerConfig) (storage.DistributedLockerInterface, error) {
	return storage.NewRedisDistributedLocker(cfg.RedisConfig)
}

func ProvideResponseManager(cfg *config.ServerConfig) pkg_snmp.ResponseManagerInterface {
	return pkg_snmp.NewResponseManager(cfg.ResponseManagerConfig)
}

func ProvideSNMPServer(cfg *config.ServerConfig,
	locker storage.DistributedLockerInterface,
	responseMgr pkg_snmp.ResponseManagerInterface,
) (*snmp.SNMPServer, error) {
	return snmp.NewSNMPServer(cfg, locker, responseMgr)
}

func ProviderBaizieServer(
	config *config.ServerConfig,
	adminServer *server.AdminServer,
	alertServer *server.AlertServer,
	snmpServer *snmp.SNMPServer,
) *server.BaiZeServer {
	return server.NewBaiZeServer(
		config,
		adminServer,
		alertServer,
		snmpServer,
	)
}
