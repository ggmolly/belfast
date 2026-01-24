package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func NewEducateGetEndings(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29003
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29004, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29004, err
	}
	endings := state.Permanent.ActiveEndings
	if len(endings) == 0 {
		endings = state.Permanent.Endings
	}
	cache := ensureEducateCache(state.Info)
	cache.CacheEnd[0].Ends = append([]uint32{}, endings...)
	cache.CacheEnd[0].Select = proto.Uint32(0)
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemEnding)
	response := protobuf.SC_29004{
		Result:  proto.Uint32(0),
		Endings: endings,
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29004, err
	}
	return client.SendMessage(29004, &response)
}

func NewEducateSelectEnding(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29005
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29006, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29006, err
	}
	state.Permanent.Endings = appendUniqueUint32(state.Permanent.Endings, payload.GetEndingId())
	cache := ensureEducateCache(state.Info)
	cache.CacheEnd[0].Select = proto.Uint32(payload.GetEndingId())
	response := protobuf.SC_29006{Result: proto.Uint32(0)}
	if err := saveEducateState(state); err != nil {
		return 0, 29006, err
	}
	return client.SendMessage(29006, &response)
}

func NewEducateReset(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29007
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29008, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29008, err
	}
	state.Info = ensureTBInfoDefaults(tbInfoPlaceholder())
	state.Info.Id = proto.Uint32(payload.GetId())
	state.Permanent.NgPlusCount = proto.Uint32(state.Permanent.GetNgPlusCount() + 1)
	response := protobuf.SC_29008{
		Result: proto.Uint32(0),
		Tb:     state.Info,
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29008, err
	}
	return client.SendMessage(29008, &response)
}

func NewEducateSetCall(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29009
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29010, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29010, err
	}
	state.Info.Name = proto.String(payload.GetName())
	response := protobuf.SC_29010{Result: proto.Uint32(0)}
	if err := saveEducateState(state); err != nil {
		return 0, 29010, err
	}
	return client.SendMessage(29010, &response)
}

func NewEducateMainEvent(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29011
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29012, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29012, err
	}
	firstNode := state.Info.Fsm.GetCurrentNode()
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemEvent)
	response := protobuf.SC_29012{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(firstNode),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29012, err
	}
	return client.SendMessage(29012, &response)
}

func NewEducateAssess(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29013
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29014, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29014, err
	}
	round := state.Info.Round.GetRound()
	state.Info.Evaluations = upsertKVDATA(state.Info.Evaluations, round, payload.GetRank())
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemAssess)
	firstNode := state.Info.Fsm.GetCurrentNode()
	response := protobuf.SC_29014{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(firstNode),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29014, err
	}
	return client.SendMessage(29014, &response)
}

func NewEducateGetTopics(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29015
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29016, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29016, err
	}
	cache := ensureEducateCache(state.Info)
	chats := cache.CacheChat[0].Chats
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemTopic)
	response := protobuf.SC_29016{
		Result: proto.Uint32(0),
		Chats:  chats,
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29016, err
	}
	return client.SendMessage(29016, &response)
}

func NewEducateSelectTopic(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29017
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29018, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29018, err
	}
	cache := ensureEducateCache(state.Info)
	cache.CacheChat[0].Finished = proto.Uint32(1)
	cache.CacheChat[0].Chats = appendUniqueUint32(cache.CacheChat[0].Chats, payload.GetChatId())
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemTopic)
	state.Info.Fsm.CurrentNode = proto.Uint32(0)
	response := protobuf.SC_29018{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(0),
		Drop:      emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29018, err
	}
	return client.SendMessage(29018, &response)
}

func NewEducateGetTalents(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29019
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29020, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29020, err
	}
	cache := ensureEducateCache(state.Info)
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemTalent)
	response := protobuf.SC_29020{
		Result:  proto.Uint32(0),
		Talents: cache.CacheTalent[0].Talents,
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29020, err
	}
	return client.SendMessage(29020, &response)
}

func NewEducateRefreshTalent(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29021
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29022, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29022, err
	}
	cache := ensureEducateCache(state.Info)
	cache.CacheTalent[0].Retalents = appendUniqueUint32(cache.CacheTalent[0].Retalents, payload.GetTalent())
	response := protobuf.SC_29022{
		Result: proto.Uint32(0),
		Talent: proto.Uint32(payload.GetTalent()),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29022, err
	}
	return client.SendMessage(29022, &response)
}

