package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	// Client expects roughly 20 rows per page; we return at most one.
	billboardRankSupportedMaxType = 50

	// Minimal implementation: return a coherent single-row leaderboard.
	billboardRankPoint = 1
	billboardRankRank  = 1
)

func BillboardRankListPage(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_18201
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		response := protobuf.SC_18202{List: []*protobuf.RANK_INFO_P18{}}
		return client.SendMessage(18202, &response)
	}

	page := payload.GetPage()
	rankType := payload.GetType()
	if page != 1 || !isSupportedBillboardRankType(rankType) {
		response := protobuf.SC_18202{List: []*protobuf.RANK_INFO_P18{}}
		return client.SendMessage(18202, &response)
	}
	if client.Commander == nil {
		response := protobuf.SC_18202{List: []*protobuf.RANK_INFO_P18{}}
		return client.SendMessage(18202, &response)
	}

	row := billboardRankRow(client.Commander)
	response := protobuf.SC_18202{List: []*protobuf.RANK_INFO_P18{row}}
	return client.SendMessage(18202, &response)
}

func BillboardMyRank(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_18203
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		response := protobuf.SC_18204{Point: proto.Uint32(0), Rank: proto.Uint32(0)}
		return client.SendMessage(18204, &response)
	}

	rankType := payload.GetType()
	if !isSupportedBillboardRankType(rankType) {
		response := protobuf.SC_18204{Point: proto.Uint32(0), Rank: proto.Uint32(0)}
		return client.SendMessage(18204, &response)
	}
	if client.Commander == nil {
		response := protobuf.SC_18204{Point: proto.Uint32(0), Rank: proto.Uint32(0)}
		return client.SendMessage(18204, &response)
	}

	response := protobuf.SC_18204{Point: proto.Uint32(billboardRankPoint), Rank: proto.Uint32(billboardRankRank)}
	return client.SendMessage(18204, &response)
}

func isSupportedBillboardRankType(rankType uint32) bool {
	// Client sends small numeric constants (PowerRank.TYPE_*). Keep this bounded so unknown
	// types return empty responses.
	return rankType >= 1 && rankType <= billboardRankSupportedMaxType
}

func billboardRankRow(commander *orm.Commander) *protobuf.RANK_INFO_P18 {
	return &protobuf.RANK_INFO_P18{
		UserId:    proto.Uint32(commander.CommanderID),
		Point:     proto.Uint32(billboardRankPoint),
		Name:      proto.String(commander.Name),
		Lv:        proto.Uint32(uint32(commander.Level)),
		ArenaRank: proto.Uint32(billboardRankRank),
		Display:   billboardRankDisplay(commander),
	}
}

func billboardRankDisplay(commander *orm.Commander) *protobuf.DISPLAYINFO {
	icon := commander.DisplayIconID
	skin := commander.DisplaySkinID
	secretaries := commander.GetSecretaries()
	if icon == 0 && len(secretaries) > 0 {
		icon = secretaries[0].ShipID
	}
	if skin == 0 && len(secretaries) > 0 {
		skin = secretaries[0].SkinID
	}

	return &protobuf.DISPLAYINFO{
		Icon:          proto.Uint32(icon),
		Skin:          proto.Uint32(skin),
		IconFrame:     proto.Uint32(commander.SelectedIconFrameID),
		ChatFrame:     proto.Uint32(commander.SelectedChatFrameID),
		IconTheme:     proto.Uint32(commander.DisplayIconThemeID),
		MarryFlag:     proto.Uint32(0),
		TransformFlag: proto.Uint32(0),
	}
}
