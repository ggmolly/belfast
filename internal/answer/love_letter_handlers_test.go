package answer

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

type decodedSC12407 struct {
	ConvertedItems   []orm.LoveLetterConvertedItem
	RewardedIDs      []uint32
	Medals           []orm.LoveLetterMedalState
	Letters          []orm.LoveLetterLetterState
	ConvertedLetters []orm.LoveLetterLetterState
}

func TestLoveLetterGetAllData12406ReturnsSnapshot(t *testing.T) {
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
	payload := encodeCS12406Payload(0)
	if _, _, err := LoveLetterGetAllData12406(&payload, client); err != nil {
		t.Fatalf("LoveLetterGetAllData12406 failed: %v", err)
	}
	responsePayload := decodeRawPacketPayload(t, client, 12407)
	decoded, err := decodeSC12407Payload(responsePayload)
	if err != nil {
		t.Fatalf("decode sc_12407 failed: %v", err)
	}
	if len(decoded.ConvertedItems) != 1 || decoded.ConvertedItems[0].ItemID != 41002 {
		t.Fatalf("unexpected converted items: %+v", decoded.ConvertedItems)
	}
	if len(decoded.RewardedIDs) != 1 || decoded.RewardedIDs[0] != 1 {
		t.Fatalf("unexpected rewarded ids: %+v", decoded.RewardedIDs)
	}
	if len(decoded.Medals) != 1 || decoded.Medals[0].GroupID != 10000 {
		t.Fatalf("unexpected medals: %+v", decoded.Medals)
	}
	if !hasLetter(decoded.Letters, 10000, 2018001) || !hasLetter(decoded.Letters, 10000, 2019001) {
		t.Fatalf("unexpected letters: %+v", decoded.Letters)
	}
	if !hasLetter(decoded.ConvertedLetters, 10000, 2018001) || hasLetter(decoded.ConvertedLetters, 10000, 2019001) {
		t.Fatalf("unexpected converted letters: %+v", decoded.ConvertedLetters)
	}
}

func TestLoveLetterUnlock12400SuccessAndFailures(t *testing.T) {
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
	payload := encodeCS12400Payload(2018001)
	if _, _, err := LoveLetterUnlock12400(&payload, client); err != nil {
		t.Fatalf("LoveLetterUnlock12400 failed: %v", err)
	}
	responsePayload := decodeRawPacketPayload(t, client, 12401)
	result, ok, err := decodeSingleVarintField(responsePayload, 1)
	if err != nil || !ok {
		t.Fatalf("decode sc_12401 failed: %v", err)
	}
	if result != 0 {
		t.Fatalf("expected result 0, got %d", result)
	}
	loaded, err := orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if !hasLetter(loaded.ManualLetters, 10000, 2018001) {
		t.Fatalf("expected manual unlock persisted, got %+v", loaded.ManualLetters)
	}

	payload = encodeCS12400Payload(2018001)
	if _, _, err := LoveLetterUnlock12400(&payload, client); err != nil {
		t.Fatalf("LoveLetterUnlock12400 duplicate failed: %v", err)
	}
	responsePayload = decodeRawPacketPayload(t, client, 12401)
	result, _, _ = decodeSingleVarintField(responsePayload, 1)
	if result != 1 {
		t.Fatalf("expected duplicate unlock to fail, got %d", result)
	}

	payload = encodeCS12400Payload(2019001)
	if _, _, err := LoveLetterUnlock12400(&payload, client); err != nil {
		t.Fatalf("LoveLetterUnlock12400 low level failed: %v", err)
	}
	responsePayload = decodeRawPacketPayload(t, client, 12401)
	result, _, _ = decodeSingleVarintField(responsePayload, 1)
	if result != 1 {
		t.Fatalf("expected low-level unlock to fail, got %d", result)
	}
}

