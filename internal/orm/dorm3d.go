package orm

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

const (
	Dorm3dInstagramOpRead  uint32 = 3
	Dorm3dInstagramOpLike  uint32 = 4
	Dorm3dInstagramOpShare uint32 = 5
	Dorm3dInstagramOpExit  uint32 = 6
)

type Dorm3dApartment struct {
	CommanderID        uint32             `gorm:"primary_key" json:"commander_id"`
	DailyVigorMax      uint32             `gorm:"not_null;default:0" json:"daily_vigor_max"`
	Gifts              Dorm3dGiftList     `gorm:"type:text;not_null;default:'[]'" json:"gifts"`
	Ships              Dorm3dShipList     `gorm:"type:text;not_null;default:'[]'" json:"ships"`
	GiftDaily          Dorm3dGiftShopList `gorm:"type:text;not_null;default:'[]'" json:"gift_daily"`
	GiftPermanent      Dorm3dGiftShopList `gorm:"type:text;not_null;default:'[]'" json:"gift_permanent"`
	FurnitureDaily     Dorm3dGiftShopList `gorm:"type:text;not_null;default:'[]'" json:"furniture_daily"`
	FurniturePermanent Dorm3dGiftShopList `gorm:"type:text;not_null;default:'[]'" json:"furniture_permanent"`
	Rooms              Dorm3dRoomList     `gorm:"type:text;not_null;default:'[]'" json:"rooms"`
	Ins                Dorm3dInsList      `gorm:"type:text;not_null;default:'[]'" json:"ins"`
}

type Dorm3dGift struct {
	GiftID     uint32 `json:"gift_id"`
	Number     uint32 `json:"number"`
	UsedNumber uint32 `json:"used_number"`
}

type Dorm3dGiftList []Dorm3dGift

type Dorm3dGiftShop struct {
	GiftID uint32 `json:"gift_id"`
	Count  uint32 `json:"count"`
}

type Dorm3dGiftShopList []Dorm3dGiftShop

type Dorm3dFurniture struct {
	FurnitureID uint32 `json:"furniture_id"`
	SlotID      uint32 `json:"slot_id"`
}

type Dorm3dRoom struct {
	ID          uint32            `json:"id"`
	Furnitures  []Dorm3dFurniture `json:"furnitures"`
	Collections []uint32          `json:"collections"`
	Ships       []uint32          `json:"ships"`
}

type Dorm3dRoomList []Dorm3dRoom

type Dorm3dSkinHiddenInfo struct {
	SkinID      uint32   `json:"skin_id"`
	HiddenParts []uint32 `json:"hidden_parts"`
}

type Dorm3dShip struct {
	ShipGroup      uint32                 `json:"ship_group"`
	FavorLv        uint32                 `json:"favor_lv"`
	FavorExp       uint32                 `json:"favor_exp"`
	RegularTrigger []uint32               `json:"regular_trigger"`
	DailyFavor     uint32                 `json:"daily_favor"`
	Dialogues      []uint32               `json:"dialogues"`
	Skins          []uint32               `json:"skins"`
	CurSkin        uint32                 `json:"cur_skin"`
	Name           string                 `json:"name"`
	NameCd         uint32                 `json:"name_cd"`
	VisitTime      uint32                 `json:"visit_time"`
	HiddenInfo     []Dorm3dSkinHiddenInfo `json:"hidden_info"`
}

type Dorm3dShipList []Dorm3dShip

type Dorm3dKeyValue struct {
	Key   uint32 `json:"key"`
	Value uint32 `json:"value"`
}

type Dorm3dCommInfo struct {
	ID        uint32           `json:"id"`
	Time      uint32           `json:"time"`
	ReadFlag  uint32           `json:"read_flag"`
	ReplyList []Dorm3dKeyValue `json:"reply_list"`
}

type Dorm3dPhoneInfo struct {
	ID       uint32 `json:"id"`
	Time     uint32 `json:"time"`
	ReadFlag uint32 `json:"read_flag"`
}

type Dorm3dReplyFriend struct {
	Key   uint32 `json:"key"`
	Value uint32 `json:"value"`
	Time  uint32 `json:"time"`
}

