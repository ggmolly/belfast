package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestProposeExchangeRing15010SuccessConsumesAndGrants(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	seedConfigEntry(t, "ShareCfg/gameset.json", "vow_prop_conversion", `{"key_value":0,"description":[15006,15011]}`)
	seedHandlerCommanderItem(t, client, 15006, 1)

	payload := protobuf.CS_15010{Id: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ProposeExchangeRing(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var resp protobuf.SC_15011
	decodeResponse(t, client, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}

	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	if got := client.Commander.GetItemCount(15006); got != 0 {
		t.Fatalf("expected ring consumed (0), got %d", got)
	}
	if got := client.Commander.GetItemCount(15011); got != 1 {
		t.Fatalf("expected tiara granted (1), got %d", got)
	}
}

func TestProposeExchangeRing15010FailureMissingItemDoesNotModifyInventory(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	seedConfigEntry(t, "ShareCfg/gameset.json", "vow_prop_conversion", `{"key_value":0,"description":[15006,15011]}`)

	payload := protobuf.CS_15010{Id: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ProposeExchangeRing(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var resp protobuf.SC_15011
	decodeResponse(t, client, &resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	if got := client.Commander.GetItemCount(15006); got != 0 {
		t.Fatalf("expected ring count unchanged (0), got %d", got)
	}
	if got := client.Commander.GetItemCount(15011); got != 0 {
		t.Fatalf("expected tiara count unchanged (0), got %d", got)
	}
}

func TestProposeExchangeRing15010ConfigDrivenPair(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	seedConfigEntry(t, "ShareCfg/gameset.json", "vow_prop_conversion", `{"key_value":0,"description":[20001,20002]}`)
	seedHandlerCommanderItem(t, client, 20001, 1)

	payload := protobuf.CS_15010{Id: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ProposeExchangeRing(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var resp protobuf.SC_15011
	decodeResponse(t, client, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}

	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	if got := client.Commander.GetItemCount(20001); got != 0 {
		t.Fatalf("expected from item consumed (0), got %d", got)
	}
	if got := client.Commander.GetItemCount(20002); got != 1 {
		t.Fatalf("expected to item granted (1), got %d", got)
	}
}
