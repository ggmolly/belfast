package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/debug"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

type loveLetterSnapshot struct {
	ConvertedItems   []orm.LoveLetterConvertedItem
	RewardedIDs      []uint32
	Medals           []orm.LoveLetterMedalState
	Letters          []orm.LoveLetterLetterState
	ConvertedLetters []orm.LoveLetterLetterState
}

func decodeCS12400(payload []byte) (uint32, error) {
	value, ok, err := decodeSingleVarintField(payload, 1)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("missing required field id")
	}
	return value, nil
}

func decodeCS12402(payload []byte) ([]uint32, error) {
	values := make([]uint32, 0)
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return nil, protowire.ParseError(n)
		}
		payload = payload[n:]
		switch fieldNumber {
		case 1:
			if wireType != protowire.VarintType {
				return nil, fmt.Errorf("field id_list has unexpected wire type %v", wireType)
			}
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return nil, protowire.ParseError(m)
			}
			payload = payload[m:]
			values = append(values, uint32(value))
		default:
			skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
			if err != nil {
				return nil, err
			}
			payload = payload[skipped:]
		}
	}
	return values, nil
}

func decodeCS12404(payload []byte) ([]orm.LoveLetterConvertedItem, error) {
	items := make([]orm.LoveLetterConvertedItem, 0)
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return nil, protowire.ParseError(n)
		}
		payload = payload[n:]
		switch fieldNumber {
		case 1:
			if wireType != protowire.BytesType {
				return nil, fmt.Errorf("field item_list has unexpected wire type %v", wireType)
			}
			itemPayload, m := protowire.ConsumeBytes(payload)
			if m < 0 {
				return nil, protowire.ParseError(m)
			}
			payload = payload[m:]
			item, err := decodePTOldLoverItem(itemPayload)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		default:
			skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
			if err != nil {
				return nil, err
			}
			payload = payload[skipped:]
		}
	}
	return items, nil
}

func decodeCS12406(payload []byte) (uint32, error) {
	value, ok, err := decodeSingleVarintField(payload, 1)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("missing required field type")
	}
	return value, nil
}

func decodeCS12408(payload []byte) (uint32, error) {
	value, ok, err := decodeSingleVarintField(payload, 1)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("missing required field group_id")
	}
	return value, nil
}

func decodeCS12410(payload []byte) (uint32, error) {
	value, ok, err := decodeSingleVarintField(payload, 1)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("missing required field letter_id")
	}
	return value, nil
}

func decodeSingleVarintField(payload []byte, targetField protowire.Number) (uint32, bool, error) {
	var value uint32
	var found bool
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return 0, false, protowire.ParseError(n)
		}
		payload = payload[n:]
		if fieldNumber == targetField {
			if wireType != protowire.VarintType {
				return 0, false, fmt.Errorf("field %d has unexpected wire type %v", targetField, wireType)
			}
			decoded, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return 0, false, protowire.ParseError(m)
			}
			payload = payload[m:]
			value = uint32(decoded)
			found = true
			continue
		}
		skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
		if err != nil {
			return 0, false, err
		}
		payload = payload[skipped:]
	}
	return value, found, nil
}

func decodePTOldLoverItem(payload []byte) (orm.LoveLetterConvertedItem, error) {
	item := orm.LoveLetterConvertedItem{}
	var hasItemID bool
	var hasGroupID bool
	var hasYear bool
	for len(payload) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(payload)
		if n < 0 {
			return item, protowire.ParseError(n)
		}
		payload = payload[n:]
		switch fieldNumber {
		case 1:
			if wireType != protowire.VarintType {
				return item, fmt.Errorf("field item_id has unexpected wire type %v", wireType)
			}
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return item, protowire.ParseError(m)
			}
			payload = payload[m:]
			item.ItemID = uint32(value)
			hasItemID = true
		case 2:
			if wireType != protowire.VarintType {
				return item, fmt.Errorf("field group_id has unexpected wire type %v", wireType)
			}
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return item, protowire.ParseError(m)
			}
			payload = payload[m:]
			item.GroupID = uint32(value)
			hasGroupID = true
		case 3:
			if wireType != protowire.VarintType {
				return item, fmt.Errorf("field year has unexpected wire type %v", wireType)
			}
			value, m := protowire.ConsumeVarint(payload)
			if m < 0 {
				return item, protowire.ParseError(m)
			}
			payload = payload[m:]
			item.Year = uint32(value)
			hasYear = true
		default:
			skipped, err := skipLoveLetterField(fieldNumber, wireType, payload)
			if err != nil {
				return item, err
			}
			payload = payload[skipped:]
		}
	}
	if !hasItemID || !hasGroupID || !hasYear {
		return item, fmt.Errorf("missing required converted item fields")
	}
	return item, nil
}

func encodeSC12401(result uint32) []byte {
	payload := make([]byte, 0, 8)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(result))
	return payload
}

