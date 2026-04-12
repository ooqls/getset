package app

import "google.golang.org/grpc"

const (
	grpc_portOpt   string = "opt-grpc-port"
	grpc_serverOpt string = "opt-grpc-server"
)

type grpcOpt struct{ featureOpt }

func WithGrpcPort(port int) grpcOpt {
	return grpcOpt{featureOpt{key: grpc_portOpt, value: port}}
}

func WithGrpcServer(s *grpc.Server) grpcOpt {
	return grpcOpt{featureOpt{key: grpc_serverOpt, value: s}}
}

type GrpcFeature struct {
	Enabled bool
	Port    int
	Server  *grpc.Server
}

func (f *GrpcFeature) apply(opt grpcOpt) {
	switch opt.key {
	case grpc_portOpt:
		f.Port = opt.value.(int)
	case grpc_serverOpt:
		f.Server = opt.value.(*grpc.Server)
	}
}

func GRPC(opts ...grpcOpt) GrpcFeature {
	f := GrpcFeature{
		Enabled: true,
		Port:    9090,
		Server:  grpc.NewServer(),
	}
	for _, opt := range opts {
		f.apply(opt)
	}
	return f
}
