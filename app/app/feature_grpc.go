package app

import (
	"net"

	"google.golang.org/grpc"
)

type GrpcConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type GrpcOption func(*GrpcFeature)

func WithGrpcPort(port int) GrpcOption {
	return func(f *GrpcFeature) { f.Port = port }
}

func WithGrpcServer(s *grpc.Server) GrpcOption {
	return func(f *GrpcFeature) { f.Server = s }
}

// GrpcFeature serves Server. When the TLS feature is enabled, connections
// are TLS-terminated before reaching Server, so Server should be created
// without transport credentials.
type GrpcFeature struct {
	Enabled bool
	Port    int
	Server  *grpc.Server
	addr    net.Addr
}

func GRPC(opts ...GrpcOption) GrpcFeature {
	f := GrpcFeature{
		Enabled: true,
		Port:    9090,
		Server:  grpc.NewServer(),
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func grpcFromConfig(c GrpcConfig) GrpcFeature {
	return GrpcFeature{
		Enabled: c.Enabled,
		Port:    c.Port,
		Server:  grpc.NewServer(),
	}
}

func (f *GrpcFeature) name() string  { return "grpc" }
func (f *GrpcFeature) enabled() bool { return f.Enabled }

// Addr is the address the gRPC server is bound to, or nil before it serves.
func (f *GrpcFeature) Addr() net.Addr { return f.addr }

func (f *GrpcFeature) start(ctx *AppContext, fs *Features) error {
	if f.Server == nil {
		f.Server = grpc.NewServer()
	}
	return nil
}

func (f *GrpcFeature) serve(ctx *AppContext, fs *Features) error {
	addr, err := serveGrpc(ctx, fs, f.Port, f.Server)
	f.addr = addr
	return err
}
