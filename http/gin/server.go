package gin

import (
	"context"
	"net/http"

	"net"

	"miopkg/errors"
	"miopkg/log"
	"miopkg/server"
	"miopkg/util/constant"

	"github.com/gin-gonic/gin"
)

// Server ...
type Server struct {
	*gin.Engine
	Server   *http.Server
	config   *Config
	listener net.Listener
}

func newServer(config *Config) *Server {
	listener, err := net.Listen("tcp", config.Address())
	if err != nil {
		config.logger.Panic("new gin server err", log.FieldErrKind(errors.ErrKindListenErr), log.FieldErr(err))
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port
	gin.SetMode(config.Mode)
	return &Server{
		Engine:   gin.New(),
		config:   config,
		listener: listener,
	}
}

// Upgrade protocol to WebSocket
func (s *Server) Upgrade(ws *WebSocket) gin.IRoutes {
	return s.GET(ws.Pattern, func(c *gin.Context) {
		ws.Upgrade(c.Writer, c.Request)
	})
}

// Serve implements server.Server interface.
func (s *Server) Serve() error {
	// s.Gin.StdLogger = log.MioLogger.StdLog()
	for _, route := range s.Engine.Routes() {
		s.config.logger.Info("add route", log.FieldMethod(route.Method), log.String("path", route.Path))
	}
	s.Server = &http.Server{
		Addr:    s.config.Address(),
		Handler: s,
	}
	err := s.Server.Serve(s.listener)
	if err == http.ErrServerClosed {
		s.config.logger.Info("close gin", log.FieldAddr(s.config.Address()))
		return nil
	}

	return err
}

// Stop implements server.Server interface
// it will terminate gin server immediately
func (s *Server) Stop() error {
	return s.Server.Close()
}

// GracefulStop implements server.Server interface
// it will stop gin server gracefully
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

// Info returns server info, used by governor and consumer balancer
func (s *Server) Info() *server.ServiceInfo {
	serviceAddr := s.listener.Addr().String()
	if s.config.ServiceAddress != "" {
		serviceAddr = s.config.ServiceAddress
	}

	info := server.ApplyOptions(
		server.WithScheme("http"),
		server.WithAddress(serviceAddr),
		server.WithKind(constant.ServiceProvider),
	)
	// info.Name = info.Name + "." + ModName
	return &info
}

func (s *Server) Healthz() bool {
	if s.listener == nil {
		return false
	}

	conn, err := s.listener.Accept()
	if err != nil {
		return false
	}

	conn.Close()
	return true
}
