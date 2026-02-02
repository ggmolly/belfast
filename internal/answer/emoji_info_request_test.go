package answer

import (
	"reflect"
	"testing"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestEmojiInfoRequestFiltersAchieve(t *testing.T) {
	client := setupConfigTest(t)
	seedConfigEntry(t, "ShareCfg/emoji_template.json", "1", `{"id":1,"achieve":0}`)
	seedConfigEntry(t, "ShareCfg/emoji_template.json", "3", `{"id":3,"achieve":1}`)
	seedConfigEntry(t, "ShareCfg/emoji_template.json", "10", `{"id":10,"achieve":1}`)

	payload := protobuf.CS_11601{Type: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := EmojiInfoRequest(&buffer, client); err != nil {
		t.Fatalf("EmojiInfoRequest failed: %v", err)
	}

	packetIDs := decodePacketIDs(t, client.Buffer.Bytes())
	if len(packetIDs) != 1 || packetIDs[0] != 11602 {
		t.Fatalf("expected packet id 11602, got %v", packetIDs)
	}

	response := &protobuf.SC_11602{}
	if err := proto.Unmarshal(decodeFirstPacketPayload(t, client.Buffer.Bytes()), response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !reflect.DeepEqual(response.EmojiList, []uint32{3, 10}) {
		t.Fatalf("expected emoji list [3 10], got %v", response.EmojiList)
	}
}

func TestEmojiInfoRequestEmptyConfig(t *testing.T) {
	client := setupConfigTest(t)
	payload := protobuf.CS_11601{Type: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := EmojiInfoRequest(&buffer, client); err != nil {
		t.Fatalf("EmojiInfoRequest failed: %v", err)
	}

	response := &protobuf.SC_11602{}
	if err := proto.Unmarshal(decodeFirstPacketPayload(t, client.Buffer.Bytes()), response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(response.EmojiList) != 0 {
		t.Fatalf("expected empty emoji list, got %v", response.EmojiList)
	}
}