func NewEducateSelectTalent(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29023
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29024, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29024, err
	}
	state.Info.Talent.Talents = appendUniqueUint32(state.Info.Talent.Talents, payload.GetTalent())
	cache := ensureEducateCache(state.Info)
	cache.CacheTalent[0].Talents = appendUniqueUint32(cache.CacheTalent[0].Talents, payload.GetTalent())
	cache.CacheTalent[0].Finished = proto.Uint32(1)
	response := protobuf.SC_29024{Result: proto.Uint32(0)}
	if err := saveEducateState(state); err != nil {
		return 0, 29024, err
	}
	return client.SendMessage(29024, &response)
}

func NewEducateChangePhase(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29025
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29026, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29026, err
	}
	state.Info.Round.Round = proto.Uint32(state.Info.Round.GetRound() + 1)
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemPhase)
	state.Info.Fsm.CurrentNode = proto.Uint32(0)
	response := protobuf.SC_29026{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(0),
		Drop:      emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29026, err
	}
	return client.SendMessage(29026, &response)
}

func NewEducateUpgradeFavor(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29027
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29028, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29028, err
	}
	state.Info.FavorLv = proto.Uint32(state.Info.GetFavorLv() + 1)
	response := protobuf.SC_29028{
		Result: proto.Uint32(0),
		Drop:   emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29028, err
	}
	return client.SendMessage(29028, &response)
}

func NewEducateTriggerNode(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29030
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29031, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29031, err
	}
	state.Info.Fsm.CurrentNode = proto.Uint32(0)
	response := protobuf.SC_29031{
		Result:   proto.Uint32(0),
		NextNode: proto.Uint32(0),
		Drop:     emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29031, err
	}
	return client.SendMessage(29031, &response)
}

func NewEducateClearNodeChain(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29032
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29033, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29033, err
	}
	state.Info.Fsm.CurrentNode = proto.Uint32(0)
	response := protobuf.SC_29033{
		Result: proto.Uint32(0),
		Fsm:    state.Info.Fsm,
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29033, err
	}
	return client.SendMessage(29033, &response)
}

func NewEducateSchedule(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29040
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29041, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29041, err
	}
	cache := ensureEducateCache(state.Info)
	cache.CachePlan[0].Plans = payload.GetPlans()
	cache.CachePlan[0].CurIndex = proto.Uint32(0)
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemPlan)
	response := protobuf.SC_29041{
		Result: proto.Uint32(0),
		Plans:  payload.GetPlans(),
		Drop:   emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29041, err
	}
	return client.SendMessage(29041, &response)
}

func NewEducateNextPlan(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29042
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29043, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29043, err
	}
	cache := ensureEducateCache(state.Info)
	curIndex := cache.CachePlan[0].GetCurIndex()
	if int(curIndex) < len(cache.CachePlan[0].Plans) {
		cache.CachePlan[0].CurIndex = proto.Uint32(curIndex + 1)
	}
	response := protobuf.SC_29043{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(0),
		Drop:      emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29043, err
	}
	return client.SendMessage(29043, &response)
}

func NewEducateUpgradePlan(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29044
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29045, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29045, err
	}
	for _, planID := range payload.GetPlanIds() {
		state.Info.Plan.PlanUpgrade = appendUniqueUint32(state.Info.Plan.PlanUpgrade, planID)
	}
	response := protobuf.SC_29045{Result: proto.Uint32(0)}
	if err := saveEducateState(state); err != nil {
		return 0, 29045, err
	}
	return client.SendMessage(29045, &response)
}

func NewEducateScheduleSkip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29046
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29047, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29047, err
	}
	cache := ensureEducateCache(state.Info)
	cache.CachePlan[0].CurIndex = proto.Uint32(uint32(len(cache.CachePlan[0].Plans)))
	response := protobuf.SC_29047{
		Result: proto.Uint32(0),
		Drop:   emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29047, err
	}
	return client.SendMessage(29047, &response)
}

func NewEducateGetExtraDrop(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29048
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29049, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29049, err
	}
	response := protobuf.SC_29049{
		Result: proto.Uint32(0),
		Drop:   emptyTBDrops(),
		Res:    state.Info.Res,
	}
	return client.SendMessage(29049, &response)
}

