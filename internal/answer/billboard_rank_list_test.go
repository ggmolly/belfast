package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedSecretaryShip(t *testing.T, commander *orm.Commander, templateID uint32) {
	t.Helper()
	seedShipTemplate(t, templateID, 0, 5, 1, "Secretary", 6)
	ship := orm.OwnedShip{
		OwnerID:           commander.CommanderID,
		ShipID:            templateID,
		IsSecretary:       true,
		SecretaryPosition: proto.Uint32(0),
	}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed secretary ship: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
}

func TestBillboardRankListPageReturnsRow(t *testing.T) {
	client := setupHandlerCommander(t)
	seedSecretaryShip(t, client.Commander, 202124)

	payload := &protobuf.CS_18201{
		Page: proto.Uint32(1),
		Type: proto.Uint32(1),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := BillboardRankListPage(&buf, client); err != nil {
		t.Fatalf("BillboardRankListPage failed: %v", err)
	}

	var response protobuf.SC_18202
	decodeRegisterResponse(t, client, 18202, &response)
	if len(response.GetList()) == 0 {
		t.Fatalf("expected at least one rank row")
	}
	if len(response.GetList()) > 20 {
		t.Fatalf("expected <= 20 rows, got %d", len(response.GetList()))
	}

	row := response.GetList()[0]
	if row.UserId == nil || row.Point == nil || row.Name == nil || row.Lv == nil || row.ArenaRank == nil {
		t.Fatalf("expected required fields to be set")
	}
	if row.Display == nil {
		t.Fatalf("expected display block")
	}
	if row.Display.Icon == nil || row.Display.Skin == nil || row.Display.IconFrame == nil || row.Display.ChatFrame == nil || row.Display.IconTheme == nil || row.Display.MarryFlag == nil || row.Display.TransformFlag == nil {
		t.Fatalf("expected display subfields to be set")
	}
}

func TestBillboardMyRankReturnsRankAndPoint(t *testing.T) {
	client := setupHandlerCommander(t)
	seedSecretaryShip(t, client.Commander, 202124)
	client.Commander.Level = 10
	client.Commander.LastLogin = time.Now().UTC()
	if err := orm.GormDB.Save(client.Commander).Error; err != nil {
		t.Fatalf("save commander: %v", err)
	}

	payload := &protobuf.CS_18203{Type: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := BillboardMyRank(&buf, client); err != nil {
		t.Fatalf("BillboardMyRank failed: %v", err)
	}

	var response protobuf.SC_18204
	decodeRegisterResponse(t, client, 18204, &response)
	if response.Point == nil || response.Rank == nil {
		t.Fatalf("expected point and rank fields")
	}
	if response.GetRank() == 0 {
		t.Fatalf("expected a non-zero rank")
	}
}

func TestBillboardRankUnknownTypeReturnsEmptyAndZeroRank(t *testing.T) {
	client := setupHandlerCommander(t)
	seedSecretaryShip(t, client.Commander, 202124)

	pagePayload := &protobuf.CS_18201{Page: proto.Uint32(1), Type: proto.Uint32(999)}
	buf, err := proto.Marshal(pagePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := BillboardRankListPage(&buf, client); err != nil {
		t.Fatalf("BillboardRankListPage failed: %v", err)
	}
	var pageResp protobuf.SC_18202
	decodeRegisterResponse(t, client, 18202, &pageResp)
	if len(pageResp.GetList()) != 0 {
		t.Fatalf("expected empty list")
	}

	myPayload := &protobuf.CS_18203{Type: proto.Uint32(999)}
	buf, err = proto.Marshal(myPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := BillboardMyRank(&buf, client); err != nil {
		t.Fatalf("BillboardMyRank failed: %v", err)
	}
	var myResp protobuf.SC_18204
	decodeRegisterResponse(t, client, 18204, &myResp)
	if myResp.GetRank() != 0 || myResp.GetPoint() != 0 {
		t.Fatalf("expected rank=0 and point=0")
	}
}

func TestBillboardRankListPageNilCommanderReturnsEmpty(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_18201{Page: proto.Uint32(1), Type: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := BillboardRankListPage(&buf, client); err != nil {
		t.Fatalf("BillboardRankListPage failed: %v", err)
	}
	var response protobuf.SC_18202
	decodeRegisterResponse(t, client, 18202, &response)
	if len(response.GetList()) != 0 {
		t.Fatalf("expected empty list")
	}
}

func TestBillboardMyRankNilCommanderReturnsZero(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_18203{Type: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := BillboardMyRank(&buf, client); err != nil {
		t.Fatalf("BillboardMyRank failed: %v", err)
	}
	var response protobuf.SC_18204
	decodeRegisterResponse(t, client, 18204, &response)
	if response.GetRank() != 0 || response.GetPoint() != 0 {
		t.Fatalf("expected rank=0 and point=0")
	}
}
