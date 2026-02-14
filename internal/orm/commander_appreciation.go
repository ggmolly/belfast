package orm

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

// CommanderAppreciationState stores commander-scoped Appreciation bitsets.
// These bitsets are surfaced in SC_11003 so the client can rebuild local lists.
type CommanderAppreciationState struct {
	CommanderID        uint32
	MusicNo            uint32
	MusicMode          uint32
	CartoonReadMark    Int64List
	CartoonCollectMark Int64List
	GalleryUnlocks     Int64List
	GalleryFavorIds    Int64List
	MusicFavorIds      Int64List
}

const maxAppreciationMarkBuckets = 4096

func GetOrCreateCommanderAppreciationState(commanderID uint32) (*CommanderAppreciationState, error) {
	ctx := context.Background()
	var out *CommanderAppreciationState
	err := db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		state, err := getOrCreateCommanderAppreciationStateTx(ctx, q, commanderID)
		if err != nil {
			return err
		}
		out = state
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func getOrCreateCommanderAppreciationStateTx(ctx context.Context, q *gen.Queries, commanderID uint32) (*CommanderAppreciationState, error) {
	row, err := q.GetCommanderAppreciationStateByCommanderID(ctx, int64(commanderID))
	err = db.MapNotFound(err)
	if err == nil {
		state := &CommanderAppreciationState{
			CommanderID:        uint32(row.CommanderID),
			MusicNo:            uint32(row.MusicNo),
			MusicMode:          uint32(row.MusicMode),
			CartoonReadMark:    parseInt64List(row.CartoonReadMark),
			CartoonCollectMark: parseInt64List(row.CartoonCollectMark),
			GalleryUnlocks:     parseInt64List(row.GalleryUnlocks),
			GalleryFavorIds:    parseInt64List(row.GalleryFavorIds),
			MusicFavorIds:      parseInt64List(row.MusicFavorIds),
		}
		return state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}

	state := &CommanderAppreciationState{
		CommanderID:        commanderID,
		CartoonReadMark:    Int64List{},
		CartoonCollectMark: Int64List{},
		GalleryUnlocks:     Int64List{},
		GalleryFavorIds:    Int64List{},
		MusicFavorIds:      Int64List{},
	}
	if err := q.CreateCommanderAppreciationState(ctx, gen.CreateCommanderAppreciationStateParams{
		CommanderID:        int64(state.CommanderID),
		MusicNo:            int64(state.MusicNo),
		MusicMode:          int64(state.MusicMode),
		CartoonReadMark:    mustMarshalList(state.CartoonReadMark),
		CartoonCollectMark: mustMarshalList(state.CartoonCollectMark),
		GalleryUnlocks:     mustMarshalList(state.GalleryUnlocks),
		GalleryFavorIds:    mustMarshalList(state.GalleryFavorIds),
		MusicFavorIds:      mustMarshalList(state.MusicFavorIds),
	}); err != nil {
		return nil, err
	}
	return state, nil
}

func SaveCommanderAppreciationState(state *CommanderAppreciationState) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpdateCommanderAppreciationState(ctx, gen.UpdateCommanderAppreciationStateParams{
		CommanderID:        int64(state.CommanderID),
		MusicNo:            int64(state.MusicNo),
		MusicMode:          int64(state.MusicMode),
		CartoonReadMark:    mustMarshalList(state.CartoonReadMark),
		CartoonCollectMark: mustMarshalList(state.CartoonCollectMark),
		GalleryUnlocks:     mustMarshalList(state.GalleryUnlocks),
		GalleryFavorIds:    mustMarshalList(state.GalleryFavorIds),
		MusicFavorIds:      mustMarshalList(state.MusicFavorIds),
	})
}

func SetCommanderCartoonReadMark(commanderID uint32, cartoonID uint32) error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		state, err := getOrCreateCommanderAppreciationStateTx(ctx, q, commanderID)
		if err != nil {
			return err
		}
		marks := ToUint32List(state.CartoonReadMark)
		marks = updateBitsetMark(marks, cartoonID, true)
		state.CartoonReadMark = ToInt64List(marks)
		return q.UpdateCommanderAppreciationState(ctx, gen.UpdateCommanderAppreciationStateParams{
			CommanderID:        int64(state.CommanderID),
			MusicNo:            int64(state.MusicNo),
			MusicMode:          int64(state.MusicMode),
			CartoonReadMark:    mustMarshalList(state.CartoonReadMark),
			CartoonCollectMark: mustMarshalList(state.CartoonCollectMark),
			GalleryUnlocks:     mustMarshalList(state.GalleryUnlocks),
			GalleryFavorIds:    mustMarshalList(state.GalleryFavorIds),
			MusicFavorIds:      mustMarshalList(state.MusicFavorIds),
		})
	})
}

func SetCommanderCartoonCollectMark(commanderID uint32, cartoonID uint32, liked bool) error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		state, err := getOrCreateCommanderAppreciationStateTx(ctx, q, commanderID)
		if err != nil {
			return err
		}
		marks := ToUint32List(state.CartoonCollectMark)
		marks = updateBitsetMark(marks, cartoonID, liked)
		state.CartoonCollectMark = ToInt64List(marks)
		return q.UpdateCommanderAppreciationState(ctx, gen.UpdateCommanderAppreciationStateParams{
			CommanderID:        int64(state.CommanderID),
			MusicNo:            int64(state.MusicNo),
			MusicMode:          int64(state.MusicMode),
			CartoonReadMark:    mustMarshalList(state.CartoonReadMark),
			CartoonCollectMark: mustMarshalList(state.CartoonCollectMark),
			GalleryUnlocks:     mustMarshalList(state.GalleryUnlocks),
			GalleryFavorIds:    mustMarshalList(state.GalleryFavorIds),
			MusicFavorIds:      mustMarshalList(state.MusicFavorIds),
		})
	})
}