type Dorm3dFriendCircleInfo struct {
	ID        uint32              `json:"id"`
	Time      uint32              `json:"time"`
	ReadFlag  uint32              `json:"read_flag"`
	GoodFlag  uint32              `json:"good_flag"`
	ReplyList []Dorm3dReplyFriend `json:"reply_list"`
	ExitTime  uint32              `json:"exit_time"`
}

type Dorm3dIns struct {
	ShipGroup  uint32                   `json:"ship_group"`
	CareFlag   uint32                   `json:"care_flag"`
	CurBack    uint32                   `json:"cur_back"`
	CurCommId  uint32                   `json:"cur_comm_id"`
	CommList   []Dorm3dCommInfo         `json:"comm_list"`
	PhoneList  []Dorm3dPhoneInfo        `json:"phone_list"`
	FriendList []Dorm3dFriendCircleInfo `json:"friend_list"`
}

type Dorm3dInsList []Dorm3dIns

func NewDorm3dApartment(commanderID uint32) Dorm3dApartment {
	return Dorm3dApartment{
		CommanderID:        commanderID,
		DailyVigorMax:      0,
		Gifts:              Dorm3dGiftList{},
		Ships:              Dorm3dShipList{},
		GiftDaily:          Dorm3dGiftShopList{},
		GiftPermanent:      Dorm3dGiftShopList{},
		FurnitureDaily:     Dorm3dGiftShopList{},
		FurniturePermanent: Dorm3dGiftShopList{},
		Rooms:              Dorm3dRoomList{},
		Ins:                Dorm3dInsList{},
	}
}

func GetDorm3dApartment(commanderID uint32) (*Dorm3dApartment, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id,
       daily_vigor_max,
       gifts,
       ships,
       gift_daily,
       gift_permanent,
       furniture_daily,
       furniture_permanent,
       rooms,
       ins
FROM dorm3d_apartments
WHERE commander_id = $1
`, int64(commanderID))
	apartment, err := scanDorm3dApartment(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	apartment.EnsureDefaults()
	return &apartment, nil
}

func ListDorm3dApartments(offset int, limit int) ([]Dorm3dApartment, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM dorm3d_apartments`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT commander_id,
       daily_vigor_max,
       gifts,
       ships,
       gift_daily,
       gift_permanent,
       furniture_daily,
       furniture_permanent,
       rooms,
       ins
FROM dorm3d_apartments
ORDER BY commander_id ASC
OFFSET $1
LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	apartments := make([]Dorm3dApartment, 0)
	for rows.Next() {
		apartment, err := scanDorm3dApartment(rows)
		if err != nil {
			return nil, 0, err
		}
		apartment.EnsureDefaults()
		apartments = append(apartments, apartment)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return apartments, total, nil
}

func GetOrCreateDorm3dApartment(commanderID uint32) (*Dorm3dApartment, error) {
	apartment, err := GetDorm3dApartment(commanderID)
	if err == nil {
		return apartment, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}
	ctx := context.Background()
	if err := db.DefaultStore.Queries.CreateDorm3dApartment(ctx, int64(commanderID)); err != nil {
		return nil, err
	}
	return GetDorm3dApartment(commanderID)
}

func SaveDorm3dApartment(apartment *Dorm3dApartment) error {
	apartment.EnsureDefaults()
	ctx := context.Background()
	gifts, err := marshalDorm3dJSONB(apartment.Gifts)
	if err != nil {
		return err
	}
	ships, err := marshalDorm3dJSONB(apartment.Ships)
	if err != nil {
		return err
	}
	giftDaily, err := marshalDorm3dJSONB(apartment.GiftDaily)
	if err != nil {
		return err
	}
	giftPermanent, err := marshalDorm3dJSONB(apartment.GiftPermanent)
	if err != nil {
		return err
	}
	furnitureDaily, err := marshalDorm3dJSONB(apartment.FurnitureDaily)
	if err != nil {
		return err
	}
	furniturePermanent, err := marshalDorm3dJSONB(apartment.FurniturePermanent)
	if err != nil {
		return err
	}
	rooms, err := marshalDorm3dJSONB(apartment.Rooms)
	if err != nil {
		return err
	}
	ins, err := marshalDorm3dJSONB(apartment.Ins)
	if err != nil {
		return err
	}
	return db.DefaultStore.Queries.UpsertDorm3dApartment(ctx, gen.UpsertDorm3dApartmentParams{
		CommanderID:        int64(apartment.CommanderID),
		DailyVigorMax:      int64(apartment.DailyVigorMax),
		Gifts:              gifts,
		Ships:              ships,
		GiftDaily:          giftDaily,
		GiftPermanent:      giftPermanent,
		FurnitureDaily:     furnitureDaily,
		FurniturePermanent: furniturePermanent,
		Rooms:              rooms,
		Ins:                ins,
	})
}

func DeleteDorm3dApartment(commanderID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM dorm3d_apartments WHERE commander_id = $1`, int64(commanderID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func CreateDorm3dApartment(apartment *Dorm3dApartment) error {
	apartment.EnsureDefaults()
	ctx := context.Background()
	gifts, err := marshalDorm3dJSONB(apartment.Gifts)
	if err != nil {
		return err
	}
	ships, err := marshalDorm3dJSONB(apartment.Ships)
	if err != nil {
		return err
	}
	giftDaily, err := marshalDorm3dJSONB(apartment.GiftDaily)
	if err != nil {
		return err
	}
	giftPermanent, err := marshalDorm3dJSONB(apartment.GiftPermanent)
	if err != nil {
		return err
	}
	furnitureDaily, err := marshalDorm3dJSONB(apartment.FurnitureDaily)
	if err != nil {
		return err
	}
	furniturePermanent, err := marshalDorm3dJSONB(apartment.FurniturePermanent)
	if err != nil {
		return err
	}
	rooms, err := marshalDorm3dJSONB(apartment.Rooms)
	if err != nil {
		return err
	}
	ins, err := marshalDorm3dJSONB(apartment.Ins)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO dorm3d_apartments (
	commander_id,
	daily_vigor_max,
	gifts,
	ships,
	gift_daily,
	gift_permanent,
	furniture_daily,
	furniture_permanent,
	rooms,
	ins
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`,
		int64(apartment.CommanderID),
		int64(apartment.DailyVigorMax),
		gifts,
		ships,
		giftDaily,
		giftPermanent,
		furnitureDaily,
		furniturePermanent,
		rooms,
		ins,
	)
	return err
}

