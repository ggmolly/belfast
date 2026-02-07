package answer_test

import (
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func clearTable(t *testing.T, model any) {
	t.Helper()
	if err := orm.GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}

func TestCompositeSpWeaponSuccess(t *testing.T) {
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.OwnedSpWeapon{})
	clearTable(t, &orm.Commander{})
	commander := orm.Commander{CommanderID: 1, AccountID: 1, Name: "SpWeapon Commander"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_14209{
		TemplateId:     proto.Uint32(12345),
		ItemIdList:     []uint32{1, 2, 3},
		SpweaponIdList: []uint32{10, 11},
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.CompositeSpWeapon(&buf, client); err != nil {
		t.Fatalf("CompositeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14210{}
	packetId := decodeTestPacket(t, client, 14210, response)
	if packetId != 14210 {
		t.Fatalf("expected packet 14210, got %d", packetId)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetSpweapon() == nil {
		t.Fatalf("expected spweapon to be populated")
	}
	if response.GetSpweapon().GetTemplateId() != payload.GetTemplateId() {
		t.Fatalf("expected spweapon.template_id %d, got %d", payload.GetTemplateId(), response.GetSpweapon().GetTemplateId())
	}
	if response.GetSpweapon().GetId() == 0 {
		t.Fatalf("expected spweapon.id to be non-zero")
	}
}

func TestCompositeSpWeaponMissingTemplateId(t *testing.T) {
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.OwnedSpWeapon{})
	clearTable(t, &orm.Commander{})
	commander := orm.Commander{CommanderID: 1, AccountID: 1, Name: "SpWeapon Commander"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_14209{
		TemplateId: proto.Uint32(0),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.CompositeSpWeapon(&buf, client); err != nil {
		t.Fatalf("CompositeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14210{}
	decodeTestPacket(t, client, 14210, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
	if response.GetSpweapon() != nil {
		t.Fatalf("expected spweapon to be nil on failure")
	}
}
