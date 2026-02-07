package answer

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

type furnitureTemplate struct {
	ID        uint32   `json:"id"`
	Type      uint32   `json:"type"`
	Belong    uint32   `json:"belong"`
	Size      []uint32 `json:"size"`
	CanRotate uint32   `json:"can_rotate"`
	Count     uint32   `json:"count"`
	Name      string   `json:"name"`
}

type dormMapSize struct {
	MinX uint32
	MinY uint32
	MaxX uint32
	MaxY uint32
}

func dormStaticMapSize(level uint32) dormMapSize {
	min := uint32(12)
	if level > 1 {
		min = uint32(12 - (level-1)*4)
	}
	// Client constant: BackYardConst.MAX_MAP_SIZE = Vector2(23,23)
	return dormMapSize{MinX: min, MinY: min, MaxX: 23, MaxY: 23}
}

func loadFurnitureTemplate(id uint32) (*furnitureTemplate, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/furniture_data_template.json", fmt.Sprintf("%d", id))
	if err != nil {
		return nil, err
	}
	var tpl furnitureTemplate
	if err := json.Unmarshal(entry.Data, &tpl); err != nil {
		return nil, err
	}
	if tpl.ID == 0 {
		tpl.ID = id
	}
	return &tpl, nil
}

func resolveFurnitureTemplateID(rawID uint32) (uint32, error) {
	// Raw ids can be in several forms in client code. Prefer direct match.
	if _, err := loadFurnitureTemplate(rawID); err == nil {
		return rawID, nil
	}
	// Try uniqueId form: configId*100 + idx
	if rawID >= 100 {
		base := rawID / 100
		if _, err := loadFurnitureTemplate(base); err == nil {
			return base, nil
		}
	}
	// Try "template+idx" form used by convertor when count > (raw-template)
	for i := uint32(0); i < 100; i++ {
		if rawID < i {
			break
		}
		base := rawID - i
		tpl, err := loadFurnitureTemplate(base)
		if err != nil {
			continue
		}
		if tpl.Count > i {
			return base, nil
		}
	}
	// Try "template*10000000+idx" form
	if rawID > 10000000 {
		base := rawID / 10000000
		idx := rawID % 10
		tpl, err := loadFurnitureTemplate(base)
		if err == nil && tpl.Count > idx {
			return base, nil
		}
	}
	return 0, fmt.Errorf("unknown furniture template id %d", rawID)
}

func matOrPaper(tpl *furnitureTemplate) bool {
	// Client: type==5 (mat) or type==10 (wall mat) or type==1 (wallpaper) or type==4 (floorpaper)
	switch tpl.Type {
	case 1, 4, 5, 10:
		return true
	default:
		return false
	}
}

type rawFurniture struct {
	RawID   uint32
	TplID   uint32
	Tpl     *furnitureTemplate
	X       uint32
	Y       uint32
	Dir     uint32
	Parent  uint64
	ChildID []uint32
}

func validateFurniturePutList(list []*protobuf.FURNITUREPUTINFO, floor uint32, mapSize dormMapSize) error {
	// Build raw entries and an index by raw id.
	index := make(map[uint32]*rawFurniture, len(list))
	items := make([]*rawFurniture, 0, len(list))
	for _, f := range list {
		rawID64, err := strconv.ParseUint(f.GetId(), 10, 32)
		if err != nil {
			return fmt.Errorf("invalid furniture id %q", f.GetId())
		}
		rawID := uint32(rawID64)
		tplID, err := resolveFurnitureTemplateID(rawID)
		if err != nil {
			return err
		}
		tpl, err := loadFurnitureTemplate(tplID)
		if err != nil {
			return err
		}
		if f.GetDir() > 2 {
			return fmt.Errorf("invalid dir %d", f.GetDir())
		}
		childIDs := make([]uint32, 0, len(f.GetChild()))
		for _, c := range f.GetChild() {
			cid64, err := strconv.ParseUint(c.GetId(), 10, 32)
			if err != nil {
				return fmt.Errorf("invalid child id %q", c.GetId())
			}
			childIDs = append(childIDs, uint32(cid64))
		}
		item := &rawFurniture{
			RawID:   rawID,
			TplID:   tplID,
			Tpl:     tpl,
			X:       f.GetX(),
			Y:       f.GetY(),
			Dir:     f.GetDir(),
			Parent:  f.GetParent(),
			ChildID: childIDs,
		}
		items = append(items, item)
		index[rawID] = item
	}

	// FillMap: only for floor-plane, non-mat/paper, no parent.
	occupied := map[uint32]map[uint32]bool{}
	for _, item := range items {
		if item.Tpl.Belong != 1 {
			continue
		}
		if matOrPaper(item.Tpl) {
			continue
		}
		if item.Parent != 0 {
			continue
		}
		sizeX, sizeY := footprint(item.Tpl, item.Dir)
		for x := item.X; x < item.X+sizeX; x++ {
			col := occupied[x]
			if col == nil {
				col = map[uint32]bool{}
				occupied[x] = col
			}
			for y := item.Y; y < item.Y+sizeY; y++ {
				if col[y] {
					return fmt.Errorf("incorrect position")
				}
				col[y] = true
			}
		}
	}

	// Check each furniture.
	for _, item := range items {
		if floor == 0 {
			return fmt.Errorf("floor should exist")
		}
		// parent -> child relation
		if item.Parent != 0 {
			parent := index[uint32(item.Parent)]
			if parent == nil {
				return fmt.Errorf("incorrect [parent -> child] relation")
			}
			found := false
			for _, cid := range parent.ChildID {
				if cid == item.RawID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("incorrect [parent -> child] relation")
			}
		}
		// child -> parent relation
		for _, cid := range item.ChildID {
			child := index[cid]
			if child == nil {
				return fmt.Errorf("incorrect [child -> parent] relation")
			}
			if child.Parent != uint64(item.RawID) {
				return fmt.Errorf("incorrect [child -> parent] relation")
			}
		}
		// in bounds
		if item.Tpl.Belong == 1 && item.Tpl.Type != 1 && item.Tpl.Type != 4 && item.Parent == 0 {
			sizeX, sizeY := footprint(item.Tpl, item.Dir)
			for x := item.X; x < item.X+sizeX; x++ {
				for y := item.Y; y < item.Y+sizeY; y++ {
					if x < mapSize.MinX || y < mapSize.MinY || x > mapSize.MaxX || y > mapSize.MaxY {
						return fmt.Errorf("out side")
					}
				}
			}
		}
		if item.Tpl.Belong == 3 && item.X >= mapSize.MaxX+1 {
			return fmt.Errorf("out side")
		}
		if item.Tpl.Belong == 4 && item.Y >= mapSize.MaxY+1 {
			return fmt.Errorf("out side")
		}
	}

	return nil
}

func footprint(tpl *furnitureTemplate, dir uint32) (uint32, uint32) {
	// dir==1 means no rotation, else swap (client logic)
	if len(tpl.Size) < 2 {
		return 0, 0
	}
	if dir == 1 {
		return tpl.Size[0], tpl.Size[1]
	}
	return tpl.Size[1], tpl.Size[0]
}