func NewEducateGetMap(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29060
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29061, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29061, err
	}
	cache := ensureEducateCache(state.Info)
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemMap)
	response := protobuf.SC_29061{
		Result:     proto.Uint32(0),
		FsmSite:    cache.CacheSite[0],
		Characters: state.Info.Site.Characters,
		Drop:       emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29061, err
	}
	return client.SendMessage(29061, &response)
}

func NewEducateMapNormal(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29062
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29063, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29063, err
	}
	cache := ensureEducateCache(state.Info)
	cache.CacheSite[0].State = &protobuf.KVDATA{Key: proto.Uint32(newEducateSiteStateNormal), Value: proto.Uint32(payload.GetWorkId())}
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemMap)
	response := protobuf.SC_29063{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(0),
		Drop:      emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29063, err
	}
	return client.SendMessage(29063, &response)
}

func NewEducateMapEvent(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29064
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29065, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29065, err
	}
	cache := ensureEducateCache(state.Info)
	cache.CacheSite[0].State = &protobuf.KVDATA{Key: proto.Uint32(newEducateSiteStateEvent), Value: proto.Uint32(payload.GetEvent())}
	cache.CacheSite[0].Events = appendUniqueUint32(cache.CacheSite[0].Events, payload.GetEvent())
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemMap)
	response := protobuf.SC_29065{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(0),
		Drop:      emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29065, err
	}
	return client.SendMessage(29065, &response)
}

func NewEducateShopping(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29066
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29067, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29067, err
	}
	cache := ensureEducateCache(state.Info)
	cache.CacheSite[0].Buys = upsertKVDATACount(cache.CacheSite[0].Buys, payload.GetShop(), payload.GetNum())
	cache.CacheSite[0].Shops = appendUniqueUint32(cache.CacheSite[0].Shops, payload.GetShop())
	response := protobuf.SC_29067{
		Result: proto.Uint32(0),
		Drop:   emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29067, err
	}
	return client.SendMessage(29067, &response)
}

func NewEducateMapShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29068
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29069, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29069, err
	}
	cache := ensureEducateCache(state.Info)
	cache.CacheSite[0].State = &protobuf.KVDATA{Key: proto.Uint32(newEducateSiteStateShip), Value: proto.Uint32(payload.GetCharacter())}
	cache.CacheSite[0].CharacterThisRound = appendUniqueUint32(cache.CacheSite[0].CharacterThisRound, payload.GetCharacter())
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemMap)
	response := protobuf.SC_29069{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(0),
		Drop:      emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29069, err
	}
	return client.SendMessage(29069, &response)
}

func NewEducateUpgradeNormalSite(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29070
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29071, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29071, err
	}
	state.Info.Site.Works = appendUniqueUint32(state.Info.Site.Works, payload.GetWorkId())
	response := protobuf.SC_29071{Result: proto.Uint32(0)}
	if err := saveEducateState(state); err != nil {
		return 0, 29071, err
	}
	return client.SendMessage(29071, &response)
}

func NewEducateSelectMind(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29090
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29091, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29091, err
	}
	state.Info.Fsm.SystemNo = proto.Uint32(newEducateSystemMind)
	response := protobuf.SC_29091{
		Result:    proto.Uint32(0),
		FirstNode: proto.Uint32(0),
		Drop:      emptyTBDrops(),
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29091, err
	}
	return client.SendMessage(29091, &response)
}

func NewEducateRefresh(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29092
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29093, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29093, err
	}
	response := protobuf.SC_29093{
		Result: proto.Uint32(0),
		Tb:     state.Info,
	}
	return client.SendMessage(29093, &response)
}

func appendUniqueUint32(values []uint32, value uint32) []uint32 {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func upsertKVDATA(values []*protobuf.KVDATA, key uint32, value uint32) []*protobuf.KVDATA {
	for _, entry := range values {
		if entry.GetKey() == key {
			entry.Value = proto.Uint32(value)
			return values
		}
	}
	return append(values, &protobuf.KVDATA{Key: proto.Uint32(key), Value: proto.Uint32(value)})
}

func upsertKVDATACount(values []*protobuf.KVDATA, key uint32, increment uint32) []*protobuf.KVDATA {
	for _, entry := range values {
		if entry.GetKey() == key {
			entry.Value = proto.Uint32(entry.GetValue() + increment)
			return values
		}
	}
	return append(values, &protobuf.KVDATA{Key: proto.Uint32(key), Value: proto.Uint32(increment)})
}