func scanDorm3dApartment(scanner rowScanner) (Dorm3dApartment, error) {
	var (
		apartment = Dorm3dApartment{
			Gifts:              Dorm3dGiftList{},
			Ships:              Dorm3dShipList{},
			GiftDaily:          Dorm3dGiftShopList{},
			GiftPermanent:      Dorm3dGiftShopList{},
			FurnitureDaily:     Dorm3dGiftShopList{},
			FurniturePermanent: Dorm3dGiftShopList{},
			Rooms:              Dorm3dRoomList{},
			Ins:                Dorm3dInsList{},
		}
		commanderID       int64
		dailyVigorMax     int64
		giftsPayload      []byte
		shipsPayload      []byte
		giftDailyPayload  []byte
		giftPermPayload   []byte
		furnitureDPayload []byte
		furniturePPayload []byte
		roomsPayload      []byte
		insPayload        []byte
	)
	if err := scanner.Scan(
		&commanderID,
		&dailyVigorMax,
		&giftsPayload,
		&shipsPayload,
		&giftDailyPayload,
		&giftPermPayload,
		&furnitureDPayload,
		&furniturePPayload,
		&roomsPayload,
		&insPayload,
	); err != nil {
		return Dorm3dApartment{}, err
	}
	apartment.CommanderID = uint32(commanderID)
	apartment.DailyVigorMax = uint32(dailyVigorMax)
	if err := unmarshalDorm3dJSONB(giftsPayload, &apartment.Gifts); err != nil {
		return Dorm3dApartment{}, err
	}
	if err := unmarshalDorm3dJSONB(shipsPayload, &apartment.Ships); err != nil {
		return Dorm3dApartment{}, err
	}
	if err := unmarshalDorm3dJSONB(giftDailyPayload, &apartment.GiftDaily); err != nil {
		return Dorm3dApartment{}, err
	}
	if err := unmarshalDorm3dJSONB(giftPermPayload, &apartment.GiftPermanent); err != nil {
		return Dorm3dApartment{}, err
	}
	if err := unmarshalDorm3dJSONB(furnitureDPayload, &apartment.FurnitureDaily); err != nil {
		return Dorm3dApartment{}, err
	}
	if err := unmarshalDorm3dJSONB(furniturePPayload, &apartment.FurniturePermanent); err != nil {
		return Dorm3dApartment{}, err
	}
	if err := unmarshalDorm3dJSONB(roomsPayload, &apartment.Rooms); err != nil {
		return Dorm3dApartment{}, err
	}
	if err := unmarshalDorm3dJSONB(insPayload, &apartment.Ins); err != nil {
		return Dorm3dApartment{}, err
	}
	return apartment, nil
}

