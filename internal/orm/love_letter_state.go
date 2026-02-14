package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type LoveLetterMedalState struct {
	GroupID uint32 `json:"group_id"`
	Exp     uint32 `json:"exp"`
	Level   uint32 `json:"level"`
}

type LoveLetterLetterState struct {
	GroupID      uint32   `json:"group_id"`
	LetterIDList []uint32 `json:"letter_id_list"`
}

type LoveLetterConvertedItem struct {
	ItemID  uint32 `json:"item_id"`
	GroupID uint32 `json:"group_id"`
	Year    uint32 `json:"year"`
}

type CommanderLoveLetterState struct {
	CommanderID    uint32
	Medals         []LoveLetterMedalState
	ManualLetters  []LoveLetterLetterState
	ConvertedItems []LoveLetterConvertedItem
	RewardedIDs    []uint32
	LetterContents map[uint32]string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func GetCommanderLoveLetterState(commanderID uint32) (*CommanderLoveLetterState, error) {
	ctx := context.Background()
	state := &CommanderLoveLetterState{}
	var medalsRaw []byte
	var manualLettersRaw []byte
	var convertedItemsRaw []byte
	var rewardedIDsRaw []byte
	var letterContentsRaw []byte
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, medals, manual_letters, converted_items, rewarded_ids, letter_contents, created_at, updated_at
FROM commander_love_letter_states
WHERE commander_id = $1
`, int64(commanderID)).Scan(
		&state.CommanderID,
		&medalsRaw,
		&manualLettersRaw,
		&convertedItemsRaw,
		&rewardedIDsRaw,
		&letterContentsRaw,
		&state.CreatedAt,
		&state.UpdatedAt,
	)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	if err := unmarshalLoveLetterState(state, medalsRaw, manualLettersRaw, convertedItemsRaw, rewardedIDsRaw, letterContentsRaw); err != nil {
		return nil, err
	}
	return state, nil
}

func GetOrCreateCommanderLoveLetterState(commanderID uint32) (*CommanderLoveLetterState, error) {
	state, err := GetCommanderLoveLetterState(commanderID)
	if err == nil {
		return state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}
	state = &CommanderLoveLetterState{
		CommanderID:    commanderID,
		Medals:         []LoveLetterMedalState{},
		ManualLetters:  []LoveLetterLetterState{},
		ConvertedItems: []LoveLetterConvertedItem{},
		RewardedIDs:    []uint32{},
		LetterContents: map[uint32]string{},
	}
	if err := SaveCommanderLoveLetterState(state); err != nil {
		return nil, err
	}
	return state, nil
}

func SaveCommanderLoveLetterState(state *CommanderLoveLetterState) error {
	ctx := context.Background()
	return saveCommanderLoveLetterStateWithExec(ctx, db.DefaultStore.Pool, state)
}

func SaveCommanderLoveLetterStateTx(ctx context.Context, tx pgx.Tx, state *CommanderLoveLetterState) error {
	return saveCommanderLoveLetterStateWithExec(ctx, tx, state)
}

type loveLetterStateExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func saveCommanderLoveLetterStateWithExec(ctx context.Context, executor loveLetterStateExecutor, state *CommanderLoveLetterState) error {
	medalsRaw, err := json.Marshal(state.Medals)
	if err != nil {
		return err
	}
	manualLettersRaw, err := json.Marshal(state.ManualLetters)
	if err != nil {
		return err
	}
	convertedItemsRaw, err := json.Marshal(state.ConvertedItems)
	if err != nil {
		return err
	}
	rewardedIDsRaw, err := json.Marshal(state.RewardedIDs)
	if err != nil {
		return err
	}
	letterContentsRaw, err := marshalLoveLetterContents(state.LetterContents)
	if err != nil {
		return err
	}
	_, err = executor.Exec(ctx, `
INSERT INTO commander_love_letter_states (commander_id, medals, manual_letters, converted_items, rewarded_ids, letter_contents, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (commander_id)
DO UPDATE SET
  medals = EXCLUDED.medals,
  manual_letters = EXCLUDED.manual_letters,
  converted_items = EXCLUDED.converted_items,
  rewarded_ids = EXCLUDED.rewarded_ids,
  letter_contents = EXCLUDED.letter_contents,
  updated_at = NOW()
`, int64(state.CommanderID), medalsRaw, manualLettersRaw, convertedItemsRaw, rewardedIDsRaw, letterContentsRaw)
	return err
}

func DeleteCommanderLoveLetterState(commanderID uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM commander_love_letter_states
WHERE commander_id = $1
`, int64(commanderID))
	return err
}

func unmarshalLoveLetterState(
	state *CommanderLoveLetterState,
	medalsRaw []byte,
	manualLettersRaw []byte,
	convertedItemsRaw []byte,
	rewardedIDsRaw []byte,
	letterContentsRaw []byte,
) error {
	if len(medalsRaw) == 0 {
		state.Medals = []LoveLetterMedalState{}
	} else if err := json.Unmarshal(medalsRaw, &state.Medals); err != nil {
		return err
	}
	if len(manualLettersRaw) == 0 {
		state.ManualLetters = []LoveLetterLetterState{}
	} else if err := json.Unmarshal(manualLettersRaw, &state.ManualLetters); err != nil {
		return err
	}
	if len(convertedItemsRaw) == 0 {
		state.ConvertedItems = []LoveLetterConvertedItem{}
	} else if err := json.Unmarshal(convertedItemsRaw, &state.ConvertedItems); err != nil {
		return err
	}
	if len(rewardedIDsRaw) == 0 {
		state.RewardedIDs = []uint32{}
	} else if err := json.Unmarshal(rewardedIDsRaw, &state.RewardedIDs); err != nil {
		return err
	}
	contents, err := unmarshalLoveLetterContents(letterContentsRaw)
	if err != nil {
		return err
	}
	state.LetterContents = contents
	return nil
}

func marshalLoveLetterContents(contents map[uint32]string) ([]byte, error) {
	if len(contents) == 0 {
		return []byte("{}"), nil
	}
	encoded := make(map[string]string, len(contents))
	for letterID, content := range contents {
		encoded[strconv.FormatUint(uint64(letterID), 10)] = content
	}
	return json.Marshal(encoded)
}

func unmarshalLoveLetterContents(raw []byte) (map[uint32]string, error) {
	if len(raw) == 0 {
		return map[uint32]string{}, nil
	}
	decoded := make(map[string]string)
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	contents := make(map[uint32]string, len(decoded))
	for letterIDRaw, content := range decoded {
		letterID, err := strconv.ParseUint(letterIDRaw, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid letter id %q: %w", letterIDRaw, err)
		}
		contents[uint32(letterID)] = content
	}
	return contents, nil
}
