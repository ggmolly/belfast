package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	newEducateSystemEvent  = 1
	newEducateSystemTalent = 2
	newEducateSystemTopic  = 3
	newEducateSystemMap    = 4
	newEducateSystemPlan   = 5
	newEducateSystemAssess = 6
	newEducateSystemPhase  = 7
	newEducateSystemEnding = 8
	newEducateSystemMind   = 9

	newEducateSiteStateEvent  = 1
	newEducateSiteStateNormal = 2
	newEducateSiteStateShip   = 3
)

type educateState struct {
	Entry     *orm.CommanderTB
	Info      *protobuf.TBINFO
	Permanent *protobuf.TBPERMANENT
}

func loadEducateState(client *connection.Client, tbID uint32) (*educateState, error) {
	entry, err := orm.GetCommanderTB(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		entry, err = orm.NewCommanderTB(client.Commander.CommanderID, tbInfoPlaceholder(), tbPermanentPlaceholder())
		if err != nil {
			return nil, err
		}
		if err := orm.GormDB.Create(entry).Error; err != nil {
			return nil, err
		}
	}
	info, permanent, err := entry.Decode()
	if err != nil {
		return nil, err
	}
	info = ensureTBInfoDefaults(info)
	permanent = ensureTBPermanentDefaults(permanent)
	info.Id = proto.Uint32(tbID)
	return &educateState{Entry: entry, Info: info, Permanent: permanent}, nil
}

func saveEducateState(state *educateState) error {
	return orm.SaveCommanderTB(orm.GormDB, state.Entry, state.Info, state.Permanent)
}

func ensureTBInfoDefaults(info *protobuf.TBINFO) *protobuf.TBINFO {
	defaults := tbInfoPlaceholder()
	if info == nil {
		return defaults
	}
	if info.Fsm == nil {
		info.Fsm = defaults.Fsm
	}
	if len(info.Fsm.Cache) == 0 {
		info.Fsm.Cache = defaults.Fsm.Cache
	}
	cache := info.Fsm.Cache[0]
	if len(cache.CachePlan) == 0 {
		cache.CachePlan = defaults.Fsm.Cache[0].CachePlan
	}
	if len(cache.CacheTalent) == 0 {
		cache.CacheTalent = defaults.Fsm.Cache[0].CacheTalent
	}
	if len(cache.CacheSite) == 0 {
		cache.CacheSite = defaults.Fsm.Cache[0].CacheSite
	}
	if len(cache.CacheChat) == 0 {
		cache.CacheChat = defaults.Fsm.Cache[0].CacheChat
	}
	if len(cache.CacheEnd) == 0 {
		cache.CacheEnd = defaults.Fsm.Cache[0].CacheEnd
	}
	if len(cache.CacheMind) == 0 {
		cache.CacheMind = defaults.Fsm.Cache[0].CacheMind
	}
	if info.Round == nil {
		info.Round = defaults.Round
	}
	if info.Res == nil {
		info.Res = defaults.Res
	}
	if info.Talent == nil {
		info.Talent = defaults.Talent
	}
	if info.Plan == nil {
		info.Plan = defaults.Plan
	}
	if info.Site == nil {
		info.Site = defaults.Site
	}
	if info.Benefit == nil {
		info.Benefit = defaults.Benefit
	}
	if info.Evaluations == nil {
		info.Evaluations = []*protobuf.KVDATA{}
	}
	return info
}

func ensureTBPermanentDefaults(permanent *protobuf.TBPERMANENT) *protobuf.TBPERMANENT {
	if permanent == nil {
		return tbPermanentPlaceholder()
	}
	if permanent.Polaroids == nil {
		permanent.Polaroids = []uint32{}
	}
	if permanent.Endings == nil {
		permanent.Endings = []uint32{}
	}
	if permanent.ActiveEndings == nil {
		permanent.ActiveEndings = []uint32{}
	}
	return permanent
}

func emptyTBDrops() *protobuf.TBDROPS {
	return &protobuf.TBDROPS{
		BaseDrop:    []*protobuf.TBDROP{},
		BenefitDrop: []*protobuf.TBDROP{},
	}
}

func ensureEducateCache(info *protobuf.TBINFO) *protobuf.TBFSMCACHE {
	info = ensureTBInfoDefaults(info)
	return info.Fsm.Cache[0]
}
