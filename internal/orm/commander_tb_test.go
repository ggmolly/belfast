package orm

import (
	"testing"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestSaveCommanderTBUpdatesState(t *testing.T) {
	initRandomFlagShipTestDB(t)
	commanderID := uint32(4242)
	if err := GormDB.Where("commander_id = ?", commanderID).Delete(&CommanderTB{}).Error; err != nil {
		t.Fatalf("clear tb state: %v", err)
	}
	info := buildTestTBInfo()
	permanent := &protobuf.TBPERMANENT{
		NgPlusCount:   proto.Uint32(1),
		Polaroids:     []uint32{},
		Endings:       []uint32{},
		ActiveEndings: []uint32{},
	}
	entry, err := NewCommanderTB(commanderID, info, permanent)
	if err != nil {
		t.Fatalf("new commander tb: %v", err)
	}
	if err := GormDB.Create(entry).Error; err != nil {
		t.Fatalf("create commander tb: %v", err)
	}
	info.Name = proto.String("TEST_NAME")
	if err := SaveCommanderTB(GormDB, entry, info, permanent); err != nil {
		t.Fatalf("save commander tb: %v", err)
	}
	loaded, err := GetCommanderTB(GormDB, commanderID)
	if err != nil {
		t.Fatalf("load commander tb: %v", err)
	}
	loadedInfo, _, err := loaded.Decode()
	if err != nil {
		t.Fatalf("decode commander tb: %v", err)
	}
	if loadedInfo.GetName() != "TEST_NAME" {
		t.Fatalf("expected updated name, got %q", loadedInfo.GetName())
	}
}

func TestCommanderTBEncodeDecodeErrors(t *testing.T) {
	entry := &CommanderTB{CommanderID: 1, State: []byte{0x01}, Permanent: []byte{0x02}}
	if _, _, err := entry.Decode(); err == nil {
		t.Fatalf("expected decode error for invalid proto")
	}
	if _, err := NewCommanderTB(1, nil, &protobuf.TBPERMANENT{}); err == nil {
		t.Fatalf("expected error for nil info")
	}
	if err := entry.Encode(nil, &protobuf.TBPERMANENT{}); err == nil {
		t.Fatalf("expected encode error for nil info")
	}
}

func buildTestTBInfo() *protobuf.TBINFO {
	return &protobuf.TBINFO{
		Id: proto.Uint32(1),
		Fsm: &protobuf.TBFSM{
			SystemNo:    proto.Uint32(0),
			CurrentNode: proto.Uint32(0),
			Cache: []*protobuf.TBFSMCACHE{
				{
					CachePlan: []*protobuf.TBFSMCACHEPLAN{{
						CurIndex: proto.Uint32(0),
						Plans:    []*protobuf.KVDATA{},
					}},
					CacheTalent: []*protobuf.TBFSMCACHETALENT{{
						Finished:  proto.Uint32(0),
						Talents:   []uint32{},
						Retalents: []uint32{},
					}},
					CacheSite: []*protobuf.TBFSMCACHESITE{{
						Events:             []uint32{},
						Shops:              []uint32{},
						Buys:               []*protobuf.KVDATA{},
						State:              &protobuf.KVDATA{Key: proto.Uint32(0), Value: proto.Uint32(0)},
						CharacterThisRound: []uint32{},
					}},
					CacheChat: []*protobuf.TBFSMCACHECHAT{{
						Finished: proto.Uint32(0),
						Chats:    []uint32{},
					}},
					CacheEnd: []*protobuf.TBFSMCACHEEND{{
						Ends:   []uint32{},
						Select: proto.Uint32(0),
					}},
					CacheMind: []*protobuf.TBFSMCACHEMIND{{}},
				},
			},
		},
		Round: &protobuf.TBROUND{Round: proto.Uint32(1)},
		Res: &protobuf.TBRES{
			Attrs:    []*protobuf.KVDATA{},
			Resource: []*protobuf.KVDATA{},
		},
		Talent: &protobuf.TBTALENT{Talents: []uint32{}},
		Plan:   &protobuf.TBPLAN{PlanUpgrade: []uint32{}},
		Site: &protobuf.TBSITE{
			Characters:   []uint32{},
			WorkCounter:  []*protobuf.KVDATA{},
			Works:        []uint32{},
			EventCounter: []*protobuf.KVDATA{},
		},
		Evaluations: []*protobuf.KVDATA{},
		Name:        proto.String(""),
		FavorLv:     proto.Uint32(0),
		Benefit: &protobuf.TBBENEFIT{
			Actives:  []*protobuf.TBBF{},
			Pendings: []uint32{},
		},
	}
}
