package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const shutdownTimeout = 10 * time.Second

// listen binds port, wrapping the listener in TLS when the TLS feature is
// enabled. protos is the ALPN protocol list offered to clients.
func listen(fs *Features, port int, protos ...string) (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	if fs.TLS.Enabled {
		cfg := fs.TLS.ServerConfig()
		cfg.NextProtos = protos
		lis = tls.NewListener(lis, cfg)
	}
	return lis, nil
}

// serveHTTP serves h on port until ctx is cancelled, then shuts it down
// gracefully. It returns the bound address.
func serveHTTP(ctx *AppContext, fs *Features, name string, port int, h http.Handler) (net.Addr, error) {
	l := ctx.L().With(zap.String("server", name))
	lis, err := listen(fs, port, "http/1.1")
	if err != nil {
		return nil, err
	}

	srv := &http.Server{Handler: h}
	ctx.spawn(func() {
		l.Debug("[Serve] listening", zap.Stringer("addr", lis.Addr()), zap.Bool("tls", fs.TLS.Enabled))
		if err := srv.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("[Serve] server stopped with an error", zap.Error(err))
		}
	})
	ctx.spawn(func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			l.Error("[Serve] failed to shut down", zap.Error(err))
		}
		l.Info("[Serve] stopped")
	})

	return lis.Addr(), nil
}

// serveGrpc is serveHTTP for a gRPC server.
func serveGrpc(ctx *AppContext, fs *Features, port int, srv *grpc.Server) (net.Addr, error) {
	l := ctx.L().With(zap.String("server", "grpc"))
	lis, err := listen(fs, port, "h2")
	if err != nil {
		return nil, err
	}

	ctx.spawn(func() {
		l.Debug("[Serve] listening", zap.Stringer("addr", lis.Addr()), zap.Bool("tls", fs.TLS.Enabled))
		if err := srv.Serve(lis); err != nil {
			l.Error("[Serve] server stopped with an error", zap.Error(err))
		}
	})
	ctx.spawn(func() {
		<-ctx.Done()
		srv.GracefulStop()
		l.Info("[Serve] stopped")
	})

	return lis.Addr(), nil
}
