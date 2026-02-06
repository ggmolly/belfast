package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestReqPlayerAssistShip_Type0_BlankShipsAligned(t *testing.T) {
	client := setupPlayerUpdateTest(t)

	request := protobuf.CS_12301{Type: proto.Uint32(0), IdList: []uint32{111, 222, 333}}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := ReqPlayerAssistShip(&buffer, client); err != nil {
		t.Fatalf("req player assist ship failed: %v", err)
	}

	var response protobuf.SC_12302
	decodeResponse(t, client, &response)
	if len(response.GetShipList()) != 3 {
		t.Fatalf("expected 3 ships, got %d", len(response.GetShipList()))
	}
	for _, ship := range response.GetShipList() {
		assertBlankAssistShipRequired(t, ship)
	}
}

func TestReqPlayerAssistShip_Type1_BlankShipsAligned(t *testing.T) {
	client := setupPlayerUpdateTest(t)

	request := protobuf.CS_12301{Type: proto.Uint32(1), IdList: []uint32{111, 222, 333}}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := ReqPlayerAssistShip(&buffer, client); err != nil {
		t.Fatalf("req player assist ship failed: %v", err)
	}

	var response protobuf.SC_12302
	decodeResponse(t, client, &response)
	if len(response.GetShipList()) != 3 {
		t.Fatalf("expected 3 ships, got %d", len(response.GetShipList()))
	}
	for _, ship := range response.GetShipList() {
		assertBlankAssistShipRequired(t, ship)
	}
}

func TestReqPlayerAssistShip_UnknownType_DoesNotError(t *testing.T) {
	client := setupPlayerUpdateTest(t)

	request := protobuf.CS_12301{Type: proto.Uint32(99), IdList: []uint32{111, 222, 333}}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := ReqPlayerAssistShip(&buffer, client); err != nil {
		t.Fatalf("expected unknown type to succeed, got error: %v", err)
	}

	var response protobuf.SC_12302
	decodeResponse(t, client, &response)
	if len(response.GetShipList()) != 3 {
		t.Fatalf("expected 3 ships, got %d", len(response.GetShipList()))
	}
	for _, ship := range response.GetShipList() {
		assertBlankAssistShipRequired(t, ship)
	}
}

func TestReqPlayerAssistShip_EmptyIdList_ReturnsEmptyShipList(t *testing.T) {
	client := setupPlayerUpdateTest(t)

	request := protobuf.CS_12301{Type: proto.Uint32(0), IdList: []uint32{}}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := ReqPlayerAssistShip(&buffer, client); err != nil {
		t.Fatalf("req player assist ship failed: %v", err)
	}

	var response protobuf.SC_12302
	decodeResponse(t, client, &response)
	if len(response.GetShipList()) != 0 {
		t.Fatalf("expected empty ship list, got %d", len(response.GetShipList()))
	}
}

func assertBlankAssistShipRequired(t *testing.T, ship *protobuf.SHIPINFO) {
	t.Helper()

	if ship.TemplateId == nil {
		t.Fatalf("expected TemplateId to be set")
	}
	if ship.GetTemplateId() != 0 {
		t.Fatalf("expected TemplateId 0, got %d", ship.GetTemplateId())
	}

	if ship.Id == nil {
		t.Fatalf("expected Id to be set")
	}
	if ship.Level == nil {
		t.Fatalf("expected Level to be set")
	}
	if ship.Exp == nil {
		t.Fatalf("expected Exp to be set")
	}
	if ship.Energy == nil {
		t.Fatalf("expected Energy to be set")
	}
	if ship.IsLocked == nil {
		t.Fatalf("expected IsLocked to be set")
	}
	if ship.Intimacy == nil {
		t.Fatalf("expected Intimacy to be set")
	}
	if ship.Proficiency == nil {
		t.Fatalf("expected Proficiency to be set")
	}
	if ship.CreateTime == nil {
		t.Fatalf("expected CreateTime to be set")
	}
	if ship.SkinId == nil {
		t.Fatalf("expected SkinId to be set")
	}
	if ship.Propose == nil {
		t.Fatalf("expected Propose to be set")
	}
	if ship.MaxLevel == nil {
		t.Fatalf("expected MaxLevel to be set")
	}
	if ship.ActivityNpc == nil {
		t.Fatalf("expected ActivityNpc to be set")
	}

	if ship.State == nil {
		t.Fatalf("expected State to be set")
	}
	if ship.State.State == nil {
		t.Fatalf("expected State.State to be set")
	}
}
