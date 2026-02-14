package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestLoveLetterGetAllDataReturnsSnapshot(t *testing.T) {
	client := setupLoveLetterTestClient(t)
	state := &orm.CommanderLoveLetterState{
		CommanderID: client.Commander.CommanderID,
		Medals: []orm.LoveLetterMedalState{{
			GroupID: 10000,
			Exp:     10,
			Level:   1,
		}},
		ManualLetters: []orm.LoveLetterLetterState{{
			GroupID:      10000,
			LetterIDList: []uint32{2019001},
		}},
		ConvertedItems: []orm.LoveLetterConvertedItem{{
			ItemID:  41002,
			GroupID: 10000,
			Year:    2018,
		}},
		RewardedIDs: []uint32{1},
	}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save love letter state: %v", err)
	}
	payload := marshalPacketRequest(t, &protobuf.CS_12406{Type: proto.Uint32(0)})
	if _, _, err := LoveLetterGetAllData(&payload, client); err != nil {
		t.Fatalf("LoveLetterGetAllData failed: %v", err)
	}
	response := &protobuf.SC_12407{}
	decodeLoveLetterPacketMessage(t, client, 12407, response)
	if len(response.GetConvertedList()) != 1 || response.GetConvertedList()[0].GetItemId() != 41002 {
		t.Fatalf("unexpected converted items: %+v", response.GetConvertedList())
	}
	if len(response.GetRewardedList()) != 1 || response.GetRewardedList()[0] != 1 {
		t.Fatalf("unexpected rewarded ids: %+v", response.GetRewardedList())
	}
	if len(response.GetMedalList()) != 1 || response.GetMedalList()[0].GetGroupId() != 10000 {
		t.Fatalf("unexpected medals: %+v", response.GetMedalList())
	}
	if !hasProtoLetter(response.GetLetterList(), 10000, 2018001) || !hasProtoLetter(response.GetLetterList(), 10000, 2019001) {
		t.Fatalf("unexpected letters: %+v", response.GetLetterList())
	}
	if !hasProtoLetter(response.GetConvertedLetterList(), 10000, 2018001) || hasProtoLetter(response.GetConvertedLetterList(), 10000, 2019001) {
		t.Fatalf("unexpected converted letters: %+v", response.GetConvertedLetterList())
	}
}

