package http_server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofreego/mediabase/api/mediabase_v1"
	"github.com/gofreego/mediabase/internal/configs"
	"github.com/gofreego/mediabase/internal/service"
	minioStorage "github.com/gofreego/mediabase/internal/storage/minio"

	"github.com/gofreego/goutils/api"
	"github.com/gofreego/goutils/api/debug"

	"github.com/gofreego/goutils/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type HTTPServer struct {
	cfg    *configs.Configuration
	server *http.Server
}

func (a *HTTPServer) Name() string {
	return "HTTP_Server"
}

func (a *HTTPServer) Shutdown(ctx context.Context) {
	if err := a.server.Shutdown(ctx); err != nil {
		logger.Panic(ctx, "failed to shutdown %s : %v", a.Name(), err)
	}
}

func NewHTTPServer(cfg *configs.Configuration) *HTTPServer {
	return &HTTPServer{
		cfg: cfg,
	}
}

func (a *HTTPServer) Run(ctx context.Context) error {

	if a.cfg.Server.HTTPPort == 0 {
		logger.Panic(ctx, "http port is not provided")
	}

	// Initialize MinIO storage
	storage, err := minioStorage.NewMinIOStorage(a.cfg.Storage)
	if err != nil {
		logger.Panic(ctx, "failed to initialize storage: %v", err)
	}

	service := service.NewService(ctx, &a.cfg.Service, storage)

	mux := runtime.NewServeMux()

	api.RegisterSwaggerHandler(ctx, mux, "/mediabase/v1/swagger", "./api/docs/proto", "/mediabase/v1/mediabase.swagger.json")
	err = mediabase_v1.RegisterMediabaseServiceHandlerServer(ctx, mux, service)
	if err != nil {
		logger.Panic(ctx, "failed to register ping service : %v", err)
	}

	// Register debug endpoints if enabled
	if a.cfg.Debug.Enabled {
		debug.RegisterDebugHandlersWithGateway(ctx, &a.cfg.Debug, mux, a.cfg.Logger.AppName, string(a.cfg.Logger.Build), "/mediabase/v1")
	}

	// Serve static test files at /test/ so test.html can make same-origin API calls
	testFileServer := http.StripPrefix("/test/", http.FileServer(http.Dir("./test")))
	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/test/" {
			testFileServer.ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.Server.HTTPPort),
		Handler: logger.WithRequestMiddleware(logger.WithRequestTimeMiddleware(api.CORSMiddleware(rootHandler))),
	}

	logger.Info(ctx, "Starting HTTP server on port %d", a.cfg.Server.HTTPPort)
	logger.Info(ctx, "Swagger UI is available at `http://localhost:%d/mediabase/v1/swagger`", a.cfg.Server.HTTPPort)
	if a.cfg.Debug.Enabled {
		logger.Info(ctx, "Debug dashboard available at `http://localhost:%d/mediabase/v1/debug`", a.cfg.Server.HTTPPort)
	}
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Panic(ctx, "failed to start http server : %v", err)
	}
	return nil
}
