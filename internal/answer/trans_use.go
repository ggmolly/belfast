package answer

import (
	"encoding/json"
	"errors"
)

func addTransUseItems(dst map[uint32]uint32, raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var pairs [][]uint32
	if err := json.Unmarshal(raw, &pairs); err != nil {
		return err
	}
	for _, pair := range pairs {
		if len(pair) != 2 {
			return errors.New("invalid trans_use_item")
		}
		if pair[0] == 0 || pair[1] == 0 {
			continue
		}
		dst[pair[0]] += pair[1]
	}
	return nil
}
