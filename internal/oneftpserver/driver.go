package oneftpserver

import (
	"crypto/subtle"
	"crypto/tls"
	"errors"
	"log/slog"
	"strconv"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

// errAuthFailed is what every rejected login gets, whatever the reason, so a
// caller cannot tell a wrong user from a wrong password.
var errAuthFailed = errors.New("authentication failed")

// driver serves the single user described by Config out of a single directory.
type driver struct {
	config    *Config
	settings  *ftpserver.Settings
	tlsConfig *tls.Config
	logger    *slog.Logger
}

// newDriver builds the driver, and with it the TLS certificate when the server
// is asked to run over TLS.
func newDriver(config *Config, logger *slog.Logger) (*driver, error) {
	passivePorts, err := parsePassivePorts(config.PassivePorts)
	if err != nil {
		return nil, err
	}

	settings := &ftpserver.Settings{
		ListenAddr:               ":" + strconv.Itoa(config.Port),
		PublicHost:               config.PublicHost,
		PassiveTransferPortRange: passivePorts,
		Banner:                   "one-ftpserver",
	}

	drv := &driver{config: config, settings: settings, logger: logger}

	if config.SSL {
		// Implicit TLS keeps the ftps:// URL of the usage output working as
		// is, which is how this server was used before it moved to Go.
		settings.TLSRequired = ftpserver.ImplicitEncryption

		if config.OwnCertificate() {
			drv.tlsConfig, err = loadTLSConfig(config.Cert, config.Key)
		} else {
			drv.tlsConfig, err = selfSignedTLSConfig()
		}

		if err != nil {
			return nil, err
		}
	}

	return drv, nil
}

// GetSettings returns the settings built from the command line.
func (d *driver) GetSettings() (*ftpserver.Settings, error) {
	return d.settings, nil
}

// ClientConnected returns the message a client sees before it logs in.
func (d *driver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	d.clientLogger(cc).Info("client connected")

	return "one-ftpserver", nil
}

// ClientDisconnected has nothing to clean up: no state is kept per client.
func (d *driver) ClientDisconnected(cc ftpserver.ClientContext) {
	d.clientLogger(cc).Info("client disconnected")
}

// AuthUser checks the credentials and hands back a file system rooted at the
// home directory, so a client cannot reach anything above it.
func (d *driver) AuthUser(cc ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	logger := d.clientLogger(cc).With("id", user)

	if !d.authenticate(user, pass) {
		// The password is never logged, not even a rejected one: a mistyped
		// password is usually a real one with a character out of place.
		logger.Warn("login refused")

		return nil, errAuthFailed
	}

	logger.Info("login")

	// BasePathFs names files by their full path on disk in its errors, and
	// those errors are repeated to the client, so they go through a wrapper
	// that puts the path the client used back in.
	fs := afero.NewBasePathFs(afero.NewOsFs(), d.config.Home)
	hiding := &pathHidingFs{Fs: fs, home: d.config.Home}

	return &loggingFs{Fs: hiding, logger: logger}, nil
}

// clientLogger tags every line of a client with who it is, which is what makes
// the log readable when several clients are connected at once.
func (d *driver) clientLogger(cc ftpserver.ClientContext) *slog.Logger {
	return d.logger.With("client", cc.ID(), "from", cc.RemoteAddr().String())
}

// authenticate reports whether these credentials are the ones the server was
// started with.
func (d *driver) authenticate(user, pass string) bool {
	if d.config.Anonymous() {
		// "ftp" is the other name clients traditionally send for an anonymous
		// login, and the password is an e-mail address nobody checks.
		return user == AnonymousID || user == "ftp"
	}

	// Constant time, so a wrong password cannot be found one character at a
	// time by measuring how long the answer takes.
	sameUser := subtle.ConstantTimeCompare([]byte(user), []byte(d.config.ID)) == 1
	samePassword := subtle.ConstantTimeCompare([]byte(pass), []byte(d.config.Password)) == 1

	return sameUser && samePassword
}

// GetTLSConfig returns the generated certificate, and fails when the server was
// not started with --ssl.
func (d *driver) GetTLSConfig() (*tls.Config, error) {
	if d.tlsConfig == nil {
		return nil, errors.New("TLS is not enabled, start the server with --ssl")
	}

	return d.tlsConfig, nil
}