func unmarshalDorm3dJSONB(value []byte, target any) error {
	if len(value) == 0 {
		return nil
	}
	return json.Unmarshal(value, target)
}

func marshalDorm3dJSONB(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return []byte("[]"), nil
	}
	return payload, nil
}

func UpdateDorm3dInstagramFlags(commanderID uint32, shipGroup uint32, postIDs []uint32, op uint32, now uint32) error {
	if len(postIDs) == 0 {
		return nil
	}
	apartment, err := GetOrCreateDorm3dApartment(commanderID)
	if err != nil {
		return err
	}
	ins := apartment.ensureInsEntry(shipGroup)
	for _, postID := range postIDs {
		entry := ins.ensureFriendEntry(postID, now)
		switch op {
		case Dorm3dInstagramOpRead:
			entry.ReadFlag = 1
		case Dorm3dInstagramOpLike:
			entry.GoodFlag = 1
		case Dorm3dInstagramOpExit:
			entry.ExitTime = now
		case Dorm3dInstagramOpShare:
			// No state change
		}
	}
	return SaveDorm3dApartment(apartment)
}

func AddDorm3dInstagramReply(commanderID uint32, shipGroup uint32, postID uint32, chatID uint32, value uint32, now uint32) error {
	apartment, err := GetOrCreateDorm3dApartment(commanderID)
	if err != nil {
		return err
	}
	ins := apartment.ensureInsEntry(shipGroup)
	entry := ins.ensureFriendEntry(postID, now)
	entry.ReplyList = append(entry.ReplyList, Dorm3dReplyFriend{
		Key:   chatID,
		Value: value,
		Time:  now,
	})
	return SaveDorm3dApartment(apartment)
}

func (apartment *Dorm3dApartment) EnsureDefaults() {
	if apartment.Gifts == nil {
		apartment.Gifts = Dorm3dGiftList{}
	}
	if apartment.Ships == nil {
		apartment.Ships = Dorm3dShipList{}
	}
	if apartment.GiftDaily == nil {
		apartment.GiftDaily = Dorm3dGiftShopList{}
	}
	if apartment.GiftPermanent == nil {
		apartment.GiftPermanent = Dorm3dGiftShopList{}
	}
	if apartment.FurnitureDaily == nil {
		apartment.FurnitureDaily = Dorm3dGiftShopList{}
	}
	if apartment.FurniturePermanent == nil {
		apartment.FurniturePermanent = Dorm3dGiftShopList{}
	}
	if apartment.Rooms == nil {
		apartment.Rooms = Dorm3dRoomList{}
	}
	if apartment.Ins == nil {
		apartment.Ins = Dorm3dInsList{}
	}
	for i := range apartment.Rooms {
		if apartment.Rooms[i].Furnitures == nil {
			apartment.Rooms[i].Furnitures = []Dorm3dFurniture{}
		}
		if apartment.Rooms[i].Collections == nil {
			apartment.Rooms[i].Collections = []uint32{}
		}
		if apartment.Rooms[i].Ships == nil {
			apartment.Rooms[i].Ships = []uint32{}
		}
	}
	for i := range apartment.Ships {
		if apartment.Ships[i].RegularTrigger == nil {
			apartment.Ships[i].RegularTrigger = []uint32{}
		}
		if apartment.Ships[i].Dialogues == nil {
			apartment.Ships[i].Dialogues = []uint32{}
		}
		if apartment.Ships[i].Skins == nil {
			apartment.Ships[i].Skins = []uint32{}
		}
		if apartment.Ships[i].HiddenInfo == nil {
			apartment.Ships[i].HiddenInfo = []Dorm3dSkinHiddenInfo{}
		}
	}
	for i := range apartment.Ins {
		if apartment.Ins[i].CommList == nil {
			apartment.Ins[i].CommList = []Dorm3dCommInfo{}
		}
		if apartment.Ins[i].PhoneList == nil {
			apartment.Ins[i].PhoneList = []Dorm3dPhoneInfo{}
		}
		if apartment.Ins[i].FriendList == nil {
			apartment.Ins[i].FriendList = []Dorm3dFriendCircleInfo{}
		}
		for j := range apartment.Ins[i].CommList {
			if apartment.Ins[i].CommList[j].ReplyList == nil {
				apartment.Ins[i].CommList[j].ReplyList = []Dorm3dKeyValue{}
			}
		}
		for j := range apartment.Ins[i].FriendList {
			if apartment.Ins[i].FriendList[j].ReplyList == nil {
				apartment.Ins[i].FriendList[j].ReplyList = []Dorm3dReplyFriend{}
			}
		}
	}
}

