//go:build wireinject
// +build wireinject

package wire

import (
	"baize-monitor/internal/server"
	"baize-monitor/internal/server/alert/handler"
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/internal/server/alert/service"
	"baize-monitor/internal/server/http/routes"

	user_handler "baize-monitor/internal/server/user/handler"
	user_service "baize-monitor/internal/server/user/service"

	snmp "baize-monitor/internal/server/snmp"
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"
	pkg_snmp "baize-monitor/pkg/snmp"
	"baize-monitor/pkg/storage"

	"github.com/google/wire"
	"golang.org/x/crypto/bcrypt"
)

func InitializeServer(serverConfigPath string) (*server.BaiZeServer, error) {
	wire.Build(
		ProvideConfig,
		ProvideDatabase,

		ProvideBMCParserRepository,
		ProvideBMCParserService,
		ProvideBMCParserHandler,
		ProvideAdminRouter,
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

		ProvideUserService,
		ProvideUserHandler,

		ProviderBaizieServer,
	)
	return nil, nil
}

// ProvideConfig 配置
func ProvideConfig(serverConfigPath string) (*config.ServerConfig, error) {
	return config.LoadServerConfig(serverConfigPath)
}

// ProvideDatabase 数据库
func ProvideDatabase(cfg *config.ServerConfig) (*storage.Client, error) {
	return storage.NewClient(cfg.PostGresConfig)
}

// ProvideBMCParserRepository BMC Repository
func ProvideBMCParserRepository(db *storage.Client) repository.BMCTrapParserRepository {
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

func ProvideAdminRouter(h handler.BMCTrapParserHandler, uh *user_handler.UserHandler) routes.AdminRouter {
	return routes.NewAdminRouter(h, uh)
}

func ProvideAdminServer(cfg *config.ServerConfig, r routes.AdminRouter) *server.AdminServer {
	return server.NewAdminServer(cfg, r)
}

func ProviderAlterRepository(db *storage.Client) repository.AlertRepo {
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

func ProviderDistributedLocker(db *storage.Client) (storage.DistributedLockerInterface, error) {
	return storage.NewPGTableDistributedLocker(db.DB)
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

func ProvideUserService(db *storage.Client) (user_service.UserService, error) {
	// write the default admin account during initialization.

	var count int64
	db.DB.Model(&models.User{}).Where("is_admin = ?", true).Count(&count)
	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("Bai_Ze_init_p@ss"), bcrypt.DefaultCost)
		admin := &models.User{
			Username:     "admin",
			PasswordHash: string(hashed),
			IsAdmin:      true,
			IsActive:     true,
			Permissions:  map[string]bool{"*": true},
			// admin has all permissions and does not need to be displayed, changes are not allowed
		}
		return nil, db.DB.Create(admin).Error
	}

	return user_service.NewUserService(db.DB), nil
}

func ProvideUserHandler(userService user_service.UserService) *user_handler.UserHandler {
	return user_handler.NewUserHandler(userService)
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
