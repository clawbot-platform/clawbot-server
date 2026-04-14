package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"clawbot-server/internal/config"
)

func TestCheckHTTP(t *testing.T) {
	originalTransport := httpTransport
	t.Cleanup(func() { httpTransport = originalTransport })

	httpTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "http://foundation.test/health" {
			t.Fatalf("unexpected URL %q", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})

	if err := checkHTTP(context.Background(), "http://foundation.test/health", time.Second); err != nil {
		t.Fatalf("checkHTTP() error = %v", err)
	}
}

func TestCheckRedis(t *testing.T) {
	restore := withDialStub(t, func(server net.Conn) {
		defer func() {
			_ = server.Close()
		}()

		request := make([]byte, len("*1\r\n$4\r\nPING\r\n"))
		_, err := io.ReadFull(server, request)
		if err != nil {
			t.Fatalf("ReadFull() error = %v", err)
		}
		if string(request) != "*1\r\n$4\r\nPING\r\n" {
			t.Fatalf("unexpected redis request %q", string(request))
		}
		_, _ = server.Write([]byte("+PONG\r\n"))
	})
	defer restore()

	cfg := config.Foundation{
		RedisHost: "127.0.0.1",
		RedisPort: "6379",
		Timeout:   time.Second,
	}

	if err := checkRedis(context.Background(), cfg); err != nil {
		t.Fatalf("checkRedis() error = %v", err)
	}
}

func TestCheckNATS(t *testing.T) {
	restore := withDialStub(t, func(server net.Conn) {
		defer func() {
			_ = server.Close()
		}()

		_, _ = server.Write([]byte("INFO {\"server_id\":\"test\"}\r\n"))
		reader := bufio.NewReader(server)
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("ReadString() error = %v", err)
		}
		if strings.TrimSpace(line) != "PING" {
			t.Fatalf("unexpected nats request %q", strings.TrimSpace(line))
		}
		_, _ = server.Write([]byte("PONG\r\n"))
	})
	defer restore()

	cfg := config.Foundation{
		NATSHost: "127.0.0.1",
		NATSPort: "4222",
		Timeout:  time.Second,
	}

	if err := checkNATS(context.Background(), cfg); err != nil {
		t.Fatalf("checkNATS() error = %v", err)
	}
}

func TestCheckPostgresAuthRequestIsReady(t *testing.T) {
	restore := withDialStub(t, func(server net.Conn) {
		defer func() {
			_ = server.Close()
		}()

		header := make([]byte, 4)
		if _, err := io.ReadFull(server, header); err != nil {
			t.Fatalf("ReadFull(header) error = %v", err)
		}
		length := int(binary.BigEndian.Uint32(header))
		body := make([]byte, length-4)
		if _, err := io.ReadFull(server, body); err != nil {
			t.Fatalf("ReadFull(body) error = %v", err)
		}

		reply := make([]byte, 9)
		reply[0] = 'R'
		binary.BigEndian.PutUint32(reply[1:5], 8)
		binary.BigEndian.PutUint32(reply[5:9], 3)
		_, _ = server.Write(reply)
	})
	defer restore()

	cfg := config.Foundation{
		PostgresHost: "127.0.0.1",
		PostgresPort: "5432",
		PostgresUser: "clawbot",
		PostgresDB:   "clawbot",
		Timeout:      time.Second,
	}

	if err := checkPostgres(context.Background(), cfg); err != nil {
		t.Fatalf("checkPostgres() error = %v", err)
	}
}

func TestCheckPostgresServerStartingUpError(t *testing.T) {
	restore := withDialStub(t, func(server net.Conn) {
		defer func() {
			_ = server.Close()
		}()

		header := make([]byte, 4)
		if _, err := io.ReadFull(server, header); err != nil {
			t.Fatalf("ReadFull(header) error = %v", err)
		}
		length := int(binary.BigEndian.Uint32(header))
		body := make([]byte, length-4)
		if _, err := io.ReadFull(server, body); err != nil {
			t.Fatalf("ReadFull(body) error = %v", err)
		}

		payload := []byte("Mthe database system is starting up\x00\x00")
		reply := make([]byte, 5+len(payload))
		reply[0] = 'E'
		binary.BigEndian.PutUint32(reply[1:5], uint32(4+len(payload)))
		copy(reply[5:], payload)
		_, _ = server.Write(reply)
	})
	defer restore()

	cfg := config.Foundation{
		PostgresHost: "127.0.0.1",
		PostgresPort: "5432",
		PostgresUser: "clawbot",
		PostgresDB:   "clawbot",
		Timeout:      time.Second,
	}

	err := checkPostgres(context.Background(), cfg)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "starting up") {
		t.Fatalf("expected starting up error, got %v", err)
	}
}

func TestParsePostgresError(t *testing.T) {
	message := parsePostgresError([]byte("SERROR\x00Mprimary message\x00Ddetail text\x00\x00"))
	if message != "primary message: detail text" {
		t.Fatalf("unexpected postgres error parse %q", message)
	}

	unknown := parsePostgresError([]byte("\x00\x00"))
	if unknown != "unknown postgres error" {
		t.Fatalf("expected unknown fallback, got %q", unknown)
	}
}

func TestResolveIdentityHealthURL(t *testing.T) {
	t.Setenv("CLAWBOT_IDENTITY_BASE_URL", "http://127.0.0.1:9090")
	if got := resolveIdentityHealthURL(); got != "http://127.0.0.1:9090/healthz" {
		t.Fatalf("unexpected identity health url: %q", got)
	}
}

func TestResolveIdentityHealthURLEmpty(t *testing.T) {
	t.Setenv("CLAWBOT_IDENTITY_BASE_URL", "")
	if got := resolveIdentityHealthURL(); got != "" {
		t.Fatalf("expected empty identity health url, got %q", got)
	}
}

func withDialStub(t *testing.T, handler func(net.Conn)) func() {
	t.Helper()

	originalDial := dialTimeout
	dialTimeout = func(network string, address string, timeout time.Duration) (net.Conn, error) {
		client, server := net.Pipe()
		go handler(server)
		return client, nil
	}

	return func() {
		dialTimeout = originalDial
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
