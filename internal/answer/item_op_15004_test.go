package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestItemOp15004Success(t *testing.T) {
	client := setupHandlerCommander(t)
	client.Buffer.Reset()

	payload := protobuf.CS_15004{Id: proto.Uint32(1001), Count: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := ItemOp15004(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_15005
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestItemOp15004UnmarshalFailure(t *testing.T) {
	client := setupHandlerCommander(t)
	client.Buffer.Reset()

	buffer := []byte{0xff}
	if _, _, err := ItemOp15004(&buffer, client); err == nil {
		t.Fatalf("expected unmarshal error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response to be written")
	}
}
