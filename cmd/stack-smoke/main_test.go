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
		defer server.Close()

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
		defer server.Close()

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
		defer server.Close()

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
