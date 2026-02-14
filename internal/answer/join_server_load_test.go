package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestJoinServerResponseContainsDbLoad(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_10022{
		AccountId:    proto.Uint32(0),
		ServerTicket: proto.String(serverTicketPrefix),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(1),
		CheckKey:     proto.String("status_probe"),
		DeviceId:     proto.String(""),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.JoinServer(&buf, client); err != nil {
		t.Fatalf("JoinServer failed: %v", err)
	}
	response := &protobuf.SC_10023{}
	decodeResponsePacket(t, client, 10023, response)

	if response.DbLoad == nil {
		t.Fatalf("expected db_load to be set")
	}
	if response.GetDbLoad() > 100 {
		t.Fatalf("expected db_load <= 100, got %d", response.GetDbLoad())
	}
}

func TestJoinServerResponseDbLoadWithoutStore(t *testing.T) {
	originalStore := db.DefaultStore
	db.DefaultStore = nil
	defer func() { db.DefaultStore = originalStore }()

	client := &connection.Client{}
	payload := &protobuf.CS_10022{
		AccountId:    proto.Uint32(0),
		ServerTicket: proto.String(serverTicketPrefix),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(1),
		CheckKey:     proto.String("status_probe"),
		DeviceId:     proto.String(""),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.JoinServer(&buf, client); err != nil {
		t.Fatalf("JoinServer failed: %v", err)
	}
	response := &protobuf.SC_10023{}
	decodeResponsePacket(t, client, 10023, response)

	if response.DbLoad == nil {
		t.Fatalf("expected db_load to be set")
	}
	if response.GetDbLoad() != 0 {
		t.Fatalf("expected db_load 0 without default store, got %d", response.GetDbLoad())
	}
}
