package grpc

import (
	"net"

	"consul-journey/internal/server/grpc/handlers"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	logger *zap.Logger
	cfg    *Config
	server *grpc.Server
}

func New(logger *zap.Logger, cfg *Config) *Server {
	return &Server{
		logger: logger.Named("server.grpc"),
		cfg:    cfg,
		server: grpc.NewServer(),
	}
}

func (s *Server) RegisterHandler(handler handlers.Handler) {
	handler.Register(s.server)
}

func (s *Server) Start() {
	s.logger.Info("starting ...")
	lis, err := net.Listen("tcp", s.cfg.srvAddr())
	if err != nil {
		s.logger.Panic("failed to listen", zap.Error(err))
	}
	go func() {
		if err := s.server.Serve(lis); err != nil {
			s.logger.Error("failed to serve", zap.Error(err))
		}
	}()
	s.logger.Info("started", zap.String("address", s.cfg.srvAddr()))
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
