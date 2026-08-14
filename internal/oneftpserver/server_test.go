package oneftpserver

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAFileCanBeDownloaded(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, "shared.txt"), "my works")

	summary := startServer(t, &Config{ID: AnonymousID, Home: home})

	client := dial(t, summary.Port, false)
	client.login(AnonymousID, "someone@example.com")

	if got := client.retrieve("shared.txt"); got != "my works" {
		t.Errorf("downloaded %q, want %q", got, "my works")
	}
}

func TestAFileCanBeUploaded(t *testing.T) {
	home := t.TempDir()
	summary := startServer(t, &Config{ID: "benelog", Password: "1234", Home: home})

	client := dial(t, summary.Port, false)
	client.login("benelog", "1234")
	client.store("uploaded.txt", "my works")

	stored, err := os.ReadFile(filepath.Join(home, "uploaded.txt"))
	if err != nil {
		t.Fatalf("the uploaded file is not on disk: %v", err)
	}
	if string(stored) != "my works" {
		t.Errorf("stored %q, want %q", stored, "my works")
	}
}

func TestAWrongPasswordIsRefused(t *testing.T) {
	summary := startServer(t, &Config{ID: "benelog", Password: "1234", Home: t.TempDir()})

	client := dial(t, summary.Port, false)
	client.send("USER benelog")
	client.expect(331)

	client.send("PASS wrong")
	if code, line := client.reply(); code != 530 {
		t.Errorf("PASS with a wrong password answered %d %q, want 530", code, line)
	}
}

func TestTheServerCanRunOverTLS(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, "shared.txt"), "my works")

	summary := startServer(t, &Config{ID: AnonymousID, Home: home, SSL: true})

	// The certificate is generated for this run and signed by nobody, which is
	// why the printed commands carry curl's -k.
	client := dial(t, summary.Port, true)
	client.login(AnonymousID, "someone@example.com")

	client.send("PWD")
	client.expect(257)
}

func TestThePortIsChosenByTheOSWhenAskedTo(t *testing.T) {
	summary := startServer(t, &Config{Port: 0, ID: AnonymousID, Home: t.TempDir()})

	if summary.Port == 0 {
		t.Error("the summary must report the port the OS picked, not the requested 0")
	}
}

func TestTheServerStopsAfterItsTimeout(t *testing.T) {
	config := prepared(t, &Config{Port: 0, ID: AnonymousID, Home: t.TempDir(), Timeout: 200 * time.Millisecond})

	var out bytes.Buffer

	done := make(chan error, 1)
	go func() { done <- Run(config, &out) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the server ignored its --timeout")
	}

	if !strings.Contains(out.String(), "stopped after") {
		t.Errorf("the reason for stopping should be printed:\n%s", out.String())
	}
}

// startServer runs a server on a port of the operating system's choosing and
// stops it when the test ends.
func startServer(t *testing.T, config *Config) *Summary {
	t.Helper()

	config.Port = 0
	prepared(t, config)

	server, err := New(config)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	summary, err := server.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	go func() {
		if err := server.Serve(); err != nil {
			t.Errorf("Serve failed: %v", err)
		}
	}()

	t.Cleanup(func() {
		if err := server.Stop(); err != nil {
			t.Errorf("Stop failed: %v", err)
		}
	})

	return summary
}

// ftpClient is the little bit of RFC 959 the tests need: enough to log in and
// move one file, over a plain or a TLS connection.
type ftpClient struct {
	t      *testing.T
	port   int
	tls    bool
	conn   net.Conn
	reader *bufio.Reader
}

func dial(t *testing.T, port int, useTLS bool) *ftpClient {
	t.Helper()

	client := &ftpClient{t: t, port: port, tls: useTLS}
	client.conn = client.connect()
	client.reader = bufio.NewReader(client.conn)

	t.Cleanup(func() { _ = client.conn.Close() })

	client.expect(220)

	return client
}

// connect opens a connection to the control port, or to a passive port when
// the address of one is given by the caller through dialData.
func (c *ftpClient) connect() net.Conn {
	c.t.Helper()

	return c.dialAddr(net.JoinHostPort("127.0.0.1", strconv.Itoa(c.port)))
}

