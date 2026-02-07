package answer_test

import (
	"os"
	"slices"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupSpWeaponClient(t *testing.T) *connection.Client {
	t.Helper()
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
	return &connection.Client{Commander: &commander}
}

func TestConfirmReforgeSpWeaponExchange(t *testing.T) {
	client := setupSpWeaponClient(t)

	spweapon := orm.OwnedSpWeapon{
		OwnerID:    client.Commander.CommanderID,
		TemplateID: 1001,
		Attr1:      11,
		Attr2:      22,
		AttrTemp1:  33,
		AttrTemp2:  44,
	}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14207{
		ShipId:     proto.Uint32(0),
		SpweaponId: proto.Uint32(spweapon.ID),
		Cmd:        proto.Uint32(1),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ConfirmReforgeSpWeapon(&buf, client); err != nil {
		t.Fatalf("ConfirmReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14208{}
	decodeTestPacket(t, client, 14208, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.First(&stored, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.Attr1 != 33 || stored.Attr2 != 44 {
		t.Fatalf("expected base attrs to be exchanged to 33/44, got %d/%d", stored.Attr1, stored.Attr2)
	}
	if stored.AttrTemp1 != 0 || stored.AttrTemp2 != 0 {
		t.Fatalf("expected temp attrs to be cleared, got %d/%d", stored.AttrTemp1, stored.AttrTemp2)
	}
}

func TestConfirmReforgeSpWeaponDiscard(t *testing.T) {
	client := setupSpWeaponClient(t)

	spweapon := orm.OwnedSpWeapon{
		OwnerID:    client.Commander.CommanderID,
		TemplateID: 1001,
		Attr1:      11,
		Attr2:      22,
		AttrTemp1:  33,
		AttrTemp2:  44,
	}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14207{
		ShipId:     proto.Uint32(0),
		SpweaponId: proto.Uint32(spweapon.ID),
		Cmd:        proto.Uint32(0),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ConfirmReforgeSpWeapon(&buf, client); err != nil {
		t.Fatalf("ConfirmReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14208{}
	decodeTestPacket(t, client, 14208, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.First(&stored, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.Attr1 != 11 || stored.Attr2 != 22 {
		t.Fatalf("expected base attrs to be unchanged 11/22, got %d/%d", stored.Attr1, stored.Attr2)
	}
	if stored.AttrTemp1 != 0 || stored.AttrTemp2 != 0 {
		t.Fatalf("expected temp attrs to be cleared, got %d/%d", stored.AttrTemp1, stored.AttrTemp2)
	}
}

func TestConfirmReforgeSpWeaponUnknownUID(t *testing.T) {
	client := setupSpWeaponClient(t)

	payload := &protobuf.CS_14207{
		ShipId:     proto.Uint32(0),
		SpweaponId: proto.Uint32(9999),
		Cmd:        proto.Uint32(1),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ConfirmReforgeSpWeapon(&buf, client); err != nil {
		t.Fatalf("ConfirmReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14208{}
	decodeTestPacket(t, client, 14208, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestConfirmReforgeSpWeaponInvalidCmd(t *testing.T) {
	client := setupSpWeaponClient(t)

	spweapon := orm.OwnedSpWeapon{
		OwnerID:    client.Commander.CommanderID,
		TemplateID: 1001,
		Attr1:      11,
		Attr2:      22,
		AttrTemp1:  33,
		AttrTemp2:  44,
	}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14207{
		ShipId:     proto.Uint32(0),
		SpweaponId: proto.Uint32(spweapon.ID),
		Cmd:        proto.Uint32(2),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ConfirmReforgeSpWeapon(&buf, client); err != nil {
		t.Fatalf("ConfirmReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14208{}
	decodeTestPacket(t, client, 14208, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.First(&stored, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.Attr1 != 11 || stored.Attr2 != 22 || stored.AttrTemp1 != 33 || stored.AttrTemp2 != 44 {
		t.Fatalf("expected spweapon attrs to remain unchanged on invalid cmd")
	}
}

func TestConfirmReforgeSpWeaponAfterCompositeRebuildsMapPointers(t *testing.T) {
	client := setupSpWeaponClient(t)

	spweapon := orm.OwnedSpWeapon{
		OwnerID:    client.Commander.CommanderID,
		TemplateID: 1001,
		Attr1:      11,
		Attr2:      22,
		AttrTemp1:  33,
		AttrTemp2:  44,
	}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	// Ensure appending spweapons reallocates the slice backing array.
	client.Commander.OwnedSpWeapons = slices.Clip(client.Commander.OwnedSpWeapons)

	compositePayload := &protobuf.CS_14209{TemplateId: proto.Uint32(12345)}
	compositeBuf, err := proto.Marshal(compositePayload)
	if err != nil {
		t.Fatalf("failed to marshal composite payload: %v", err)
	}
	if _, _, err := answer.CompositeSpWeapon(&compositeBuf, client); err != nil {
		t.Fatalf("CompositeSpWeapon failed: %v", err)
	}
	decodeTestPacket(t, client, 14210, &protobuf.SC_14210{})

	confirmPayload := &protobuf.CS_14207{
		ShipId:     proto.Uint32(0),
		SpweaponId: proto.Uint32(spweapon.ID),
		Cmd:        proto.Uint32(1),
	}
	confirmBuf, err := proto.Marshal(confirmPayload)
	if err != nil {
		t.Fatalf("failed to marshal confirm payload: %v", err)
	}
	if _, _, err := answer.ConfirmReforgeSpWeapon(&confirmBuf, client); err != nil {
		t.Fatalf("ConfirmReforgeSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14208{}
	decodeTestPacket(t, client, 14208, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var inMemory *orm.OwnedSpWeapon
	for i := range client.Commander.OwnedSpWeapons {
		if client.Commander.OwnedSpWeapons[i].ID == spweapon.ID {
			inMemory = &client.Commander.OwnedSpWeapons[i]
			break
		}
	}
	if inMemory == nil {
		t.Fatalf("expected spweapon to be present in commander state")
	}
	if inMemory.Attr1 != 33 || inMemory.Attr2 != 44 {
		t.Fatalf("expected commander state base attrs 33/44, got %d/%d", inMemory.Attr1, inMemory.Attr2)
	}
	if inMemory.AttrTemp1 != 0 || inMemory.AttrTemp2 != 0 {
		t.Fatalf("expected commander state temp attrs to be cleared, got %d/%d", inMemory.AttrTemp1, inMemory.AttrTemp2)
	}
}
