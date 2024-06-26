package cmd

import (
	"fmt"
	"mocking-server/config"
	"mocking-server/internal/auth"
	"mocking-server/internal/repository/postgres"
	"mocking-server/internal/repository/postgres/mock_server_repository/children"
	"mocking-server/internal/repository/postgres/mock_server_repository/collection"
	"mocking-server/internal/repository/postgres/mock_server_repository/member"
	mockdata "mocking-server/internal/repository/postgres/mock_server_repository/mock_data"
	workspace "mocking-server/internal/repository/postgres/mock_server_repository/work_space"
	"mocking-server/internal/repository/postgres/users"
	mockserver "mocking-server/internal/rest/mock_server"
	"mocking-server/internal/rest/sample"
	"mocking-server/internal/security"
	mockserversvc "mocking-server/internal/service/mockserver_svc"
	"mocking-server/internal/service/users_svc"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
)

var (
	EnvFilePath string
	rootCmd     = &cobra.Command{
		Use:   "cobra-cli",
		Short: "dummy-server",
	}
)
var (
	rootConfig    *config.Root
	database      *sqlx.DB
	sampleHandler sample.Handler
	//dummyServerHandler dummyServer.Handler

	authHandler       auth.Handler
	mockserverHandler mockserver.Handler
	middlewareService security.MiddlewareService
)

func Execute() {
	rootCmd.PersistentFlags().StringVarP(&EnvFilePath, "env", "e", ".env", ".env file to read from")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Cannot Run CLI. err > ", err)
		os.Exit(1)
	}
}
func init() {
	cobra.OnInitialize(func() {
		configReader()
		initPostgres()
		initApp()
	})
}
func configReader() {
	log.Infof("Initialize ENV")
	rootConfig = config.Load(EnvFilePath)
}
func initApp() {
	middlewareService = initMiddleware()
	sampleHandler = initSample()
	authHandler = initAuth()
	mockserverHandler = initMockServer()

}

func initPostgres() {
	log.Infof("Initialize postgress")
	var err error
	database, err = config.OpenPostgresDatabaseConnection(config.Postgres{
		Host:                  rootConfig.Postgres.Host,
		Port:                  rootConfig.Postgres.Port,
		User:                  rootConfig.Postgres.User,
		Password:              rootConfig.Postgres.Password,
		Dbname:                rootConfig.Postgres.Dbname,
		MaxConnectionLifetime: rootConfig.Postgres.MaxConnectionLifetime,
		MaxOpenConnection:     rootConfig.Postgres.MaxOpenConnection,
		MaxIdleConnection:     rootConfig.Postgres.MaxIdleConnection,
	})
	if err != nil {
		log.Errorf("Posgress failed, error: ", err)
	}
}

func initSample() sample.Handler {
	log.Infof("Initialize sample module")
	return sample.NewHandler(
		sample.NewController(
			sample.NewService(
				sample.NewRepository(database),
			),
		),
	)
}
func initMiddleware() security.MiddlewareService {
	log.Infof("Initialize middleware")
	return security.NewMiddlewareService(security.NewJwtService(rootConfig), rootConfig)
}

func initAuth() auth.Handler {
	log.Infof("Initialize auth")
	return auth.NewHandler(
		auth.NewController(
			users_svc.NewService(
				users.NewRepository(database),
				security.NewJwtService(rootConfig),
				rootConfig,
			),
		),
	)
}

func initMockServer() mockserver.Handler {
	return mockserver.NewHandler(
		mockserver.NewController(
			mockserversvc.NewService(
				postgres.NewRepository(database),
				workspace.NewRepository(database),
				member.NewRepository(database),
				mockdata.NewRepository(database),
				collection.NewRepository(database),
				children.NewRepository(database),
			),
		))
}

func initMidleware() security.MiddlewareService {
	return security.NewMiddlewareService(security.NewJwtService(rootConfig), rootConfig)
}
