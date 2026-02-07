package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestCollectionGetAward17005SuccessPersistsAndSyncs(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderStoreupAwardProgress{})

	const (
		storeupID = 1
		groupID   = 10001
	)
	seedConfigEntry(t, storeupDataTemplateCategory, "1", `{"id":1,"char_list":[10001],"level":[4],"award_display":[[1,4,50]]}`)

	shipTemplateID := uint32(groupID*10 + 1)
	seedShipTemplate(t, shipTemplateID, 1, 2, 1, "Test Ship", 5)
	seedOwnedShip(t, client, shipTemplateID)

	payload := protobuf.CS_17005{Id: proto.Uint32(storeupID), AwardIndex: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := CollectionGetAward17005(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_17006
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	row, err := orm.GetCommanderStoreupAwardProgress(orm.GormDB, client.Commander.CommanderID, storeupID)
	if err != nil {
		t.Fatalf("expected progress row: %v", err)
	}
	if row.LastAwardIndex != 1 {
		t.Fatalf("expected last award index 1")
	}
	if resource := client.Commander.OwnedResourcesMap[4]; resource == nil || resource.Amount < 50 {
		t.Fatalf("expected resource reward applied")
	}

	client.Buffer.Reset()
	empty := []byte{}
	if _, _, err := CommanderCollection(&empty, client); err != nil {
		t.Fatalf("commander collection failed: %v", err)
	}
	var sync protobuf.SC_17001
	decodeResponse(t, client, &sync)
	if len(sync.GetShipAwardList()) != 1 {
		t.Fatalf("expected ship award list")
	}
	entry := sync.GetShipAwardList()[0]
	if entry.GetId() != storeupID {
		t.Fatalf("expected ship award id %d", storeupID)
	}
	if len(entry.GetAwardIndex()) != 1 || entry.GetAwardIndex()[0] != 1 {
		t.Fatalf("expected award_index [1]")
	}
}

func TestCollectionGetAward17005InsufficientStarsDoesNotPersist(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderStoreupAwardProgress{})

	seedConfigEntry(t, storeupDataTemplateCategory, "1", `{"id":1,"char_list":[10001],"level":[10],"award_display":[[1,4,50]]}`)
	shipTemplateID := uint32(10001*10 + 1)
	seedShipTemplate(t, shipTemplateID, 1, 2, 1, "Test Ship", 3)
	seedOwnedShip(t, client, shipTemplateID)

	payload := protobuf.CS_17005{Id: proto.Uint32(1), AwardIndex: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := CollectionGetAward17005(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_17006
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	if _, err := orm.GetCommanderStoreupAwardProgress(orm.GormDB, client.Commander.CommanderID, 1); err == nil {
		t.Fatalf("expected no progress row")
	}
	if resource := client.Commander.OwnedResourcesMap[4]; resource != nil && resource.Amount > 0 {
		t.Fatalf("expected no reward applied")
	}
}

func TestCollectionGetAward17005WrongTierRejected(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderStoreupAwardProgress{})

	seedConfigEntry(t, storeupDataTemplateCategory, "1", `{"id":1,"char_list":[10001],"level":[1,2],"award_display":[[1,4,50],[1,4,60]]}`)
	shipTemplateID := uint32(10001*10 + 1)
	seedShipTemplate(t, shipTemplateID, 1, 2, 1, "Test Ship", 5)
	seedOwnedShip(t, client, shipTemplateID)

	if err := orm.GormDB.Create(&orm.CommanderStoreupAwardProgress{CommanderID: client.Commander.CommanderID, StoreupID: 1, LastAwardIndex: 1}).Error; err != nil {
		t.Fatalf("seed progress: %v", err)
	}

	payload := protobuf.CS_17005{Id: proto.Uint32(1), AwardIndex: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := CollectionGetAward17005(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_17006
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}