func TestLoveLetterClaimRewards12402SuccessAndClaimedFailure(t *testing.T) {
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
	payload := encodeCS12402Payload([]uint32{1})
	if _, _, err := LoveLetterClaimRewards12402(&payload, client); err != nil {
		t.Fatalf("LoveLetterClaimRewards12402 failed: %v", err)
	}
	responsePayload := decodeRawPacketPayload(t, client, 12403)
	result, drops, err := decodeSC12403Payload(responsePayload)
	if err != nil {
		t.Fatalf("decode sc_12403 failed: %v", err)
	}
	if result != 0 {
		t.Fatalf("expected result 0, got %d", result)
	}
	if len(drops) != 1 || drops[0].GetType() != 1 || drops[0].GetId() != 1 || drops[0].GetNumber() != 50 {
		t.Fatalf("unexpected drops: %+v", drops)
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

	payload = encodeCS12402Payload([]uint32{1})
	if _, _, err := LoveLetterClaimRewards12402(&payload, client); err != nil {
		t.Fatalf("LoveLetterClaimRewards12402 duplicate failed: %v", err)
	}
	responsePayload = decodeRawPacketPayload(t, client, 12403)
	result, drops, err = decodeSC12403Payload(responsePayload)
	if err != nil {
		t.Fatalf("decode duplicate sc_12403 failed: %v", err)
	}
	if result != 1 || len(drops) != 0 {
		t.Fatalf("expected duplicate claim failure with empty drops, got result=%d drops=%+v", result, drops)
	}
}

func TestLoveLetterRealizeGift12404AdjustsMedals(t *testing.T) {
	client := setupLoveLetterTestClient(t)
	state := &orm.CommanderLoveLetterState{CommanderID: client.Commander.CommanderID}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save initial state: %v", err)
	}
	payload := encodeCS12404Payload([]orm.LoveLetterConvertedItem{{
		ItemID:  41002,
		GroupID: 10000,
		Year:    2018,
	}})
	if _, _, err := LoveLetterRealizeGift12404(&payload, client); err != nil {
		t.Fatalf("LoveLetterRealizeGift12404 failed: %v", err)
	}
	responsePayload := decodeRawPacketPayload(t, client, 12405)
	result, _, err := decodeSingleVarintField(responsePayload, 1)
	if err != nil {
		t.Fatalf("decode sc_12405 failed: %v", err)
	}
	if result != 0 {
		t.Fatalf("expected result 0, got %d", result)
	}
	loaded, err := orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state after realize gift: %v", err)
	}
	if len(loaded.Medals) != 1 || loaded.Medals[0].Exp != 10 || loaded.Medals[0].Level != 1 {
		t.Fatalf("unexpected medal state after gift: %+v", loaded.Medals)
	}

	payload = encodeCS12404Payload([]orm.LoveLetterConvertedItem{})
	if _, _, err := LoveLetterRealizeGift12404(&payload, client); err != nil {
		t.Fatalf("LoveLetterRealizeGift12404 reset failed: %v", err)
	}
	responsePayload = decodeRawPacketPayload(t, client, 12405)
	result, _, _ = decodeSingleVarintField(responsePayload, 1)
	if result != 0 {
		t.Fatalf("expected reset result 0, got %d", result)
	}
	loaded, err = orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state after reset: %v", err)
	}
	if len(loaded.Medals) != 1 || loaded.Medals[0].Exp != 0 || loaded.Medals[0].Level != 0 {
		t.Fatalf("unexpected medal state after reset: %+v", loaded.Medals)
	}
}

func TestLoveLetterLevelUp12408(t *testing.T) {
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
	payload := encodeCS12408Payload(10000)
	if _, _, err := LoveLetterLevelUp12408(&payload, client); err != nil {
		t.Fatalf("LoveLetterLevelUp12408 failed: %v", err)
	}
	responsePayload := decodeRawPacketPayload(t, client, 12409)
	result, _, err := decodeSingleVarintField(responsePayload, 1)
	if err != nil {
		t.Fatalf("decode sc_12409 failed: %v", err)
	}
	if result != 0 {
		t.Fatalf("expected result 0, got %d", result)
	}
	loaded, err := orm.GetCommanderLoveLetterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if loaded.Medals[0].Level != 2 {
		t.Fatalf("expected level 2, got %+v", loaded.Medals)
	}

	payload = encodeCS12408Payload(10000)
	if _, _, err := LoveLetterLevelUp12408(&payload, client); err != nil {
		t.Fatalf("LoveLetterLevelUp12408 second call failed: %v", err)
	}
	responsePayload = decodeRawPacketPayload(t, client, 12409)
	result, _, _ = decodeSingleVarintField(responsePayload, 1)
	if result != 1 {
		t.Fatalf("expected second level up to fail, got %d", result)
	}
}

