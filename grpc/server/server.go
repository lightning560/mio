package grpcsvr

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"sync"

	"miopkg/log"
	"miopkg/server"
	"miopkg/util/constant"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// TODO: 增加bufconn测试
// Server ...
type Server struct {
	*grpc.Server
	listener net.Listener
	*Config
}

func newServer(config *Config) (*Server, error) {
	var streamInterceptors = append(
		[]grpc.StreamServerInterceptor{defaultStreamServerInterceptor(config.logger, config)},
		config.streamInterceptors...,
	)

	var unaryInterceptors = append(
		[]grpc.UnaryServerInterceptor{defaultUnaryServerInterceptor(config.logger, config)},
		config.unaryInterceptors...,
	)
	if config.EnableOfficialGrpcLog {
		// grpc框架日志，因为官方grpc日志是单例，所以这里要处理下
		grpclog.SetLoggerV2(zapgrpc.NewLogger(grpcLogBuild().ZapLogger()))
	}

	if config.EnableTLS {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.PrivateFile)
		if err != nil {
			return nil, errors.Wrap(err, "tls.LoadX509KeyPair failed")
		}

		certPool := x509.NewCertPool()
		rootBuf, err := ioutil.ReadFile(config.CaFile)
		if err != nil {
			return nil, errors.Wrap(err, "ioutil.ReadFile failed")
		}
		if !certPool.AppendCertsFromPEM(rootBuf) {
			return nil, errors.New("certPool.AppendCertsFromPEM failed")
		}

		tlsConf := &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{cert},
			ClientCAs:    certPool,
		}

		config.serverOptions = append(config.serverOptions,
			grpc.Creds(credentials.NewTLS(tlsConf)),
		)
	}

	config.serverOptions = append(config.serverOptions,
		grpc.StreamInterceptor(StreamInterceptorChain(streamInterceptors...)),
		grpc.UnaryInterceptor(UnaryInterceptorChain(unaryInterceptors...)),
	)

	newServer := grpc.NewServer(config.serverOptions...)
	// ref:mio
	reflection.Register(newServer)
	healthpb.RegisterHealthServer(newServer, health.NewServer())

	listener, err := net.Listen(config.Network, config.Address())
	if err != nil {
		return nil, errors.Wrap(err, "net.Listen failed")
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port

	return &Server{
		Server:   newServer,
		listener: listener,
		Config:   config,
	}, nil
}

func (s *Server) Healthz() bool {
	conn, err := s.listener.Accept()
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

// Server implements server.Server interface.
func (s *Server) Serve() error {
	err := s.Server.Serve(s.listener)
	return err
}

// Stop implements server.Server interface
// it will terminate echo server immediately
func (s *Server) Stop() error {
	s.Server.Stop()
	return nil
}

// GracefulStop implements server.Server interface
// it will stop echo server gracefully
func (s *Server) GracefulStop(ctx context.Context) error {
	s.Server.GracefulStop()
	return nil
}

// Info returns server info, used by governor and consumer balancer
func (s *Server) Info() *server.ServiceInfo {
	serviceAddress := s.listener.Addr().String()
	if s.Config.ServiceAddress != "" {
		serviceAddress = s.Config.ServiceAddress
	}

	info := server.ApplyOptions(
		server.WithScheme("grpc"),
		server.WithAddress(serviceAddress),
		server.WithKind(constant.ServiceProvider),
	)
	return &info
}

// Build 构建日志
func grpcLogBuild() *log.Logger {
	var (
		once   sync.Once
		logger *log.Logger
	)
	once.Do(func() {
		logger = log.MioLogger.With(log.FieldName("server.grpc"))
	})
	return logger
}
