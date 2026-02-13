package answer

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var composeCommanderID uint32 = 12000

func setupComposeItemTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.CommanderMiscItem{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})

	commanderID := atomic.AddUint32(&composeCommanderID, 1)
	commander := orm.Commander{CommanderID: commanderID, AccountID: 1, Name: fmt.Sprintf("Compose Tester %d", commanderID)}
	if err := orm.CreateCommanderRoot(commanderID, 1, commander.Name, 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander = orm.Commander{CommanderID: commanderID}
	client := &connection.Client{Commander: &commander}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return client
}

func TestComposeItemSuccess(t *testing.T) {
	client := setupComposeItemTest(t)
	seedConfigEntry(t, itemDataStatisticsCategory, "1000", `{"id":1000,"compose_number":3,"target_id":2000}`)
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1000), int64(10))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	request := protobuf.CS_15006{Id: proto.Uint32(1000), Num: proto.Uint32(2)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := ComposeItem(&buffer, client); err != nil {
		t.Fatalf("compose item failed: %v", err)
	}

	var response protobuf.SC_15007
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	source := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(1000))
	if source != 4 {
		t.Fatalf("expected source count 4, got %d", source)
	}
	target := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(2000))
	if target != 2 {
		t.Fatalf("expected target count 2, got %d", target)
	}
}

func TestComposeItemInsufficientItems(t *testing.T) {
	client := setupComposeItemTest(t)
	seedConfigEntry(t, itemDataStatisticsCategory, "1000", `{"id":1000,"compose_number":3,"target_id":2000}`)
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1000), int64(5))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	request := protobuf.CS_15006{Id: proto.Uint32(1000), Num: proto.Uint32(2)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := ComposeItem(&buffer, client); err != nil {
		t.Fatalf("compose item failed: %v", err)
	}

	var response protobuf.SC_15007
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	source := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(1000))
	if source != 5 {
		t.Fatalf("expected source count to remain 5, got %d", source)
	}
	targetCount := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(2000))
	if targetCount != 0 {
		t.Fatalf("expected no target item to be granted")
	}
}

func TestComposeItemUsesCombinedItemSources(t *testing.T) {
	client := setupComposeItemTest(t)
	seedConfigEntry(t, itemDataStatisticsCategory, "1000", `{"id":1000,"compose_number":3,"target_id":2000}`)
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1000), int64(2))
	execAnswerTestSQLT(t, "INSERT INTO commander_misc_items (commander_id, item_id, data) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1000), int64(10))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	request := protobuf.CS_15006{Id: proto.Uint32(1000), Num: proto.Uint32(4)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := ComposeItem(&buffer, client); err != nil {
		t.Fatalf("compose item failed: %v", err)
	}

	var response protobuf.SC_15007
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	source := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(1000))
	if source != 0 {
		t.Fatalf("expected source count 0, got %d", source)
	}
	misc := queryAnswerTestInt64(t, "SELECT data FROM commander_misc_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(1000))
	if misc != 0 {
		t.Fatalf("expected misc count 0, got %d", misc)
	}
	target := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(2000))
	if target != 4 {
		t.Fatalf("expected target count 4, got %d", target)
	}
}

func TestComposeItemInvalidConfig(t *testing.T) {
	client := setupComposeItemTest(t)
	seedConfigEntry(t, itemDataStatisticsCategory, "1000", `{"id":1000,"compose_number":0,"target_id":2000}`)
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1000), int64(10))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	request := protobuf.CS_15006{Id: proto.Uint32(1000), Num: proto.Uint32(2)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := ComposeItem(&buffer, client); err != nil {
		t.Fatalf("compose item failed: %v", err)
	}

	var response protobuf.SC_15007
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	source := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(1000))
	if source != 10 {
		t.Fatalf("expected source count to remain 10, got %d", source)
	}
}