func TestLoveLetterUnlockSuccessAndFailures(t *testing.T) {
	client := setupLoveLetterTestClient(t)
	state := &orm.CommanderLoveLetterState{
		CommanderID: client.Commander.CommanderID,
		Medals: []orm.LoveLetterMedalState{{
			GroupID: 10000,
			Exp:     10,
			Level:   1,
		}},
	}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	payload := marshalPacketRequest(t, &protobuf.CS_12400{Id: proto.Uint32(2018001)})
	if _, _, err := LoveLetterUnlock(&payload, client); err != nil {
		t.Fatalf("LoveLetterUnlock failed: %v", err)
	}
	response := &protobuf.SC_12401{}
	decodeLoveLetterPacketMessage(t, client, 12401, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	loaded, err := orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if !hasStoredLetter(loaded.ManualLetters, 10000, 2018001) {
		t.Fatalf("expected manual unlock persisted, got %+v", loaded.ManualLetters)
	}

	payload = marshalPacketRequest(t, &protobuf.CS_12400{Id: proto.Uint32(2018001)})
	if _, _, err := LoveLetterUnlock(&payload, client); err != nil {
		t.Fatalf("LoveLetterUnlock duplicate failed: %v", err)
	}
	response = &protobuf.SC_12401{}
	decodeLoveLetterPacketMessage(t, client, 12401, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected duplicate unlock to fail, got %d", response.GetResult())
	}

	payload = marshalPacketRequest(t, &protobuf.CS_12400{Id: proto.Uint32(2019001)})
	if _, _, err := LoveLetterUnlock(&payload, client); err != nil {
		t.Fatalf("LoveLetterUnlock low level failed: %v", err)
	}
	response = &protobuf.SC_12401{}
	decodeLoveLetterPacketMessage(t, client, 12401, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected low-level unlock to fail, got %d", response.GetResult())
	}
}

func TestLoveLetterClaimRewardsSuccessAndClaimedFailure(t *testing.T) {
	client := setupLoveLetterTestClient(t)
	state := &orm.CommanderLoveLetterState{
		CommanderID: client.Commander.CommanderID,
		Medals: []orm.LoveLetterMedalState{{
			GroupID: 10000,
			Exp:     10,
			Level:   1,
		}},
	}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	startGold := client.Commander.GetResourceCount(1)
	payload := marshalPacketRequest(t, &protobuf.CS_12402{IdList: []uint32{1}})
	if _, _, err := LoveLetterClaimRewards(&payload, client); err != nil {
		t.Fatalf("LoveLetterClaimRewards failed: %v", err)
	}
	response := &protobuf.SC_12403{}
	decodeLoveLetterPacketMessage(t, client, 12403, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetDropList()) != 1 || response.GetDropList()[0].GetType() != 1 || response.GetDropList()[0].GetId() != 1 || response.GetDropList()[0].GetNumber() != 50 {
		t.Fatalf("unexpected drops: %+v", response.GetDropList())
	}
	if client.Commander.GetResourceCount(1) != startGold+50 {
		t.Fatalf("expected gold increase by 50")
	}
	loaded, err := orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if len(loaded.RewardedIDs) != 1 || loaded.RewardedIDs[0] != 1 {
		t.Fatalf("expected reward mark persisted, got %+v", loaded.RewardedIDs)
	}

	payload = marshalPacketRequest(t, &protobuf.CS_12402{IdList: []uint32{1}})
	if _, _, err := LoveLetterClaimRewards(&payload, client); err != nil {
		t.Fatalf("LoveLetterClaimRewards duplicate failed: %v", err)
	}
	response = &protobuf.SC_12403{}
	decodeLoveLetterPacketMessage(t, client, 12403, response)
	if response.GetResult() != 1 || len(response.GetDropList()) != 0 {
		t.Fatalf("expected duplicate claim failure with empty drops, got result=%d drops=%+v", response.GetResult(), response.GetDropList())
	}
}

func TestLoveLetterRealizeGiftAdjustsMedals(t *testing.T) {
	client := setupLoveLetterTestClient(t)
	state := &orm.CommanderLoveLetterState{CommanderID: client.Commander.CommanderID}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save initial state: %v", err)
	}
	payload := marshalPacketRequest(t, &protobuf.CS_12404{ItemList: []*protobuf.PT_OLD_LOVER_ITEM{{
		ItemId:  proto.Uint32(41002),
		GroupId: proto.Uint32(10000),
		Year:    proto.Uint32(2018),
	}}})
	if _, _, err := LoveLetterRealizeGift(&payload, client); err != nil {
		t.Fatalf("LoveLetterRealizeGift failed: %v", err)
	}
	response := &protobuf.SC_12405{}
	decodeLoveLetterPacketMessage(t, client, 12405, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	loaded, err := orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state after realize gift: %v", err)
	}
	if len(loaded.Medals) != 1 || loaded.Medals[0].Exp != 10 || loaded.Medals[0].Level != 1 {
		t.Fatalf("unexpected medal state after gift: %+v", loaded.Medals)
	}

	payload = marshalPacketRequest(t, &protobuf.CS_12404{})
	if _, _, err := LoveLetterRealizeGift(&payload, client); err != nil {
		t.Fatalf("LoveLetterRealizeGift reset failed: %v", err)
	}
	response = &protobuf.SC_12405{}
	decodeLoveLetterPacketMessage(t, client, 12405, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected reset result 0, got %d", response.GetResult())
	}
	loaded, err = orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state after reset: %v", err)
	}
	if len(loaded.Medals) != 1 || loaded.Medals[0].Exp != 0 || loaded.Medals[0].Level != 0 {
		t.Fatalf("unexpected medal state after reset: %+v", loaded.Medals)
	}
}

func TestLoveLetterLevelUp(t *testing.T) {
	client := setupLoveLetterTestClient(t)
	state := &orm.CommanderLoveLetterState{
		CommanderID: client.Commander.CommanderID,
		Medals: []orm.LoveLetterMedalState{{
			GroupID: 10000,
			Exp:     20,
			Level:   1,
		}},
	}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	payload := marshalPacketRequest(t, &protobuf.CS_12408{GroupId: proto.Uint32(10000)})
	if _, _, err := LoveLetterLevelUp(&payload, client); err != nil {
		t.Fatalf("LoveLetterLevelUp failed: %v", err)
	}
	response := &protobuf.SC_12409{}
	decodeLoveLetterPacketMessage(t, client, 12409, response)
	if response.GetRet() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetRet())
	}
	loaded, err := orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if loaded.Medals[0].Level != 2 {
		t.Fatalf("expected level 2, got %+v", loaded.Medals)
	}

	payload = marshalPacketRequest(t, &protobuf.CS_12408{GroupId: proto.Uint32(10000)})
	if _, _, err := LoveLetterLevelUp(&payload, client); err != nil {
		t.Fatalf("LoveLetterLevelUp second call failed: %v", err)
	}
	response = &protobuf.SC_12409{}
	decodeLoveLetterPacketMessage(t, client, 12409, response)
	if response.GetRet() != 1 {
		t.Fatalf("expected second level up to fail, got %d", response.GetRet())
	}
}