func (c *ftpClient) dialAddr(addr string) net.Conn {
	c.t.Helper()

	if c.tls {
		// The server generates its certificate on startup, so there is nothing
		// to verify it against.
		conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // test client
		if err != nil {
			c.t.Fatalf("cannot connect to %s: %v", addr, err)
		}

		return conn
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		c.t.Fatalf("cannot connect to %s: %v", addr, err)
	}

	return conn
}

func (c *ftpClient) send(command string) {
	c.t.Helper()

	if err := c.conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		c.t.Fatalf("cannot set a deadline: %v", err)
	}

	if _, err := fmt.Fprintf(c.conn, "%s\r\n", command); err != nil {
		c.t.Fatalf("cannot send %q: %v", command, err)
	}
}

// reply reads one reply, joining the lines of a multi line one.
func (c *ftpClient) reply() (int, string) {
	c.t.Helper()

	line, err := c.reader.ReadString('\n')
	if err != nil {
		c.t.Fatalf("cannot read a reply: %v", err)
	}

	line = strings.TrimRight(line, "\r\n")
	if len(line) < 4 {
		c.t.Fatalf("reply %q is too short to hold a code", line)
	}

	code, err := strconv.Atoi(line[:3])
	if err != nil {
		c.t.Fatalf("reply %q does not start with a code", line)
	}

	if line[3] == '-' {
		end := line[:3] + " "
		for !strings.HasPrefix(line, end) {
			line, err = c.reader.ReadString('\n')
			if err != nil {
				c.t.Fatalf("cannot read the end of a multi line reply: %v", err)
			}

			line = strings.TrimRight(line, "\r\n")
		}
	}

	return code, line
}

func (c *ftpClient) expect(want int) string {
	c.t.Helper()

	code, line := c.reply()
	if code != want {
		c.t.Fatalf("got %d %q, want %d", code, line, want)
	}

	return line
}

func (c *ftpClient) login(user, password string) {
	c.t.Helper()

	c.send("USER " + user)
	c.expect(331)
	c.send("PASS " + password)
	c.expect(230)
	c.send("TYPE I")
	c.expect(200)
}

// passive asks for a passive port and connects to it. Only the port of the
// reply is used: the host is the one already connected to, which is what a
// client behind NAT ends up doing anyway.
func (c *ftpClient) passive() net.Conn {
	c.t.Helper()

	c.send("PASV")
	line := c.expect(227)

	open := strings.LastIndex(line, "(")
	close := strings.LastIndex(line, ")")
	if open < 0 || close < open {
		c.t.Fatalf("PASV reply %q holds no address", line)
	}

	fields := strings.Split(line[open+1:close], ",")
	if len(fields) != 6 {
		c.t.Fatalf("PASV reply %q holds %d fields, want 6", line, len(fields))
	}

	high, err := strconv.Atoi(strings.TrimSpace(fields[4]))
	if err != nil {
		c.t.Fatalf("PASV reply %q holds no port: %v", line, err)
	}

	low, err := strconv.Atoi(strings.TrimSpace(fields[5]))
	if err != nil {
		c.t.Fatalf("PASV reply %q holds no port: %v", line, err)
	}

	return c.dialAddr(net.JoinHostPort("127.0.0.1", strconv.Itoa(high*256+low)))
}

func (c *ftpClient) store(name, content string) {
	c.t.Helper()

	data := c.passive()
	defer func() { _ = data.Close() }()

	c.send("STOR " + name)
	c.expect(150)

	if _, err := io.WriteString(data, content); err != nil {
		c.t.Fatalf("cannot send the content of %s: %v", name, err)
	}

	if err := data.Close(); err != nil {
		c.t.Fatalf("cannot close the data connection: %v", err)
	}

	c.expect(226)
}

func (c *ftpClient) retrieve(name string) string {
	c.t.Helper()

	data := c.passive()
	defer func() { _ = data.Close() }()

	c.send("RETR " + name)
	c.expect(150)

	content, err := io.ReadAll(data)
	if err != nil {
		c.t.Fatalf("cannot read the content of %s: %v", name, err)
	}

	c.expect(226)

	return string(content)
}
