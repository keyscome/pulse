package checker

import (
	"net"
	"strings"
	"testing"
	"time"
)

// startFakeRedis starts a minimal fake Redis server that listens on a random
// local TCP port. It accepts one connection, optionally validates AUTH, and
// responds to PING with +PONG.
//
// Parameters:
//   - password: if non-empty, the server expects an AUTH command with this
//     password before accepting PING; sending the wrong password returns -ERR.
//   - ready: closed when the server is ready to accept connections.
func startFakeRedis(t *testing.T, password string, ready chan struct{}) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake redis: listen: %v", err)
	}

	go func() {
		defer ln.Close()
		close(ready)

		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 512)
		authDone := password == ""

		for {
			n, err := conn.Read(buf)
			if err != nil || n == 0 {
				return
			}
			msg := string(buf[:n])

			if !authDone {
				// Expect AUTH command
				if containsAUTH(msg, password) {
					conn.Write([]byte("+OK\r\n"))
					authDone = true
				} else {
					conn.Write([]byte("-ERR invalid password\r\n"))
					return
				}
				continue
			}

			// Expect PING
			if containsPING(msg) {
				conn.Write([]byte("+PONG\r\n"))
				return
			}
		}
	}()

	return ln.Addr().String()
}

func containsAUTH(msg, password string) bool {
	return strings.Contains(msg, "AUTH") && strings.Contains(msg, password)
}

func containsPING(msg string) bool {
	return strings.Contains(msg, "PING")
}

func TestCheckConnection_Success(t *testing.T) {
	ready := make(chan struct{})
	addr := startFakeRedis(t, "", ready)
	<-ready

	// The fake server will handle one Redis PING; for CheckConnection we just
	// need the port to be open — dial immediately before the goroutine exits.
	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln2.Close()

	if err := CheckConnection(ln2.Addr().String(), time.Second); err != nil {
		t.Errorf("expected success, got: %v", err)
	}
	_ = addr // used by fake redis goroutine
}

func TestCheckConnection_Failure(t *testing.T) {
	// Use a port that is not listening
	err := CheckConnection("127.0.0.1:1", time.Second)
	if err == nil {
		t.Error("expected error for non-listening port, got nil")
	}
}

func TestCheckRedisConnection_NoPassword(t *testing.T) {
	ready := make(chan struct{})
	addr := startFakeRedis(t, "", ready)
	<-ready

	if err := CheckRedisConnection(addr, "", 2*time.Second); err != nil {
		t.Errorf("expected success without password, got: %v", err)
	}
}

func TestCheckRedisConnection_WithCorrectPassword(t *testing.T) {
	ready := make(chan struct{})
	addr := startFakeRedis(t, "secret", ready)
	<-ready

	if err := CheckRedisConnection(addr, "secret", 2*time.Second); err != nil {
		t.Errorf("expected success with correct password, got: %v", err)
	}
}

func TestCheckRedisConnection_WithWrongPassword(t *testing.T) {
	ready := make(chan struct{})
	addr := startFakeRedis(t, "secret", ready)
	<-ready

	if err := CheckRedisConnection(addr, "wrong", 2*time.Second); err == nil {
		t.Error("expected error with wrong password, got nil")
	}
}

func TestCheckRedisConnection_Unreachable(t *testing.T) {
	err := CheckRedisConnection("127.0.0.1:1", "", time.Second)
	if err == nil {
		t.Error("expected error for non-listening port, got nil")
	}
}