func (apartment *Dorm3dApartment) ensureInsEntry(shipGroup uint32) *Dorm3dIns {
	for i := range apartment.Ins {
		if apartment.Ins[i].ShipGroup == shipGroup {
			return &apartment.Ins[i]
		}
	}
	newEntry := Dorm3dIns{
		ShipGroup:  shipGroup,
		CommList:   []Dorm3dCommInfo{},
		PhoneList:  []Dorm3dPhoneInfo{},
		FriendList: []Dorm3dFriendCircleInfo{},
	}
	apartment.Ins = append(apartment.Ins, newEntry)
	return &apartment.Ins[len(apartment.Ins)-1]
}

func (ins *Dorm3dIns) ensureFriendEntry(postID uint32, now uint32) *Dorm3dFriendCircleInfo {
	for i := range ins.FriendList {
		if ins.FriendList[i].ID == postID {
			return &ins.FriendList[i]
		}
	}
	entry := Dorm3dFriendCircleInfo{
		ID:        postID,
		Time:      now,
		ReadFlag:  0,
		GoodFlag:  0,
		ReplyList: []Dorm3dReplyFriend{},
		ExitTime:  0,
	}
	ins.FriendList = append(ins.FriendList, entry)
	return &ins.FriendList[len(ins.FriendList)-1]
}

func (list Dorm3dGiftList) Value() (driver.Value, error) {
	return marshalDorm3dJSON(list)
}

func (list *Dorm3dGiftList) Scan(value any) error {
	return scanDorm3dJSON(value, list)
}

func (list Dorm3dGiftShopList) Value() (driver.Value, error) {
	return marshalDorm3dJSON(list)
}

func (list *Dorm3dGiftShopList) Scan(value any) error {
	return scanDorm3dJSON(value, list)
}

func (list Dorm3dRoomList) Value() (driver.Value, error) {
	return marshalDorm3dJSON(list)
}

func (list *Dorm3dRoomList) Scan(value any) error {
	return scanDorm3dJSON(value, list)
}

func (list Dorm3dShipList) Value() (driver.Value, error) {
	return marshalDorm3dJSON(list)
}

func (list *Dorm3dShipList) Scan(value any) error {
	return scanDorm3dJSON(value, list)
}

func (list Dorm3dInsList) Value() (driver.Value, error) {
	return marshalDorm3dJSON(list)
}

func (list *Dorm3dInsList) Scan(value any) error {
	return scanDorm3dJSON(value, list)
}

func marshalDorm3dJSON(value any) (driver.Value, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(payload), nil
}

func scanDorm3dJSON(value any, target any) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), target)
	case []byte:
		return json.Unmarshal(v, target)
	default:
		return fmt.Errorf("unsupported Dorm3d type: %T", value)
	}
}
