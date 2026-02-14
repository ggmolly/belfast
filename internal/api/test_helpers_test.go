package api_test

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/packets"
	"google.golang.org/protobuf/proto"
)

func decodeTestPacket(t *testing.T, buffer []byte, expectedId int, message proto.Message) int {
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetId := packets.GetPacketId(0, &buffer)
	if packetId != expectedId {
		t.Fatalf("expected packet %d, got %d", expectedId, packetId)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return packetId
}

func seedDb() {
	_ = execAPITestSQL("INSERT INTO resources (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", int64(1), "Gold")
	_ = execAPITestSQL("INSERT INTO resources (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", int64(2), "Fake resource")
	_ = execAPITestSQL("INSERT INTO resources (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", int64(4), "Gems")
	_ = execAPITestSQL("INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", int64(20001), "Wisdom Cube", int64(1), int64(0), int64(0), int64(0))
	_ = execAPITestSQL("INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", int64(45), "Fake Item", int64(1), int64(0), int64(0), int64(0))
	_ = execAPITestSQL("INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", int64(60), "Fake Item 2", int64(1), int64(0), int64(0), int64(0))
}

func execAPITestSQL(query string, args ...any) error {
	_, err := db.DefaultStore.Pool.Exec(context.Background(), query, args...)
	return err
}

func execAPITestSQLT(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), query, args...); err != nil {
		t.Fatalf("exec sql failed: %v", err)
	}
}

func queryAPITestInt64(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var value int64
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query row failed: %v", err)
	}
	return value
}