func SetCommanderAppreciationGalleryUnlock(commanderID uint32, galleryID uint32) error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		state, err := getOrCreateCommanderAppreciationStateTx(ctx, q, commanderID)
		if err != nil {
			return err
		}
		unlocks := ToUint32List(state.GalleryUnlocks)
		unlocks = updateBitsetMark(unlocks, galleryID, true)
		state.GalleryUnlocks = ToInt64List(unlocks)
		return q.UpdateCommanderAppreciationState(ctx, gen.UpdateCommanderAppreciationStateParams{
			CommanderID:        int64(state.CommanderID),
			MusicNo:            int64(state.MusicNo),
			MusicMode:          int64(state.MusicMode),
			CartoonReadMark:    mustMarshalList(state.CartoonReadMark),
			CartoonCollectMark: mustMarshalList(state.CartoonCollectMark),
			GalleryUnlocks:     mustMarshalList(state.GalleryUnlocks),
			GalleryFavorIds:    mustMarshalList(state.GalleryFavorIds),
			MusicFavorIds:      mustMarshalList(state.MusicFavorIds),
		})
	})
}

const (
	appreciationGalleryConfigCategory = "ShareCfg/gallery_config.json"
	appreciationMusicConfigCategory   = "ShareCfg/music_collect_config.json"
)

func SetCommanderAppreciationGalleryFavor(commanderID uint32, galleryID uint32, liked bool) error {
	if galleryID == 0 {
		return nil
	}
	known, err := configIDExists(appreciationGalleryConfigCategory, galleryID)
	if err != nil {
		return err
	}
	if !known {
		return nil
	}
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		state, err := getOrCreateCommanderAppreciationStateTx(ctx, q, commanderID)
		if err != nil {
			return err
		}
		ids := ToUint32List(state.GalleryFavorIds)
		ids = updateFavorIDList(ids, galleryID, liked)
		state.GalleryFavorIds = ToInt64List(ids)
		return q.UpdateCommanderAppreciationState(ctx, gen.UpdateCommanderAppreciationStateParams{
			CommanderID:        int64(state.CommanderID),
			MusicNo:            int64(state.MusicNo),
			MusicMode:          int64(state.MusicMode),
			CartoonReadMark:    mustMarshalList(state.CartoonReadMark),
			CartoonCollectMark: mustMarshalList(state.CartoonCollectMark),
			GalleryUnlocks:     mustMarshalList(state.GalleryUnlocks),
			GalleryFavorIds:    mustMarshalList(state.GalleryFavorIds),
			MusicFavorIds:      mustMarshalList(state.MusicFavorIds),
		})
	})
}

func SetCommanderAppreciationMusicFavor(commanderID uint32, musicID uint32, liked bool) error {
	if musicID == 0 {
		return nil
	}
	known, err := configIDExists(appreciationMusicConfigCategory, musicID)
	if err != nil {
		return err
	}
	if !known {
		return nil
	}
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		state, err := getOrCreateCommanderAppreciationStateTx(ctx, q, commanderID)
		if err != nil {
			return err
		}
		ids := ToUint32List(state.MusicFavorIds)
		ids = updateFavorIDList(ids, musicID, liked)
		state.MusicFavorIds = ToInt64List(ids)
		return q.UpdateCommanderAppreciationState(ctx, gen.UpdateCommanderAppreciationStateParams{
			CommanderID:        int64(state.CommanderID),
			MusicNo:            int64(state.MusicNo),
			MusicMode:          int64(state.MusicMode),
			CartoonReadMark:    mustMarshalList(state.CartoonReadMark),
			CartoonCollectMark: mustMarshalList(state.CartoonCollectMark),
			GalleryUnlocks:     mustMarshalList(state.GalleryUnlocks),
			GalleryFavorIds:    mustMarshalList(state.GalleryFavorIds),
			MusicFavorIds:      mustMarshalList(state.MusicFavorIds),
		})
	})
}

func configIDExists(category string, id uint32) (bool, error) {
	_, err := GetConfigEntry(category, strconv.FormatUint(uint64(id), 10))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, db.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func updateFavorIDList(ids []uint32, id uint32, enabled bool) []uint32 {
	if id == 0 {
		return ids
	}
	if enabled {
		for _, existing := range ids {
			if existing == id {
				return ids
			}
		}
		return append(ids, id)
	}
	out := ids[:0]
	for _, existing := range ids {
		if existing == id {
			continue
		}
		out = append(out, existing)
	}
	return out
}

func updateBitsetMark(marks []uint32, id uint32, enabled bool) []uint32 {
	if id == 0 {
		return marks
	}
	bucket := int((id - 1) / 32)
	bit := uint((id - 1) % 32)
	if bucket >= maxAppreciationMarkBuckets {
		return marks
	}
	if !enabled && len(marks) < bucket+1 {
		return marks
	}
	if len(marks) < bucket+1 {
		extended := make([]uint32, bucket+1)
		copy(extended, marks)
		marks = extended
	}
	if enabled {
		marks[bucket] |= (1 << bit)
		return marks
	}
	marks[bucket] &^= (1 << bit)
	return marks
}

func parseInt64List(value string) Int64List {
	var out Int64List
	if err := out.Scan(value); err != nil {
		return Int64List{}
	}
	if out == nil {
		return Int64List{}
	}
	return out
}

func mustMarshalList(list Int64List) string {
	payload, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}
	return string(payload)
}
