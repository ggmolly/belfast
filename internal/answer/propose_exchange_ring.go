package answer

import (
	"encoding/json"
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	vowPropConversionCategory = "ShareCfg/gameset.json"
	vowPropConversionKey      = "vow_prop_conversion"
)

type vowPropConversionConfig struct {
	Description []uint32 `json:"description"`
}

func loadVowPropConversionPair(db *gorm.DB) (uint32, uint32, error) {
	entry, err := orm.GetConfigEntry(db, vowPropConversionCategory, vowPropConversionKey)
	if err != nil {
		return 0, 0, err
	}
	var parsed vowPropConversionConfig
	if err := json.Unmarshal(entry.Data, &parsed); err != nil {
		return 0, 0, err
	}
	if len(parsed.Description) != 2 {
		return 0, 0, errors.New("vow_prop_conversion.description must be length 2")
	}
	return parsed.Description[0], parsed.Description[1], nil
}

// ProposeExchangeRing handles CS_15010 by consuming 1x ring and granting 1x tiara
// based on gameset.vow_prop_conversion, responding with SC_15011.
func ProposeExchangeRing(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_15010
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 15011, err
	}
	_ = payload.GetId()

	response := protobuf.SC_15011{Result: proto.Uint32(1)}

	fromItemID, toItemID, err := loadVowPropConversionPair(orm.GormDB)
	if err != nil {
		logger.LogEvent("Config", "Missing", vowPropConversionKey, logger.LOG_LEVEL_ERROR)
		return client.SendMessage(15011, &response)
	}

	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := client.Commander.ConsumeItemTx(tx, fromItemID, 1); err != nil {
			return err
		}
		if err := client.Commander.AddItemTx(tx, toItemID, 1); err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.LogEvent("Item", "Convert", "vow_prop_conversion failed", logger.LOG_LEVEL_ERROR)
		return client.SendMessage(15011, &response)
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(15011, &response)
}
