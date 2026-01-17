package answer

import (
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func boolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func tbInfoPlaceholder() *protobuf.TBINFO {
	// TODO: Seed New Educate (TB) state from storage.
	return &protobuf.TBINFO{
		Id: proto.Uint32(0),
		Fsm: &protobuf.TBFSM{
			SystemNo:    proto.Uint32(0),
			CurrentNode: proto.Uint32(0),
			Cache: []*protobuf.TBFSMCACHE{
				{
					CachePlan: []*protobuf.TBFSMCACHEPLAN{
						{
							CurIndex: proto.Uint32(0),
							Plans:    []*protobuf.KVDATA{},
						},
					},
					CacheTalent: []*protobuf.TBFSMCACHETALENT{
						{
							Finished:  proto.Uint32(0),
							Talents:   []uint32{},
							Retalents: []uint32{},
						},
					},
					CacheSite: []*protobuf.TBFSMCACHESITE{
						{
							Events:             []uint32{},
							Shops:              []uint32{},
							Buys:               []*protobuf.KVDATA{},
							State:              &protobuf.KVDATA{Key: proto.Uint32(0), Value: proto.Uint32(0)},
							CharacterThisRound: []uint32{},
						},
					},
					CacheChat: []*protobuf.TBFSMCACHECHAT{
						{
							Finished: proto.Uint32(0),
							Chats:    []uint32{},
						},
					},
					CacheEnd: []*protobuf.TBFSMCACHEEND{
						{
							Ends:   []uint32{},
							Select: proto.Uint32(0),
						},
					},
					CacheMind: []*protobuf.TBFSMCACHEMIND{{}},
				},
			},
		},
		Round: &protobuf.TBROUND{
			Round: proto.Uint32(1),
		},
		Res: &protobuf.TBRES{
			Attrs:    []*protobuf.KVDATA{},
			Resource: []*protobuf.KVDATA{},
		},
		Talent: &protobuf.TBTALENT{
			Talents: []uint32{},
		},
		Plan: &protobuf.TBPLAN{
			PlanUpgrade: []uint32{},
		},
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

func tbPermanentPlaceholder() *protobuf.TBPERMANENT {
	// TODO: Seed New Educate (TB) permanent data from storage.
	return &protobuf.TBPERMANENT{
		NgPlusCount:   proto.Uint32(1),
		Polaroids:     []uint32{},
		Endings:       []uint32{},
		ActiveEndings: []uint32{},
	}
}
