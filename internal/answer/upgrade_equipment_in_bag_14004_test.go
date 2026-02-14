package answer_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func decodeSC14005(t *testing.T, client *connection.Client) *protobuf.SC_14005 {
	t.Helper()

	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != 14005 {
		t.Fatalf("expected packet 14005, got %d", packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	var response protobuf.SC_14005
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
	return &response
}

func createTestEquipment(t *testing.T, id uint32, next uint32, transUseGold uint32, transUseItem json.RawMessage) {
	t.Helper()
	transUseItemJSON := "[]"
	if len(transUseItem) > 0 {
		transUseItemJSON = string(transUseItem)
	}
	payload := fmt.Sprintf(`{"id":%d,"next":%d,"trans_use_gold":%d,"trans_use_item":%s,"ship_type_forbidden":[]}`,
		id,
		next,
		transUseGold,
		transUseItemJSON,
	)
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", "sharecfgdata/equip_data_statistics.json", fmt.Sprintf("%d", id), payload)
}

func seedUpgradeEquipmentCostDefs(t *testing.T) {
	t.Helper()
	execAnswerExternalTestSQLT(t, "INSERT INTO resources (id, item_id, name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", int64(1), int64(0), "Gold")
	execAnswerExternalTestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING", int64(200), "Upgrade Material", int64(1), int64(0), int64(1), int64(0))
}

func TestUpgradeEquipmentInBag14004EquipIDZero(t *testing.T) {
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 9004701, AccountID: 9004701, Name: "equip-zero"}}

	payload := &protobuf.CS_14004{EquipId: proto.Uint32(0), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentInBag14004LvZero(t *testing.T) {
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 9004702, AccountID: 9004702, Name: "lv-zero"}}

	payload := &protobuf.CS_14004{EquipId: proto.Uint32(100), Lv: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentInBag14004SuccessMutatesBag(t *testing.T) {
	commander := orm.Commander{CommanderID: 9004703, AccountID: 9004703, Name: "upgrade-success"}
	if err := orm.CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}

	baseID := uint32(991000)
	upgradedID := uint32(991001)
	createTestEquipment(t, baseID, upgradedID, 0, nil)
	createTestEquipment(t, upgradedID, 0, 0, nil)

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(commander.CommanderID), int64(baseID), int64(1))

	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_14004{EquipId: proto.Uint32(baseID), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	baseRows := queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM owned_equipments WHERE commander_id = $1 AND equipment_id = $2", int64(commander.CommanderID), int64(baseID))
	if baseRows != 0 {
		t.Fatalf("expected base equipment to be removed")
	}
	upgradedCount := queryAnswerExternalTestInt64(t, "SELECT count FROM owned_equipments WHERE commander_id = $1 AND equipment_id = $2", int64(commander.CommanderID), int64(upgradedID))
	if upgradedCount != 1 {
		t.Fatalf("expected upgraded count 1, got %d", upgradedCount)
	}
}

func TestUpgradeEquipmentInBag14004MissingChainDoesNotMutate(t *testing.T) {
	commander := orm.Commander{CommanderID: 9004704, AccountID: 9004704, Name: "upgrade-missing"}
	if err := orm.CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}

	baseID := uint32(992000)
	createTestEquipment(t, baseID, 0, 0, nil)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(commander.CommanderID), int64(baseID), int64(1))

	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_14004{EquipId: proto.Uint32(baseID), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	baseCount := queryAnswerExternalTestInt64(t, "SELECT count FROM owned_equipments WHERE commander_id = $1 AND equipment_id = $2", int64(commander.CommanderID), int64(baseID))
	if baseCount != 1 {
		t.Fatalf("expected base count 1, got %d", baseCount)
	}
}

func TestUpgradeEquipmentInBag14004ChargesUpgradeCosts(t *testing.T) {
	commander := orm.Commander{CommanderID: 9004705, AccountID: 9004705, Name: "upgrade-costs"}
	if err := orm.CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	seedUpgradeEquipmentCostDefs(t)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(commander.CommanderID), int64(1), int64(100))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(commander.CommanderID), int64(200), int64(3))

	baseID := uint32(993000)
	upgradedID := uint32(993001)
	createTestEquipment(t, baseID, upgradedID, 10, json.RawMessage(`[[200,1]]`))
	createTestEquipment(t, upgradedID, 0, 0, nil)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(commander.CommanderID), int64(baseID), int64(1))

	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_14004{EquipId: proto.Uint32(baseID), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	goldAmount := queryAnswerExternalTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(commander.CommanderID), int64(1))
	if goldAmount != 90 {
		t.Fatalf("expected gold 90, got %d", goldAmount)
	}
	itemCount := queryAnswerExternalTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(commander.CommanderID), int64(200))
	if itemCount != 2 {
		t.Fatalf("expected item count 2, got %d", itemCount)
	}
}
