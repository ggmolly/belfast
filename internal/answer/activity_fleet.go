package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EditActivityFleet(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11204
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11205, err
	}

	response := protobuf.SC_11205{
		ActivityId: proto.Uint32(payload.GetActivityId()),
		Result:     proto.Uint32(1),
	}

	if _, err := loadActivityTemplate(payload.GetActivityId()); err != nil {
		return client.SendMessage(11205, &response)
	}

	for _, group := range payload.GetGroupList() {
		for _, shipID := range group.GetShipList() {
			if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
				return client.SendMessage(11205, &response)
			}
		}
	}

	groups := activityFleetGroupsFromProto(payload.GetGroupList())
	if err := orm.SaveActivityFleetGroups(client.Commander.CommanderID, payload.GetActivityId(), groups); err != nil {
		return client.SendMessage(11205, &response)
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(11205, &response)
}

func activityFleetGroupsFromProto(groups []*protobuf.GROUPINFO_P11) orm.ActivityFleetGroupList {
	if len(groups) == 0 {
		return orm.ActivityFleetGroupList{}
	}
	result := make(orm.ActivityFleetGroupList, 0, len(groups))
	for _, group := range groups {
		commanders := make([]orm.ActivityFleetCommander, 0, len(group.GetCommanders()))
		for _, commander := range group.GetCommanders() {
			commanders = append(commanders, orm.ActivityFleetCommander{
				Pos: commander.GetPos(),
				ID:  commander.GetId(),
			})
		}
		result = append(result, orm.ActivityFleetGroup{
			ID:         group.GetId(),
			ShipList:   append([]uint32(nil), group.GetShipList()...),
			Commanders: commanders,
		})
	}
	return result
}

func activityFleetGroupsToProto(groups orm.ActivityFleetGroupList) []*protobuf.GROUPINFO_P11 {
	if len(groups) == 0 {
		return []*protobuf.GROUPINFO_P11{}
	}
	result := make([]*protobuf.GROUPINFO_P11, 0, len(groups))
	for _, group := range groups {
		commanders := make([]*protobuf.COMMANDERSINFO, 0, len(group.Commanders))
		for _, commander := range group.Commanders {
			commanders = append(commanders, &protobuf.COMMANDERSINFO{
				Pos: proto.Uint32(commander.Pos),
				Id:  proto.Uint32(commander.ID),
			})
		}
		result = append(result, &protobuf.GROUPINFO_P11{
			Id:         proto.Uint32(group.ID),
			ShipList:   append([]uint32(nil), group.ShipList...),
			Commanders: commanders,
		})
	}
	return result
}