func TestLoveLetterGetContentPriority(t *testing.T) {
	client := setupLoveLetterTestClient(t)
	state := &orm.CommanderLoveLetterState{
		CommanderID:    client.Commander.CommanderID,
		LetterContents: map[uint32]string{2018001: "state text"},
	}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	payload := marshalPacketRequest(t, &protobuf.CS_12410{LetterId: proto.Uint32(2018001)})
	if _, _, err := LoveLetterGetContent(&payload, client); err != nil {
		t.Fatalf("LoveLetterGetContent state failed: %v", err)
	}
	response := &protobuf.SC_12411{}
	decodeLoveLetterPacketMessage(t, client, 12411, response)
	if response.GetContent() != "state text" {
		t.Fatalf("expected state content, got %q", response.GetContent())
	}

	state.LetterContents = map[uint32]string{}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("clear state letter contents: %v", err)
	}
	payload = marshalPacketRequest(t, &protobuf.CS_12410{LetterId: proto.Uint32(2019001)})
	if _, _, err := LoveLetterGetContent(&payload, client); err != nil {
		t.Fatalf("LoveLetterGetContent config failed: %v", err)
	}
	response = &protobuf.SC_12411{}
	decodeLoveLetterPacketMessage(t, client, 12411, response)
	if response.GetContent() != "config text" {
		t.Fatalf("expected config text, got %q", response.GetContent())
	}

	payload = marshalPacketRequest(t, &protobuf.CS_12410{LetterId: proto.Uint32(999999)})
	if _, _, err := LoveLetterGetContent(&payload, client); err != nil {
		t.Fatalf("LoveLetterGetContent empty failed: %v", err)
	}
	response = &protobuf.SC_12411{}
	decodeLoveLetterPacketMessage(t, client, 12411, response)
	if response.GetContent() != "" {
		t.Fatalf("expected empty content, got %q", response.GetContent())
	}
}

func setupLoveLetterTestClient(t *testing.T) *connection.Client {
	t.Helper()
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.CommanderLoveLetterState{})
	seedLoveLetterConfig(t)
	return client
}

func seedLoveLetterConfig(t *testing.T) {
	t.Helper()
	seedConfigEntry(t, loveLetterCharacterTemplateCategory, "10000", `{"id":10000,"exp_up":10,"exp_upper_limit":50,"relate_group_id":[]}`)
	seedConfigEntry(t, loveLetterContentTemplateCategory, "2018001", `{"id":2018001,"ship_group":10000,"year":2018,"love_item":[41002],"content":""}`)
	seedConfigEntry(t, loveLetterContentTemplateCategory, "2019001", `{"id":2019001,"ship_group":10000,"year":2019,"love_item":[51002],"content":"config text"}`)
	seedConfigEntry(t, loveLetterRewardTemplateCategory, "1", `{"id":1,"total_level":1,"show_reward":[[1,1,50]]}`)
	seedConfigEntry(t, loveLetterLegacyTemplateCategory, "41002", `{"id":41002,"ship_group_id":10000,"year":2018}`)
}

func marshalPacketRequest(t *testing.T, message proto.Message) []byte {
	t.Helper()
	payload, err := proto.Marshal(message)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return payload
}

func decodeLoveLetterPacketMessage(t *testing.T, client *connection.Client, expectedPacketID int, message proto.Message) {
	t.Helper()
	payload := decodeRawPacketPayload(t, client, expectedPacketID)
	if err := proto.Unmarshal(payload, message); err != nil {
		t.Fatalf("unmarshal packet %d: %v", expectedPacketID, err)
	}
}

func decodeRawPacketPayload(t *testing.T, client *connection.Client, expectedPacketID int) []byte {
	t.Helper()
	data := client.Buffer.Bytes()
	if len(data) < packets.HEADER_SIZE {
		t.Fatalf("expected packet header")
	}
	packetID := packets.GetPacketId(0, &data)
	if packetID != expectedPacketID {
		t.Fatalf("expected packet %d, got %d", expectedPacketID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &data) + 2
	if len(data) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(data))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	payload := append([]byte{}, data[payloadStart:payloadEnd]...)
	client.Buffer.Reset()
	return payload
}

func hasProtoLetter(entries []*protobuf.PT_SHIP_LOVE_LETTER, groupID uint32, letterID uint32) bool {
	for _, entry := range entries {
		if entry.GetGroupId() != groupID {
			continue
		}
		for _, id := range entry.GetLetterIdList() {
			if id == letterID {
				return true
			}
		}
	}
	return false
}

func hasStoredLetter(entries []orm.LoveLetterLetterState, groupID uint32, letterID uint32) bool {
	for _, entry := range entries {
		if entry.GroupID != groupID {
			continue
		}
		for _, id := range entry.LetterIDList {
			if id == letterID {
				return true
			}
		}
	}
	return false
}
