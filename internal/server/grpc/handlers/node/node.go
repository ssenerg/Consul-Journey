package node

import (
	"context"

	"consul-journey/internal/server/grpc/handlers"
	"consul-journey/proto/gen/go/node"

	"google.golang.org/grpc"
)

type Handler struct {
	node.UnimplementedServiceServer
}

var (
	_ node.ServiceServer = (*Handler)(nil)
	_ handlers.Handler   = (*Handler)(nil)
)

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Ping(ctx context.Context, req *node.PingRequest) (*node.PingResponse, error) {
	return nil, nil
}

func (h *Handler) Register(s grpc.ServiceRegistrar) {
	node.RegisterServiceServer(s, h)
}
