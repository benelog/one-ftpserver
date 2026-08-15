package oneftpserver

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
)

func TestDefaultConfigNeedsNoArgument(t *testing.T) {
	config := &Config{Port: DefaultPort, ID: AnonymousID, Home: t.TempDir()}

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if !config.Anonymous() {
		t.Error("a server started without --id must accept anonymous logins")
	}
	if config.Password != "" {
		t.Errorf("an anonymous server needs no password, got %q", config.Password)
	}
	if config.Scheme() != "ftp" {
		t.Errorf("scheme = %q, want ftp", config.Scheme())
	}
}

func TestHomeBecomesAbsolute(t *testing.T) {
	// given a relative path that exists
	config := &Config{ID: AnonymousID, Home: "."}

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if !filepath.IsAbs(config.Home) {
		t.Errorf("home = %q, want an absolute path", config.Home)
	}
}

func TestMissingHomeIsRejected(t *testing.T) {
	config := &Config{ID: AnonymousID, Home: filepath.Join(t.TempDir(), "absent")}

	if err := config.Prepare(); err == nil {
		t.Error("a home directory that does not exist must be rejected")
	}
}

func TestHomeCannotBeAFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file.txt")
	writeFile(t, file, "content")

	config := &Config{ID: AnonymousID, Home: file}

	if err := config.Prepare(); err == nil {
		t.Error("a file must not be accepted as a home directory")
	}
}

func TestPasswordIsGeneratedForANamedUser(t *testing.T) {
	// given --id without --password
	config := &Config{ID: "benelog", Home: t.TempDir()}

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if config.Password == "" {
		t.Fatal("a named user must not end up without a password")
	}
	if !config.GeneratedPassword {
		t.Error("a generated password must be flagged, so that it can be pointed out")
	}
	if len(config.Password) < 16 {
		t.Errorf("generated password is only %d characters long", len(config.Password))
	}
}

func TestGivenPasswordIsKept(t *testing.T) {
	config := &Config{ID: "benelog", Password: "1234", Home: t.TempDir()}

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if config.Password != "1234" {
		t.Errorf("password = %q, want 1234", config.Password)
	}
	if config.GeneratedPassword {
		t.Error("a password given on the command line must not be reported as generated")
	}
}

func TestPasswordOfAnAnonymousServerIsDropped(t *testing.T) {
	// given a password without an --id, which no client would ever send
	config := &Config{ID: AnonymousID, Password: "1234", Home: t.TempDir()}

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if config.Password != "" {
		t.Errorf("password = %q, want it dropped for an anonymous server", config.Password)
	}
}

func TestPortOutOfRangeIsRejected(t *testing.T) {
	config := &Config{Port: 70000, ID: AnonymousID, Home: t.TempDir()}

	if err := config.Prepare(); err == nil {
		t.Error("a port above 65535 must be rejected")
	}
}

func TestNegativeTimeoutIsRejected(t *testing.T) {
	config := &Config{ID: AnonymousID, Home: t.TempDir(), Timeout: -1}

	if err := config.Prepare(); err == nil {
		t.Error("a negative timeout must be rejected")
	}
}

func TestPassivePortsRange(t *testing.T) {
	getter, err := parsePassivePorts("10125-10199")
	if err != nil {
		t.Fatalf("parsePassivePorts failed: %v", err)
	}

	portRange, ok := getter.(*ftpserver.PortRange)
	if !ok {
		t.Fatalf("got %T, want a *ftpserver.PortRange", getter)
	}
	if portRange.Start != 10125 || portRange.End != 10199 {
		t.Errorf("range = %d-%d, want 10125-10199", portRange.Start, portRange.End)
	}
}

