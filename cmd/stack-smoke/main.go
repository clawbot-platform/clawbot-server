package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"clawbot-server/internal/config"
	"clawbot-server/internal/version"
)

type smokeCheck struct {
	name string
	run  func(context.Context, config.Foundation) error
}

var dialTimeout = net.DialTimeout
var httpTransport http.RoundTripper = http.DefaultTransport

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Value)
		return
	}

	cfg, err := config.LoadFoundationFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	checks := []smokeCheck{
		{name: "postgres", run: checkPostgres},
		{name: "redis", run: checkRedis},
		{name: "nats", run: checkNATS},
		{name: "minio", run: func(ctx context.Context, cfg config.Foundation) error {
			return checkHTTP(ctx, cfg.MinIOURL, cfg.Timeout)
		}},
		{name: "omniroute", run: func(ctx context.Context, cfg config.Foundation) error {
			return checkHTTP(ctx, cfg.OmniRouteURL, cfg.Timeout)
		}},
		{name: "zeroclaw", run: func(ctx context.Context, cfg config.Foundation) error {
			return checkHTTP(ctx, cfg.ZeroClawURL, cfg.Timeout)
		}},
		{name: "prometheus", run: func(ctx context.Context, cfg config.Foundation) error {
			return checkHTTP(ctx, cfg.PrometheusURL, cfg.Timeout)
		}},
		{name: "grafana", run: func(ctx context.Context, cfg config.Foundation) error {
			return checkHTTP(ctx, cfg.GrafanaURL, cfg.Timeout)
		}},
	}

	failed := false
	for _, check := range checks {
		if err := check.run(ctx, cfg); err != nil {
			fmt.Printf("FAIL %s: %v\n", check.name, err)
			failed = true
			continue
		}
		fmt.Printf("PASS %s\n", check.name)
	}

	if failed {
		os.Exit(1)
	}
}

func checkPostgres(_ context.Context, cfg config.Foundation) error {
	address := net.JoinHostPort(cfg.PostgresHost, cfg.PostgresPort)
	conn, err := dialTimeout("tcp", address, cfg.Timeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", address, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(cfg.Timeout))

	packet := buildStartupPacket(cfg.PostgresUser, cfg.PostgresDB)
	if _, err := conn.Write(packet); err != nil {
		return fmt.Errorf("write startup packet: %w", err)
	}

	header := make([]byte, 5)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("read response header: %w", err)
	}

	messageType := header[0]
	messageLength := int(binary.BigEndian.Uint32(header[1:5]))
	if messageLength < 4 {
		return fmt.Errorf("invalid postgres message length %d", messageLength)
	}

	body := make([]byte, messageLength-4)
	if _, err := io.ReadFull(conn, body); err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	switch messageType {
	case 'R':
		return nil
	case 'E':
		message := parsePostgresError(body)
		if strings.Contains(strings.ToLower(message), "starting up") {
			return errors.New(message)
		}
		return fmt.Errorf("postgres returned error: %s", message)
	default:
		return fmt.Errorf("unexpected postgres response type %q", string(messageType))
	}
}

func checkRedis(_ context.Context, cfg config.Foundation) error {
	address := net.JoinHostPort(cfg.RedisHost, cfg.RedisPort)
	conn, err := dialTimeout("tcp", address, cfg.Timeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", address, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(cfg.Timeout))

	if _, err := conn.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		return fmt.Errorf("write ping: %w", err)
	}

	reply := make([]byte, 7)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return fmt.Errorf("read ping reply: %w", err)
	}

	if string(reply) != "+PONG\r\n" {
		return fmt.Errorf("unexpected redis reply %q", string(reply))
	}

	return nil
}

func checkNATS(_ context.Context, cfg config.Foundation) error {
	address := net.JoinHostPort(cfg.NATSHost, cfg.NATSPort)
	conn, err := dialTimeout("tcp", address, cfg.Timeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", address, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(cfg.Timeout))

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read info line: %w", err)
	}

	if !strings.HasPrefix(line, "INFO ") {
		return fmt.Errorf("unexpected nats banner %q", strings.TrimSpace(line))
	}

	if _, err := conn.Write([]byte("PING\r\n")); err != nil {
		return fmt.Errorf("write ping: %w", err)
	}

	for {
		reply, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read ping reply: %w", err)
		}
		if strings.TrimSpace(reply) == "PONG" {
			return nil
		}
	}
}

func checkHTTP(ctx context.Context, url string, timeout time.Duration) error {
	client := &http.Client{
		Timeout:   timeout,
		Transport: httpTransport,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("get %s: %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("get %s: unexpected status %d", url, res.StatusCode)
	}

	return nil
}

func buildStartupPacket(user string, database string) []byte {
	payload := []byte("user\x00" + user + "\x00database\x00" + database + "\x00client_encoding\x00UTF8\x00\x00")
	packet := make([]byte, 8+len(payload))
	binary.BigEndian.PutUint32(packet[0:4], uint32(len(packet)))
	binary.BigEndian.PutUint32(packet[4:8], 196608)
	copy(packet[8:], payload)
	return packet
}

func parsePostgresError(body []byte) string {
	parts := strings.Split(string(body), "\x00")
	messages := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) < 2 {
			continue
		}
		kind := part[0]
		value := part[1:]
		if kind == 'M' || kind == 'D' {
			messages = append(messages, value)
		}
	}
	if len(messages) == 0 {
		return "unknown postgres error"
	}
	return strings.Join(messages, ": ")
}
