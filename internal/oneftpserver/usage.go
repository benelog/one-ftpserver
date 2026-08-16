package oneftpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// Summary is what the server prints once it is listening: the settings it ended
// up with, and ready to run commands for the clients most people reach for. It
// is also the shape of the --json output, so a script can pick up the address
// and the credentials without parsing the human readable text.
type Summary struct {
	Address      string   `json:"address"`
	Protocol     string   `json:"protocol"`
	Host         string   `json:"host"`
	Port         int      `json:"port"`
	ID           string   `json:"id"`
	Password     string   `json:"password"`
	Anonymous    bool     `json:"anonymous"`
	Home         string   `json:"home"`
	PassivePorts string   `json:"passivePorts,omitempty"`
	SSL          bool     `json:"ssl"`
	Cert         string   `json:"cert,omitempty"`
	Timeout      string   `json:"timeout,omitempty"`
	Log          string   `json:"log,omitempty"`
	Upload       string   `json:"upload"`
	Download     string   `json:"download"`
	Get          string   `json:"get,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// newSummary describes a server that is already listening on port, which is not
// necessarily the configured one: --port=0 lets the operating system pick.
func newSummary(config *Config, host string, port int) *Summary {
	summary := &Summary{
		Address:      fmt.Sprintf("%s://%s:%d", config.Scheme(), host, port),
		Protocol:     config.Scheme(),
		Host:         host,
		Port:         port,
		ID:           config.ID,
		Password:     config.Password,
		Anonymous:    config.Anonymous(),
		Home:         config.Home,
		PassivePorts: config.PassivePorts,
		SSL:          config.SSL,
		Cert:         config.Cert,
		Log:          config.LogFile(),
	}

	if config.Timeout > 0 {
		summary.Timeout = config.Timeout.String()
	}

	if config.GeneratedPassword {
		summary.Warnings = append(summary.Warnings,
			"no --password was given, so this one was generated for this run")
	}

	summary.fillCommands(config)

	return summary
}

// fillCommands writes the client commands that match this configuration, with
// the credentials already filled in.
func (s *Summary) fillCommands(config *Config) {
	credentials := ""
	if !config.Anonymous() {
		credentials = " -u " + shellQuote(config.ID+":"+config.Password)
	}

	insecure := ""
	if config.SSL && !config.OwnCertificate() {
		// The generated certificate is signed by nobody, so a client has no
		// way to verify it. A certificate of your own can be verified as is.
		insecure = " -k"
	}

	s.Upload = fmt.Sprintf("curl -T [filename] %s/%s%s", s.Address, credentials, insecure)
	s.Download = fmt.Sprintf("curl -O %s/[filename]%s%s", s.Address, credentials, insecure)

	if !config.SSL {
		s.Get = "wget " + s.Address + "/[filename]"
		if !config.Anonymous() {
			s.Get = fmt.Sprintf("wget --user=%s --password=%s %s/[filename]",
				shellQuote(config.ID), shellQuote(config.Password), s.Address)
		}
	}
}

// Print writes the summary for a person to read.
func (s *Summary) Print(out io.Writer) error {
	writer := &errWriter{out: out}

	writer.printf("FTP server started : %s\n\n", s.Address)

	writer.printf("# Settings\n")
	writer.printf("- ssl : %t\n", s.SSL)

	if s.SSL {
		writer.printf("- cert : %s\n", orGenerated(s.Cert))
	}

	writer.printf("- port : %d\n", s.Port)
	writer.printf("- passivePorts : %s\n", orNone(s.PassivePorts))
	writer.printf("- id : %s\n", s.ID)
	writer.printf("- password : %s\n", orNone(s.Password))
	writer.printf("- home : %s\n", s.Home)
	writer.printf("- timeout : %s\n", orNone(s.Timeout))
	writer.printf("- log : %s\n", orNone(s.Log))

	for _, warning := range s.Warnings {
		writer.printf("\n! %s\n", warning)
	}

	writer.printf("\n# Client commands\n")
	writer.printf("- upload : %s\n", s.Upload)
	writer.printf("- download : %s\n", s.Download)

	if s.Get != "" {
		writer.printf("- download : %s\n", s.Get)
	}

	writer.printf("\nRun with --help to see every setting.\n")

	return writer.err
}

// errWriter keeps the summary readable by holding on to the first write error
// instead of checking every single line.
type errWriter struct {
	out io.Writer
	err error
}

func (w *errWriter) printf(format string, args ...any) {
	if w.err != nil {
		return
	}

	_, w.err = fmt.Fprintf(w.out, format, args...)
}

// PrintJSON writes the summary as one JSON object for a script to read.
func (s *Summary) PrintJSON(out io.Writer) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")

	return encoder.Encode(s)
}

func orNone(value string) string {
	if value == "" {
		return "[none]"
	}

	return value
}

func orGenerated(value string) string {
	if value == "" {
		return "[generated for this run]"
	}

	return value
}

// shellQuote makes a value safe to paste into a shell, which matters because
// a password can hold anything.
func shellQuote(value string) string {
	if isPlainShellWord(value) {
		return value
	}

	return strconv.Quote(value)
}

func isPlainShellWord(value string) bool {
	if value == "" {
		return false
	}

	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z',
			char >= 'A' && char <= 'Z',
			char >= '0' && char <= '9',
			char == '_', char == '-', char == '.', char == '@', char == ':', char == '/':
		default:
			return false
		}
	}

	return true
}