func TestASinglePassivePortIsARangeOfOne(t *testing.T) {
	getter, err := parsePassivePorts("10125")
	if err != nil {
		t.Fatalf("parsePassivePorts failed: %v", err)
	}

	portRange, ok := getter.(*ftpserver.PortRange)
	if !ok {
		t.Fatalf("got %T, want a *ftpserver.PortRange", getter)
	}
	if portRange.Start != 10125 || portRange.End != 10125 {
		t.Errorf("range = %d-%d, want 10125-10125", portRange.Start, portRange.End)
	}
}

func TestNoPassivePortsLeavesTheChoiceToTheOS(t *testing.T) {
	getter, err := parsePassivePorts("")
	if err != nil {
		t.Fatalf("parsePassivePorts failed: %v", err)
	}

	if getter != nil {
		t.Errorf("got %v, want no range at all", getter)
	}
}

func TestBadPassivePortsAreRejected(t *testing.T) {
	for _, value := range []string{"abc", "10199-10125", "0-10", "10-70000", "10125-"} {
		if _, err := parsePassivePorts(value); err == nil {
			t.Errorf("passivePorts %q must be rejected", value)
		}
	}
}

func TestPassivePortsAreCheckedBeforeTheServerStarts(t *testing.T) {
	config := &Config{ID: AnonymousID, Home: t.TempDir(), PassivePorts: "nonsense"}

	err := config.Prepare()
	if err == nil {
		t.Fatal("a passive port range that cannot be parsed must be rejected")
	}
	if !strings.Contains(err.Error(), "passivePorts") {
		t.Errorf("error = %q, want it to name the offending flag", err)
	}
}

func TestSSLChangesTheScheme(t *testing.T) {
	config := &Config{ID: AnonymousID, Home: t.TempDir(), SSL: true}

	if config.Scheme() != "ftps" {
		t.Errorf("scheme = %q, want ftps", config.Scheme())
	}
}

func TestCertAndKeyMustComeTogether(t *testing.T) {
	certFile, keyFile := writeTestCertificate(t)

	for _, config := range []*Config{
		{ID: AnonymousID, Home: t.TempDir(), Cert: certFile},
		{ID: AnonymousID, Home: t.TempDir(), Key: keyFile},
	} {
		if err := config.Prepare(); err == nil {
			t.Errorf("cert=%q key=%q must be rejected: one half of the pair is missing", config.Cert, config.Key)
		}
	}
}

func TestOwnCertificateTurnsSSLOn(t *testing.T) {
	certFile, keyFile := writeTestCertificate(t)
	config := &Config{ID: AnonymousID, Home: t.TempDir(), Cert: certFile, Key: keyFile}

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if !config.SSL {
		t.Error("giving a certificate should imply --ssl")
	}
	if !config.OwnCertificate() {
		t.Error("a given certificate must be reported as the server's own")
	}
}

func TestBadCertificateIsRejectedBeforeTheServerStarts(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	writeFile(t, certFile, "not a certificate")
	writeFile(t, keyFile, "not a key")

	config := &Config{ID: AnonymousID, Home: t.TempDir(), Cert: certFile, Key: keyFile}

	if err := config.Prepare(); err == nil {
		t.Error("a certificate that cannot be loaded must be rejected at startup")
	}
}

// writeTestCertificate puts a self-signed certificate and its key into files,
// the shape a real one would arrive in.
func writeTestCertificate(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("cannot generate a test key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-certificate"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("cannot generate a test certificate: %v", err)
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("cannot marshal the test key: %v", err)
	}

	dir := t.TempDir()
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	writePEM(t, certFile, "CERTIFICATE", der)
	writePEM(t, keyFile, "PRIVATE KEY", keyDER)

	return certFile, keyFile
}

func writePEM(t *testing.T, path, blockType string, der []byte) {
	t.Helper()

	var out bytes.Buffer
	if err := pem.Encode(&out, &pem.Block{Type: blockType, Bytes: der}); err != nil {
		t.Fatalf("cannot encode %s: %v", blockType, err)
	}

	writeFile(t, path, out.String())
}