func TestLoveLetterGetContent12410Priority(t *testing.T) {
	client := setupLoveLetterTestClient(t)
	state := &orm.CommanderLoveLetterState{
		CommanderID:    client.Commander.CommanderID,
		LetterContents: map[uint32]string{2018001: "state text"},
	}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	payload := encodeCS12410Payload(2018001)
	if _, _, err := LoveLetterGetContent12410(&payload, client); err != nil {
		t.Fatalf("LoveLetterGetContent12410 state failed: %v", err)
	}
	responsePayload := decodeRawPacketPayload(t, client, 12411)
	content, err := decodeSC12411Payload(responsePayload)
	if err != nil {
		t.Fatalf("decode sc_12411 state failed: %v", err)
	}
	if content != "state text" {
		t.Fatalf("expected state content, got %q", content)
	}

	state.LetterContents = map[uint32]string{}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("clear state letter contents: %v", err)
	}
	payload = encodeCS12410Payload(2019001)
	if _, _, err := LoveLetterGetContent12410(&payload, client); err != nil {
		t.Fatalf("LoveLetterGetContent12410 config failed: %v", err)
	}
	responsePayload = decodeRawPacketPayload(t, client, 12411)
	content, err = decodeSC12411Payload(responsePayload)
	if err != nil {
		t.Fatalf("decode sc_12411 config failed: %v", err)
	}
	if content != "config text" {
		t.Fatalf("expected config text, got %q", content)
	}

	payload = encodeCS12410Payload(999999)
	if _, _, err := LoveLetterGetContent12410(&payload, client); err != nil {
		t.Fatalf("LoveLetterGetContent12410 empty failed: %v", err)
	}
	responsePayload = decodeRawPacketPayload(t, client, 12411)
	content, err = decodeSC12411Payload(responsePayload)
	if err != nil {
		t.Fatalf("decode sc_12411 empty failed: %v", err)
	}
	if content != "" {
		t.Fatalf("expected empty content, got %q", content)
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

func encodeCS12400Payload(id uint32) []byte {
	payload := make([]byte, 0, 8)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(id))
	return payload
}

func encodeCS12402Payload(ids []uint32) []byte {
	payload := make([]byte, 0, len(ids)*4)
	for _, id := range ids {
		payload = protowire.AppendTag(payload, 1, protowire.VarintType)
		payload = protowire.AppendVarint(payload, uint64(id))
	}
	return payload
}

func encodeCS12404Payload(items []orm.LoveLetterConvertedItem) []byte {
	payload := make([]byte, 0, len(items)*8)
	for _, item := range items {
		encoded := encodePTOldLoverItem(item)
		payload = protowire.AppendTag(payload, 1, protowire.BytesType)
		payload = protowire.AppendBytes(payload, encoded)
	}
	return payload
}

func encodeCS12406Payload(requestType uint32) []byte {
	payload := make([]byte, 0, 8)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(requestType))
	return payload
}

func encodeCS12408Payload(groupID uint32) []byte {
	payload := make([]byte, 0, 8)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(groupID))
	return payload
}

func encodeCS12410Payload(letterID uint32) []byte {
	payload := make([]byte, 0, 8)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(letterID))
	return payload
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

func decodeSC12403Payload(payload []byte) (uint32, []*protobuf.DROPINFO, error) {
	result := uint32(0)
	drops := make([]*protobuf.DROPINFO, 0)
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return 0, nil, protowire.ParseError(n)
		}
		payload = payload[n:]
		switch fieldNumber {
		case 1:
			if wireType != protowire.VarintType {
				return 0, nil, errUnexpectedWireType(fieldNumber, wireType)
			}
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return 0, nil, protowire.ParseError(m)
			}
			payload = payload[m:]
			result = uint32(value)
		case 2:
			if wireType != protowire.BytesType {
				return 0, nil, errUnexpectedWireType(fieldNumber, wireType)
			}
			value, m := protowire.ConsumeBytes(payload)
			if m < 0 {
				return 0, nil, protowire.ParseError(m)
			}
			payload = payload[m:]
			drop := &protobuf.DROPINFO{}
			if err := proto.Unmarshal(value, drop); err != nil {
				return 0, nil, err
			}
			drops = append(drops, drop)
		default:
			skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
			if err != nil {
				return 0, nil, err
			}
			payload = payload[skipped:]
		}
	}
	return result, drops, nil
}

