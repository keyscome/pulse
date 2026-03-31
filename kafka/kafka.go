// kafka/kafka.go
package kafka

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// CheckConnection verifies that the given address hosts a live Kafka broker by
// performing a Kafka protocol-level handshake (ApiVersionsRequest v0).
// Unlike a plain TCP check, this confirms the remote endpoint speaks the Kafka
// binary protocol and is ready to accept client connections.
func CheckConnection(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}

	// ApiVersionsRequest v0 wire format (Kafka Request/Response API):
	//   [4]  request size   – byte count that follows (big-endian int32)
	//   [2]  api_key        – 18 = ApiVersions
	//   [2]  api_version    – 0
	//   [4]  correlation_id – echoed back in the response
	//   [2]  client_id      – 0xFFFF = null string
	// Payload is 10 bytes (2+2+4+2), so request size field = 10.
	var req [14]byte
	binary.BigEndian.PutUint32(req[0:4], 10)       // request size
	binary.BigEndian.PutUint16(req[4:6], 18)       // ApiVersions key
	binary.BigEndian.PutUint16(req[6:8], 0)        // version 0
	binary.BigEndian.PutUint32(req[8:12], 1)       // correlation ID = 1
	binary.BigEndian.PutUint16(req[12:14], 0xffff) // null client ID

	if _, err := conn.Write(req[:]); err != nil {
		return fmt.Errorf("write ApiVersionsRequest: %w", err)
	}

	// Response wire format:
	//   [4]  response size  – byte count that follows (big-endian int32)
	//   [4]  correlation_id – must match the request
	//   [2]  error_code     – 0 = success
	//   ...  api versions list (ignored)
	var respSize uint32
	if err := binary.Read(conn, binary.BigEndian, &respSize); err != nil {
		return fmt.Errorf("read response size: %w", err)
	}
	// Sanity check: a real ApiVersionsResponse will never be larger than a few KB.
	// Guard against a pathological server sending a huge size value.
	const maxRespSize = 1 << 20 // 1 MiB
	if respSize < 6 || respSize > maxRespSize {
		return fmt.Errorf("unexpected response size: %d bytes", respSize)
	}

	header := make([]byte, 6)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("read response header: %w", err)
	}

	if corrID := binary.BigEndian.Uint32(header[:4]); corrID != 1 {
		return fmt.Errorf("correlation ID mismatch: expected 1, got %d", corrID)
	}
	if code := binary.BigEndian.Uint16(header[4:6]); code != 0 {
		return fmt.Errorf("broker returned error code %d", code)
	}

	return nil
}
