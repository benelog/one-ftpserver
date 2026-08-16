package oneftpserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os/signal"
	"strconv"
	"syscall"

	ftpserver "github.com/fclairamb/ftpserverlib"
)

// Server is a listening FTP server. Its zero value is not usable; call New.
type Server struct {
	config *Config
	ftp    *ftpserver.FtpServer
	logger *slog.Logger
}

// New builds a server from a prepared Config, writing what its clients do to
// logger. Nothing is listening yet.
func New(config *Config, logger *slog.Logger) (*Server, error) {
	drv, err := newDriver(config, logger)
	if err != nil {
		return nil, err
	}

	ftp := ftpserver.NewFtpServer(drv)
	// The library drops its own diagnostics by default; they belong in the same
	// log as the activity of the clients.
	ftp.Logger = logger

	return &Server{config: config, ftp: ftp, logger: logger}, nil
}

// Start binds the port and returns what the server ended up with, which is
// the only way to learn the port when --port=0 was used. Serve has to be
// called afterwards to accept clients.
func (s *Server) Start() (*Summary, error) {
	if err := s.ftp.Listen(); err != nil {
		return nil, err
	}

	port, err := s.listeningPort()
	if err != nil {
		return nil, err
	}

	host := s.config.PublicHost
	if host == "" {
		host = localIP()
	}

	return newSummary(s.config, host, port), nil
}

// Serve accepts clients until Stop is called, which it reports as a nil error.
func (s *Server) Serve() error {
	return s.ftp.Serve()
}

// Stop closes the listener, which makes Serve return.
func (s *Server) Stop() error {
	return s.ftp.Stop()
}

// Addr is the address the server listens on, empty before Start.
func (s *Server) Addr() string {
	return s.ftp.Addr()
}

// listeningPort reads the port back from the listener rather than from the
// configuration, because the operating system chooses it when --port=0.
func (s *Server) listeningPort() (int, error) {
	_, port, err := net.SplitHostPort(s.ftp.Addr())
	if err != nil {
		return 0, fmt.Errorf("cannot read the listening address: %w", err)
	}

	return strconv.Atoi(port)
}

// Run starts a server, prints what it is serving to out, and blocks until it
// is interrupted or the timeout of the configuration runs out.
func Run(config *Config, out io.Writer) error {
	logger, logFile, err := newLogger(config)
	if err != nil {
		return err
	}

	defer func() { _ = logFile.Close() }()

	server, err := New(config, logger)
	if err != nil {
		return err
	}

	summary, err := server.Start()
	if err != nil {
		return err
	}

	logger.Info("server started", "address", summary.Address, "home", config.Home)
	defer logger.Info("server stopped")

	printSummary := summary.Print
	if config.JSON {
		printSummary = summary.PrintJSON
	}

	if err := printSummary(out); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if config.Timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	go func() {
		<-ctx.Done()

		if err := server.Stop(); err != nil {
			_, _ = fmt.Fprintf(out, "could not stop cleanly: %v\n", err)
		}
	}()

	if err := server.Serve(); err != nil {
		return err
	}

	if config.Timeout > 0 && errors.Is(context.Cause(ctx), context.DeadlineExceeded) {
		_, _ = fmt.Fprintf(out, "stopped after %s\n", config.Timeout)
	}

	return nil
}