func decodeSC12407Payload(payload []byte) (decodedSC12407, error) {
	decoded := decodedSC12407{}
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return decoded, protowire.ParseError(n)
		}
		payload = payload[n:]
		switch fieldNumber {
		case 1:
			if wireType != protowire.BytesType {
				return decoded, errUnexpectedWireType(fieldNumber, wireType)
			}
			value, m := protowire.ConsumeBytes(payload)
			if m < 0 {
				return decoded, protowire.ParseError(m)
			}
			payload = payload[m:]
			item, err := decodePTOldLoverItem(value)
			if err != nil {
				return decoded, err
			}
			decoded.ConvertedItems = append(decoded.ConvertedItems, item)
		case 2:
			if wireType != protowire.VarintType {
				return decoded, errUnexpectedWireType(fieldNumber, wireType)
			}
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return decoded, protowire.ParseError(m)
			}
			payload = payload[m:]
			decoded.RewardedIDs = append(decoded.RewardedIDs, uint32(value))
		case 3:
			if wireType != protowire.BytesType {
				return decoded, errUnexpectedWireType(fieldNumber, wireType)
			}
			value, m := protowire.ConsumeBytes(payload)
			if m < 0 {
				return decoded, protowire.ParseError(m)
			}
			payload = payload[m:]
			medal, err := decodePTLoveLetterMedal(value)
			if err != nil {
				return decoded, err
			}
			decoded.Medals = append(decoded.Medals, medal)
		case 4:
			if wireType != protowire.BytesType {
				return decoded, errUnexpectedWireType(fieldNumber, wireType)
			}
			value, m := protowire.ConsumeBytes(payload)
			if m < 0 {
				return decoded, protowire.ParseError(m)
			}
			payload = payload[m:]
			letters, err := decodePTShipLoveLetter(value)
			if err != nil {
				return decoded, err
			}
			decoded.Letters = append(decoded.Letters, letters)
		case 5:
			if wireType != protowire.BytesType {
				return decoded, errUnexpectedWireType(fieldNumber, wireType)
			}
			value, m := protowire.ConsumeBytes(payload)
			if m < 0 {
				return decoded, protowire.ParseError(m)
			}
			payload = payload[m:]
			letters, err := decodePTShipLoveLetter(value)
			if err != nil {
				return decoded, err
			}
			decoded.ConvertedLetters = append(decoded.ConvertedLetters, letters)
		default:
			skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
			if err != nil {
				return decoded, err
			}
			payload = payload[skipped:]
		}
	}
	return decoded, nil
}

func decodePTLoveLetterMedal(payload []byte) (orm.LoveLetterMedalState, error) {
	medal := orm.LoveLetterMedalState{}
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return medal, protowire.ParseError(n)
		}
		payload = payload[n:]
		switch fieldNumber {
		case 1:
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return medal, protowire.ParseError(m)
			}
			payload = payload[m:]
			medal.GroupID = uint32(value)
		case 2:
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return medal, protowire.ParseError(m)
			}
			payload = payload[m:]
			medal.Exp = uint32(value)
		case 3:
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return medal, protowire.ParseError(m)
			}
			payload = payload[m:]
			medal.Level = uint32(value)
		default:
			skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
			if err != nil {
				return medal, err
			}
			payload = payload[skipped:]
		}
	}
	return medal, nil
}

func decodePTShipLoveLetter(payload []byte) (orm.LoveLetterLetterState, error) {
	letters := orm.LoveLetterLetterState{}
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return letters, protowire.ParseError(n)
		}
		payload = payload[n:]
		switch fieldNumber {
		case 1:
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return letters, protowire.ParseError(m)
			}
			payload = payload[m:]
			letters.GroupID = uint32(value)
		case 2:
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return letters, protowire.ParseError(m)
			}
			payload = payload[m:]
			letters.LetterIDList = append(letters.LetterIDList, uint32(value))
		default:
			skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
			if err != nil {
				return letters, err
			}
			payload = payload[skipped:]
		}
	}
	return letters, nil
}

func decodeSC12411Payload(payload []byte) (string, error) {
	content := ""
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return "", protowire.ParseError(n)
		}
		payload = payload[n:]
		switch fieldNumber {
		case 1:
			if wireType != protowire.BytesType {
				return "", errUnexpectedWireType(fieldNumber, wireType)
			}
			value, m := protowire.ConsumeBytes(payload)
			if m < 0 {
				return "", protowire.ParseError(m)
			}
			payload = payload[m:]
			content = string(value)
		default:
			skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
			if err != nil {
				return "", err
			}
			payload = payload[skipped:]
		}
	}
	return content, nil
}

func hasLetter(states []orm.LoveLetterLetterState, groupID uint32, letterID uint32) bool {
	for _, state := range states {
		if state.GroupID != groupID {
			continue
		}
		for _, id := range state.LetterIDList {
			if id == letterID {
				return true
			}
		}
	}
	return false
}

func errUnexpectedWireType(field protowire.Number, wireType protowire.Type) error {
	return fmt.Errorf("field %d has unexpected wire type %v", field, wireType)
}
