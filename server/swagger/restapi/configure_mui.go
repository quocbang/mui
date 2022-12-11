// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"
	"github.com/gorilla/handlers"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"

	"gitlab.kenda.com.tw/kenda/mcom"

	"gitlab.kenda.com.tw/kenda/mui/server"
	"gitlab.kenda.com.tw/kenda/mui/server/configs"
	mcomImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils/role"
	"gitlab.kenda.com.tw/kenda/mui/server/middleware"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations"
)

var (
	options        = new(configs.Options)
	configurations = new(configs.Configs)
)

func configureFlags(api *operations.MuiAPI) {
	api.CommandLineOptionsGroups = append(api.CommandLineOptionsGroups, swag.CommandLineOptionsGroup{
		ShortDescription: "Configuration Options",
		LongDescription:  "Configuration Options",
		Options:          options,
	})
}

func parseConfigurations(filePath string) (*configs.Configs, error) {
	configsFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfgs configs.Configs
	if err := yaml.UnmarshalStrict(configsFile, &cfgs); err != nil {
		return nil, err
	}

	return &cfgs, nil
}

func configureAPI(api *operations.MuiAPI) http.Handler {
	var err error
	// Read and parse server configuration settings
	configurations, err = parseConfigurations(options.ServerConfig)
	if err != nil {
		log.Fatalf("failed to parse configurations file. err= %s", err.Error())
	}

	setLogger(configurations.DevMode) // api.Logger will be using log.Printf

	if err := role.InitPermission(configurations.FunctionRolePermissions); err != nil {
		zap.L().Fatal("failed to initialize permission", zap.Error(err))
	}

	api.ServeError = errors.ServeError
	api.UseSwaggerUI() // for documentation on /docs
	api.JSONConsumer = runtime.JSONConsumer()
	api.JSONProducer = runtime.JSONProducer()

	dm, err := server.RegisterDataManager(*configurations)
	if err != nil {
		zap.L().Fatal("failed to register data manager", zap.Error(err))
	}

	// [NOTE] if you want try a test without real api, please switch import path from `/server/impl/mcom` to `/server/impl/mock`
	serviceConfig := mcomImpl.ServiceConfig{
		TokenLifeTime:         time.Duration(configurations.TokenExpiredSeconds) * time.Second,
		Printers:              configurations.Printers,
		FontPath:              configurations.FontPath,
		StationFunctionConfig: configurations.StationFunctionConfig,
		MesPath:               configurations.MesPath,
	}
	if err := mcomImpl.RegisterHandlers(dm, api, serviceConfig); err != nil {
		zap.L().Fatal("failed to register handlers", zap.Error(err))
	}

	// Protected data endpoints
	// api.ProtectedGetDataHandler = protected.GetDataHandlerFunc(protectedImpl.GetData)

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {
		role.ClearPermission()
		zap.L().Info("Closing DataManager Services...")
		if err := dm.Close(); err != nil {
			zap.L().Error("server shutdown error..", zap.Error(err))
		}
	}

	return setupGlobalMiddleware(api.Context().BasePath(), dm, api.Serve(setupMiddleware), configurations.CorsAllowedOrigins)
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
	if configurations.CreateUIConfig {
		cfg := uiConfig{
			scheme: scheme,
			addr:   addr,
		}
		if err := buildUIConfig(configurations.UIDir, cfg); err != nil {
			zap.L().Fatal("failed to create UI config", zap.Error(err))
		}
	}
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddleware(handler http.Handler) http.Handler {
	return middleware.ContextTimeoutMiddleware(handler, configurations.Timeout)
}

// checkServerAlive will check if DataManager is registered or not, otherwise will never serve server.
func checkServerAlive(dm mcom.DataManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dm == nil {
			zap.L().Error("non registered server[dm]")
			http.Error(w, server.ErrorNilConnection.Error(), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(apiBasePath string, dm mcom.DataManager, handler http.Handler, corsAllowedOrigins []string) http.Handler {
	handler = middleware.LoggingMiddleware(handler)
	handler = checkServerAlive(dm, handler)
	handler = middleware.MaybeServeUI(apiBasePath, configurations.UIDir, handler)
	handler = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(handler)
	// default without cors handler.
	if len(corsAllowedOrigins) > 0 {
		handler = addCORSOrigins(handler, corsAllowedOrigins)
	}
	return handler
}

// setLogger replaces global logger and redirects STD logger.
func setLogger(devMode bool) {
	var config zap.Config
	if devMode {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.EncoderConfig.EncodeTime = func(tm time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(tm.In(time.Local).Format(time.RFC3339Nano))
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatalln("failed to initialize logger", err)
	}
	zap.ReplaceGlobals(logger)
	zap.RedirectStdLog(logger)
	defer logger.Sync() // nolint
	logger.Info("logger initialized")
}

type uiConfig struct {
	scheme string
	addr   string // host:port
}

// buildUIConfig creates an UI config.
func buildUIConfig(dir string, cfg uiConfig) error {
	f, err := os.Create(filepath.Join(dir, "config.js"))
	if err != nil {
		return err
	}
	defer f.Close()
	defer f.Sync() // nolint

	_, err = f.Write([]byte(fmt.Sprintf(`
window.config = {
  ApiUrl: '%s://%s/api'
}
`, cfg.scheme, cfg.addr)))
	return err
}

// addCORSOrigins enable cross-origin resource sharing.
func addCORSOrigins(handler http.Handler, allowedOrigins []string) http.Handler {
	return cors.New(
		cors.Options{
			AllowedOrigins: allowedOrigins,
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: false,
		},
	).Handler(handler)
}