func encodeSC12403(result uint32, drops []*protobuf.DROPINFO) ([]byte, error) {
	payload := make([]byte, 0, 16)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(result))
	for _, drop := range drops {
		encodedDrop, err := proto.Marshal(drop)
		if err != nil {
			return nil, err
		}
		payload = protowire.AppendTag(payload, 2, protowire.BytesType)
		payload = protowire.AppendBytes(payload, encodedDrop)
	}
	return payload, nil
}

func encodeSC12405(result uint32) []byte {
	payload := make([]byte, 0, 8)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(result))
	return payload
}

func encodeSC12407(snapshot loveLetterSnapshot) []byte {
	payload := make([]byte, 0, 64)
	for _, item := range snapshot.ConvertedItems {
		encodedItem := encodePTOldLoverItem(item)
		payload = protowire.AppendTag(payload, 1, protowire.BytesType)
		payload = protowire.AppendBytes(payload, encodedItem)
	}
	for _, rewardedID := range snapshot.RewardedIDs {
		payload = protowire.AppendTag(payload, 2, protowire.VarintType)
		payload = protowire.AppendVarint(payload, uint64(rewardedID))
	}
	for _, medal := range snapshot.Medals {
		encodedMedal := encodePTLoveLetterMedal(medal)
		payload = protowire.AppendTag(payload, 3, protowire.BytesType)
		payload = protowire.AppendBytes(payload, encodedMedal)
	}
	for _, letter := range snapshot.Letters {
		encodedLetter := encodePTShipLoveLetter(letter)
		payload = protowire.AppendTag(payload, 4, protowire.BytesType)
		payload = protowire.AppendBytes(payload, encodedLetter)
	}
	for _, letter := range snapshot.ConvertedLetters {
		encodedLetter := encodePTShipLoveLetter(letter)
		payload = protowire.AppendTag(payload, 5, protowire.BytesType)
		payload = protowire.AppendBytes(payload, encodedLetter)
	}
	return payload
}

func encodeSC12409(result uint32) []byte {
	payload := make([]byte, 0, 8)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(result))
	return payload
}

func encodeSC12411(content string) []byte {
	payload := make([]byte, 0, len(content)+8)
	payload = protowire.AppendTag(payload, 1, protowire.BytesType)
	payload = protowire.AppendString(payload, content)
	return payload
}

func encodePTOldLoverItem(item orm.LoveLetterConvertedItem) []byte {
	payload := make([]byte, 0, 16)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(item.ItemID))
	payload = protowire.AppendTag(payload, 2, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(item.GroupID))
	payload = protowire.AppendTag(payload, 3, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(item.Year))
	return payload
}

func encodePTLoveLetterMedal(medal orm.LoveLetterMedalState) []byte {
	payload := make([]byte, 0, 16)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(medal.GroupID))
	payload = protowire.AppendTag(payload, 2, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(medal.Exp))
	payload = protowire.AppendTag(payload, 3, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(medal.Level))
	return payload
}

func encodePTShipLoveLetter(letter orm.LoveLetterLetterState) []byte {
	payload := make([]byte, 0, 16)
	payload = protowire.AppendTag(payload, 1, protowire.VarintType)
	payload = protowire.AppendVarint(payload, uint64(letter.GroupID))
	for _, letterID := range letter.LetterIDList {
		payload = protowire.AppendTag(payload, 2, protowire.VarintType)
		payload = protowire.AppendVarint(payload, uint64(letterID))
	}
	return payload
}

func skipLoveLetterField(fieldNumber protowire.Number, wireType protowire.Type, payload []byte) (int, error) {
	switch wireType {
	case protowire.VarintType:
		_, n := protowire.ConsumeVarint(payload)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	case protowire.Fixed32Type:
		_, n := protowire.ConsumeFixed32(payload)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	case protowire.Fixed64Type:
		_, n := protowire.ConsumeFixed64(payload)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	case protowire.BytesType:
		_, n := protowire.ConsumeBytes(payload)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	case protowire.StartGroupType:
		_, n := protowire.ConsumeGroup(fieldNumber, payload)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("unsupported wire type %v", wireType)
	}
}

func sendRawPacket(packetID int, payload []byte, client *connection.Client) (int, int, error) {
	packet := append([]byte{}, payload...)
	debug.InsertPacket(packetID, &packet)
	connection.InjectPacketHeader(packetID, &packet, client.PacketIndex)
	n, err := client.Buffer.Write(packet)
	if err != nil {
		logger.LogEvent("Connection", "Buffer", fmt.Sprintf("SC_%d -> %v", packetID, err), logger.LOG_LEVEL_ERROR)
		client.RecordHandlerError()
		client.CloseWithError(err)
		return n, packetID, err
	}
	logger.LogEvent("Connection", "SendMessage", fmt.Sprintf("SC_%d - %d bytes buffered", packetID, n), logger.LOG_LEVEL_DEBUG)
	return n, packetID, nil
}
