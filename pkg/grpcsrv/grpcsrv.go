package grpcsrv

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/kwlailab/goutil/pkg/health"
	"github.com/kwlailab/goutil/pkg/log"
	"github.com/kwlailab/goutil/pkg/shutdown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const MAX_MESSAGE_SIZE = 20 * 1024 * 1024

func DefaultDialOptions() []grpc.DialOption {
	credential := grpc.WithTransportCredentials(insecure.NewCredentials())
	msgMaxOptions := grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(MAX_MESSAGE_SIZE),
		grpc.MaxCallSendMsgSize(MAX_MESSAGE_SIZE),
	)
	keepalive := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time: 5 * time.Minute,
	})
	return []grpc.DialOption{credential, msgMaxOptions, keepalive, grpc.WithBlock()}
}

func TLSDialOption() grpc.DialOption {
	return grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
}

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
}

func NewGRPCServer(opt ...grpc.ServerOption) *GRPCServer {
	serverOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 2 * time.Minute,
		}),
		grpc.MaxRecvMsgSize(MAX_MESSAGE_SIZE),
		grpc.MaxSendMsgSize(MAX_MESSAGE_SIZE),
	}
	serverOpts = append(serverOpts, opt...)
	server := grpc.NewServer(serverOpts...)

	grpc_health_v1.RegisterHealthServer(server, health.DefaultHealthCheck())

	return &GRPCServer{
		server:   server,
		listener: nil,
	}
}

func (s *GRPCServer) RegisterService(desc *grpc.ServiceDesc, impl any) {
	s.server.RegisterService(desc, impl)
}

func (s *GRPCServer) ListenAndServe(addr string) {
	// 在所有服務註冊完成後註冊 reflection
	reflection.Register(s.server)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.FatalCtx(context.Background(), "failed to listen", "error", err)
		return
	}
	s.listener = listener
	go func() {
		if err := s.server.Serve(listener); err != nil {
			health.SetHealthStatus(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
			log.FatalCtx(context.Background(), "failed to serve", "error", err)
		}
	}()
	health.SetHealthStatusServing()
	log.Info("server listening on", "addr", addr)
	shutdown.Register(func(ctx context.Context) error {
		health.SetHealthStatus(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		s.server.GracefulStop()
		return nil
	})
}
