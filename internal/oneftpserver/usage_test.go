package oneftpserver

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSummaryOfAnAnonymousServerNeedsNoCredentials(t *testing.T) {
	config := prepared(t, &Config{ID: AnonymousID, Home: t.TempDir()})

	summary := newSummary(config, "192.168.0.10", 2121)

	if summary.Address != "ftp://192.168.0.10:2121" {
		t.Errorf("address = %q", summary.Address)
	}
	if strings.Contains(summary.Upload, "-u ") {
		t.Errorf("upload command should carry no credentials: %s", summary.Upload)
	}
	if summary.Get == "" {
		t.Error("a plain FTP server should also show a wget command")
	}
}

func TestSummaryFillsInTheCredentials(t *testing.T) {
	config := prepared(t, &Config{ID: "benelog", Password: "1234", Home: t.TempDir()})

	summary := newSummary(config, "192.168.0.10", 2121)

	if !strings.Contains(summary.Upload, "-u benelog:1234") {
		t.Errorf("upload command should be ready to run: %s", summary.Upload)
	}
	if !strings.Contains(summary.Get, "--user=benelog") {
		t.Errorf("wget command should be ready to run: %s", summary.Get)
	}
}

func TestSummaryQuotesAPasswordThatNeedsIt(t *testing.T) {
	config := prepared(t, &Config{ID: "benelog", Password: "pass word;rm", Home: t.TempDir()})

	summary := newSummary(config, "192.168.0.10", 2121)

	if !strings.Contains(summary.Upload, `"benelog:pass word;rm"`) {
		t.Errorf("a password with shell characters must be quoted: %s", summary.Upload)
	}
}

func TestSummaryOfATLSServerSkipsWgetAndVerification(t *testing.T) {
	config := prepared(t, &Config{ID: AnonymousID, Home: t.TempDir(), SSL: true})

	summary := newSummary(config, "192.168.0.10", 2121)

	if !strings.HasPrefix(summary.Address, "ftps://") {
		t.Errorf("address = %q, want an ftps one", summary.Address)
	}
	if !strings.HasSuffix(summary.Upload, " -k") {
		t.Errorf("a generated certificate cannot be verified, so -k is needed: %s", summary.Upload)
	}
	if summary.Get != "" {
		t.Errorf("wget has no implicit FTPS support, so it should not be suggested: %s", summary.Get)
	}
}

func TestSummaryOfAnOwnCertificateDropsTheInsecureFlag(t *testing.T) {
	certFile, keyFile := writeTestCertificate(t)
	config := prepared(t, &Config{ID: AnonymousID, Home: t.TempDir(), Cert: certFile, Key: keyFile})

	summary := newSummary(config, "192.168.0.10", 2121)

	if !strings.HasPrefix(summary.Address, "ftps://") {
		t.Errorf("address = %q, want an ftps one: a certificate implies --ssl", summary.Address)
	}
	if strings.Contains(summary.Upload, "-k") {
		t.Errorf("a certificate of your own can be verified, so -k must go: %s", summary.Upload)
	}
}

func TestSummaryPointsOutAGeneratedPassword(t *testing.T) {
	config := prepared(t, &Config{ID: "benelog", Home: t.TempDir()})

	summary := newSummary(config, "192.168.0.10", 2121)

	if len(summary.Warnings) != 1 {
		t.Fatalf("warnings = %v, want the generated password to be pointed out", summary.Warnings)
	}

	var out bytes.Buffer
	if err := summary.Print(&out); err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	if !strings.Contains(out.String(), summary.Warnings[0]) {
		t.Error("the warning should show up in the printed summary")
	}
}

func TestPrintedSummaryHoldsEverySetting(t *testing.T) {
	config := prepared(t, &Config{
		ID: "benelog", Password: "1234", Home: t.TempDir(),
		PassivePorts: "10125-10199", Timeout: 30 * time.Minute,
	})

	var out bytes.Buffer
	if err := newSummary(config, "192.168.0.10", 10021).Print(&out); err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	printed := out.String()
	for _, want := range []string{"10021", "benelog", "1234", "10125-10199", "30m0s", config.Home} {
		if !strings.Contains(printed, want) {
			t.Errorf("printed summary is missing %q:\n%s", want, printed)
		}
	}
}

func TestJSONSummaryCanBeParsed(t *testing.T) {
	config := prepared(t, &Config{ID: "benelog", Password: "1234", Home: t.TempDir()})

	var out bytes.Buffer
	if err := newSummary(config, "192.168.0.10", 10021).PrintJSON(&out); err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}

	var parsed Summary
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("the --json output must be valid JSON: %v", err)
	}

	if parsed.Port != 10021 {
		t.Errorf("port = %d, want 10021", parsed.Port)
	}
	if parsed.Password != "1234" {
		t.Errorf("password = %q, want 1234", parsed.Password)
	}
	if parsed.Anonymous {
		t.Error("a named user must not be reported as anonymous")
	}
}

func prepared(t *testing.T, config *Config) *Config {
	t.Helper()

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	return config
}
