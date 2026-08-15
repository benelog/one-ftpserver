// Package oneftpserver runs a single user FTP server that is configured
// entirely by command line arguments. Nothing is read from, and nothing is
// written to, a configuration file.
package oneftpserver

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
)

// Defaults of the settings that have one.
const (
	DefaultPort = 2121
	AnonymousID = "anonymous"
)

// Config is the whole configuration of the server. Every field comes from a
// command line flag, so a Config is all there is to know about a running
// instance.
type Config struct {
	Port         int
	PassivePorts string
	ID           string
	Password     string
	Home         string
	SSL          bool
	Cert         string
	Key          string
	Timeout      time.Duration
	PublicHost   string
	JSON         bool

	// GeneratedPassword records that Password was not given on the command
	// line and had to be generated, which the usage output points out.
	GeneratedPassword bool
}

// Anonymous reports whether the server accepts anonymous logins, which is the
// case as long as no --id was given.
func (c *Config) Anonymous() bool {
	return c.ID == AnonymousID
}

// OwnCertificate reports whether the server was handed a certificate, instead
// of generating one for the run.
func (c *Config) OwnCertificate() bool {
	return c.Cert != ""
}

// Scheme is the URL scheme clients have to use.
func (c *Config) Scheme() string {
	if c.SSL {
		return "ftps"
	}

	return "ftp"
}

// Prepare fills in the values that could not be defaulted by the flag package
// and rejects the combinations that cannot work. It is called once, before the
// server is started.
func (c *Config) Prepare() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535, got %d", c.Port)
	}

	if c.Timeout < 0 {
		return fmt.Errorf("timeout must not be negative, got %s", c.Timeout)
	}

	if err := c.resolveHome(); err != nil {
		return err
	}

	if _, err := parsePassivePorts(c.PassivePorts); err != nil {
		return err
	}

	if err := c.resolveTLS(); err != nil {
		return err
	}

	return c.resolvePassword()
}

// resolveTLS checks the certificate pair before the server starts. Giving one
// turns --ssl on by itself: nobody hands over a certificate to then serve
// plain FTP.
func (c *Config) resolveTLS() error {
	if (c.Cert == "") != (c.Key == "") {
		return errors.New("cert and key must be given together")
	}

	if !c.OwnCertificate() {
		return nil
	}

	if _, err := tls.LoadX509KeyPair(c.Cert, c.Key); err != nil {
		return fmt.Errorf("cannot load the certificate: %w", err)
	}

	c.SSL = true

	return nil
}

// resolveHome turns the home directory into an absolute path and makes sure it
// can actually serve files.
func (c *Config) resolveHome() error {
	if c.Home == "" {
		c.Home = "."
	}

	home, err := filepath.Abs(c.Home)
	if err != nil {
		return fmt.Errorf("cannot resolve home directory %q: %w", c.Home, err)
	}

	info, err := os.Stat(home)
	if err != nil {
		return fmt.Errorf("cannot use home directory %q: %w", home, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("home %q is not a directory", home)
	}

	c.Home = home

	return nil
}

// resolvePassword generates a password when a named user was asked for without
// one. Refusing to start would be the safe move, but it would also force a
// second attempt for something the server can decide on its own; the generated
// password is printed with the rest of the settings.
func (c *Config) resolvePassword() error {
	if c.Anonymous() {
		c.Password = ""
		c.GeneratedPassword = false

		return nil
	}

	if c.Password != "" {
		return nil
	}

	c.Password = rand.Text()
	c.GeneratedPassword = true

	return nil
}

// parsePassivePorts reads a "10125-10199" range, or a single port. An empty
// value leaves the choice of the passive ports to the operating system.
func parsePassivePorts(value string) (ftpserver.PasvPortGetter, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil //nolint:nilnil // no range asked for is not an error
	}

	start, end, found := strings.Cut(value, "-")
	if !found {
		end = start
	}

	first, err := portNumber(start)
	if err != nil {
		return nil, err
	}

	last, err := portNumber(end)
	if err != nil {
		return nil, err
	}

	if first > last {
		return nil, fmt.Errorf("passivePorts %q starts after it ends", value)
	}

	return &ftpserver.PortRange{Start: first, End: last}, nil
}

func portNumber(value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("passivePorts must be a port or a 'start-end' range, got %q", value)
	}

	if port < 1 || port > 65535 {
		return 0, errors.New("passivePorts must be between 1 and 65535, got " + value)
	}

	return port, nil
}
