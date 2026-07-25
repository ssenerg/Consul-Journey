package handlers

import "google.golang.org/grpc"

type Handler interface {
	Register(s grpc.ServiceRegistrar)
}
