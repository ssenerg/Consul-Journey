package main

import (
	"os"
	"os/signal"
	"path"
	"syscall"

	"consul-journey/internal/node"
	"consul-journey/internal/server/grpc"
	"consul-journey/internal/server/http"
	"consul-journey/internal/server/http/handlers/dashboard"
	"consul-journey/internal/server/http/handlers/home"
	"consul-journey/internal/utils"
	"consul-journey/internal/utils/logging"

	node_handler "consul-journey/internal/server/grpc/handlers/node"

	"go.uber.org/zap"
)

func main() {
	// -----------------
	// --- Pre-setup ---
	// -----------------

	// load config
	cfg := utils.Must(LoadConfig(
		utils.EnvOr("CONFIG_ENV_PREFIX", "CJS"),
		utils.EnvOr("CONFIG_PATH", path.Join(".", "config")),
		utils.EnvOr("CONFIG_NAME", "config"),
	))

	// setup logger
	logger := utils.Must(logging.New(cfg.Logging))
	defer func() { _ = logger.Sync() }()
	mainLogger := logger.Named("main")

	// ------------------
	// --- Setting up ---
	// ------------------

	// setup main logger
	mainLogger.Info("setting up ...")
	defer mainLogger.Info("shutdown complete")

	// setup node
	mainLogger.Info("setting up node ...")
	node, err := node.New(logger, cfg.Node, cfg.Server.HTTP.Port)
	if err != nil {
		mainLogger.Error("failed to setup node", zap.Error(err))
		return
	}
	mainLogger.Info("node setup complete")

	// setup http server
	mainLogger.Info("setting up http server ...")
	httpServer, err := http.New(logger, cfg.Server.HTTP)
	if err != nil {
		mainLogger.Error("failed to setup http server", zap.Error(err))
		return
	}
	mainLogger.Info("http server setup complete")

	// setup http handlers
	httpServer.RegisterHandler("/", home.New("/dashboard"))
	httpServer.RegisterHandler("/dashboard", dashboard.New(node))

	// setup grpc server
	mainLogger.Info("setting up grpc server ...")
	grpcServer := grpc.New(logger, cfg.Server.GRPC)
	mainLogger.Info("grpc server setup complete")

	// setup grpc handlers
	grpcServer.RegisterHandler(node_handler.New())

	mainLogger.Info("setting up completed")

	// ----------------
	// --- Starting ---
	// ----------------

	mainLogger.Info("starting ...")

	// start node
	if err := node.Start(); err != nil {
		mainLogger.Error("failed to start node", zap.Error(err))
		return
	}
	defer node.Stop()

	// start http server
	httpServer.Start()
	defer httpServer.Stop()

	// start grpc server
	grpcServer.Start()
	defer grpcServer.Stop()

	mainLogger.Info("starting completed")

	// ---------------------
	// --- Shutting down ---
	// ---------------------

	// Keep the handler installed for the rest of the process lifetime: once
	// shutdown starts, any further SIGINT/SIGTERM (e.g. a redundant signal from
	// the process manager, or an impatient second Ctrl+C) must be absorbed by
	// the runtime rather than revert to the default disposition and hard-kill us
	// mid-shutdown. So we deliberately do not signal.Stop or close the channel.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	mainLogger.Info("shutting down ...", zap.String("signal", (<-sigChan).String()))
	_ = logger.Sync() // flush before undesired kill by container runtime
}
