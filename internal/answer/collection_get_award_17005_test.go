package answer

import (
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
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

func TestCollectionGetAward17005ConcurrentClaimDoesNotDuplicateReward(t *testing.T) {
	client1 := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.CommanderStoreupAwardProgress{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.Ship{})

	// Make sqlite wait briefly instead of returning "database is locked" on concurrent writes.
	_ = orm.GormDB.Exec("PRAGMA busy_timeout = 5000").Error

	commanderID := client1.Commander.CommanderID
	commander2 := orm.Commander{CommanderID: commanderID}
	if err := commander2.Load(); err != nil {
		t.Fatalf("load commander2: %v", err)
	}
	client2 := &connection.Client{Commander: &commander2}
	client2.Server = connection.NewServer("127.0.0.1", 0, func(pkt *[]byte, c *connection.Client, size int) {})

	const (
		storeupID         = 1
		eligibleGroupID   = 10001
		dropShipTemplate  = 20001
		dropShipCount     = 30
		eligibleShipStar  = 5
		awardLevelRequire = 1
	)

	seedConfigEntry(t, storeupDataTemplateCategory, "1", `{"id":1,"char_list":[10001],"level":[1],"award_display":[[4,20001,30]]}`)

	eligibleShipTemplateID := uint32(eligibleGroupID*10 + 1)
	seedShipTemplate(t, eligibleShipTemplateID, 1, 2, 1, "Eligible Ship", eligibleShipStar)
	seedOwnedShip(t, client1, eligibleShipTemplateID)
	seedShipTemplate(t, dropShipTemplate, 1, 2, 1, "Drop Ship", 1)

	payload := protobuf.CS_17005{Id: proto.Uint32(storeupID), AwardIndex: proto.Uint32(awardLevelRequire)}
	buffer1, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	buffer2 := append([]byte(nil), buffer1...)

	var (
		wg       sync.WaitGroup
		start    = make(chan struct{})
		resp1    protobuf.SC_17006
		resp2    protobuf.SC_17006
		callErr1 error
		callErr2 error
	)

	call := func(client *connection.Client, buf *[]byte, out *protobuf.SC_17006, callErr *error) {
		defer wg.Done()
		<-start
		client.Buffer.Reset()
		if _, _, err := CollectionGetAward17005(buf, client); err != nil {
			*callErr = err
			return
		}
		decodeResponse(t, client, out)
	}

	wg.Add(2)
	go call(client1, &buffer1, &resp1, &callErr1)
	go call(client2, &buffer2, &resp2, &callErr2)
	close(start)
	wg.Wait()

	if callErr1 != nil || callErr2 != nil {
		t.Fatalf("handler errors: %v / %v", callErr1, callErr2)
	}
	if (resp1.GetResult() == 0) == (resp2.GetResult() == 0) {
		t.Fatalf("expected exactly one success result, got %d / %d", resp1.GetResult(), resp2.GetResult())
	}

	var count int64
	if err := orm.GormDB.Model(&orm.OwnedShip{}).
		Where("owner_id = ? AND ship_id = ?", commanderID, uint32(dropShipTemplate)).
		Count(&count).Error; err != nil {
		t.Fatalf("count owned ships: %v", err)
	}
	if count != int64(dropShipCount) {
		t.Fatalf("expected %d drop ships, got %d", dropShipCount, count)
	}
}
