package oneftpserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnonymousLoginTakesAnyPassword(t *testing.T) {
	drv := newTestDriver(t, &Config{ID: AnonymousID, Home: t.TempDir()})

	for _, user := range []string{AnonymousID, "ftp"} {
		if _, err := drv.AuthUser(nil, user, "someone@example.com"); err != nil {
			t.Errorf("%q should be able to log in anonymously: %v", user, err)
		}
	}
}

func TestAnonymousLoginIsRefusedForAnotherUser(t *testing.T) {
	drv := newTestDriver(t, &Config{ID: AnonymousID, Home: t.TempDir()})

	if _, err := drv.AuthUser(nil, "benelog", ""); err == nil {
		t.Error("an anonymous server must only accept the anonymous user")
	}
}

func TestNamedUserNeedsTheRightPassword(t *testing.T) {
	drv := newTestDriver(t, &Config{ID: "benelog", Password: "1234", Home: t.TempDir()})

	if _, err := drv.AuthUser(nil, "benelog", "1234"); err != nil {
		t.Errorf("the configured credentials should be accepted: %v", err)
	}
	if _, err := drv.AuthUser(nil, "benelog", "wrong"); err == nil {
		t.Error("a wrong password must be refused")
	}
	if _, err := drv.AuthUser(nil, "someone", "1234"); err == nil {
		t.Error("a wrong user must be refused")
	}
	if _, err := drv.AuthUser(nil, AnonymousID, ""); err == nil {
		t.Error("anonymous must be refused once --id is given")
	}
}

func TestClientIsConfinedToTheHomeDirectory(t *testing.T) {
	// given a file next to the home directory, which must stay out of reach
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.Mkdir(home, 0o750); err != nil {
		t.Fatalf("prepare home: %v", err)
	}
	writeFile(t, filepath.Join(root, "secret.txt"), "secret")
	writeFile(t, filepath.Join(home, "shared.txt"), "shared")

	drv := newTestDriver(t, &Config{ID: AnonymousID, Home: home})

	fs, err := drv.AuthUser(nil, AnonymousID, "")
	if err != nil {
		t.Fatalf("AuthUser failed: %v", err)
	}

	if _, err := fs.Stat("/shared.txt"); err != nil {
		t.Errorf("a file of the home directory should be reachable: %v", err)
	}
	if _, err := fs.Stat("/../secret.txt"); err == nil {
		t.Error("a file above the home directory must not be reachable")
	}
}

func TestErrorsDoNotRevealTheHomeDirectory(t *testing.T) {
	home := t.TempDir()
	drv := newTestDriver(t, &Config{ID: AnonymousID, Home: home})

	fs, err := drv.AuthUser(nil, AnonymousID, "")
	if err != nil {
		t.Fatalf("AuthUser failed: %v", err)
	}

	// Stat of a missing file fails with the path in the error, and so does a
	// rename, whose error carries two paths.
	_, statErr := fs.Stat("/missing.txt")
	if statErr == nil {
		t.Fatal("Stat of a missing file should fail")
	}
	renameErr := fs.Rename("/missing.txt", "/elsewhere.txt")
	if renameErr == nil {
		t.Fatal("renaming a missing file should fail")
	}

	for _, err := range []error{statErr, renameErr} {
		if strings.Contains(err.Error(), home) {
			t.Errorf("the error %q tells the client where the home directory lives", err)
		}
	}
	if !strings.Contains(statErr.Error(), "/missing.txt") {
		t.Errorf("the error %q should name the file by the path the client used", statErr)
	}
	if !os.IsNotExist(statErr) {
		t.Errorf("the cleaned error should still count as not-exist, got %v", statErr)
	}
}

func TestTLSIsOnlyAvailableWithTheSSLFlag(t *testing.T) {
	plain := newTestDriver(t, &Config{ID: AnonymousID, Home: t.TempDir()})

	if _, err := plain.GetTLSConfig(); err == nil {
		t.Error("a server started without --ssl has no certificate to offer")
	}

	secure := newTestDriver(t, &Config{ID: AnonymousID, Home: t.TempDir(), SSL: true})

	tlsConfig, err := secure.GetTLSConfig()
	if err != nil {
		t.Fatalf("GetTLSConfig failed: %v", err)
	}
	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("got %d certificates, want the one generated for this run", len(tlsConfig.Certificates))
	}
}

func TestSettingsCarryTheConfiguredPorts(t *testing.T) {
	drv := newTestDriver(t, &Config{ID: AnonymousID, Home: t.TempDir(), Port: 10021, PassivePorts: "10125-10199"})

	settings, err := drv.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}

	if settings.ListenAddr != ":10021" {
		t.Errorf("listen address = %q, want :10021", settings.ListenAddr)
	}
	if settings.PassiveTransferPortRange == nil {
		t.Error("the passive port range of the command line must reach the settings")
	}
}

func newTestDriver(t *testing.T, config *Config) *driver {
	t.Helper()

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	drv, err := newDriver(config)
	if err != nil {
		t.Fatalf("newDriver failed: %v", err)
	}

	return drv
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("cannot write %s: %v", path, err)
	}
}
