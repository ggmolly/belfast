package answer

import (
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupMonthShopFlagTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.CommanderCommonFlag{})
	clearTable(t, &orm.Commander{})
	if err := orm.CreateCommanderRoot(1, 1, "Month Shop Flag Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 1}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestMonthShopFlagPersistsAsCommonFlag(t *testing.T) {
	client := setupMonthShopFlagTest(t)

	payload := &protobuf.CS_16203{Flag: proto.Uint32(123)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if _, _, err := MonthShopFlag(&buf, client); err != nil {
		t.Fatalf("MonthShopFlag: %v", err)
	}
	var resp protobuf.SC_16204
	decodePacketAt(t, client, 0, 16204, &resp)
	client.Buffer.Reset()
	if resp.GetRet() != 0 {
		t.Fatalf("expected ret=0, got %d", resp.GetRet())
	}

	flags, err := orm.ListCommanderCommonFlags(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	if !containsUint32(flags, 123) {
		t.Fatalf("expected flag to be persisted")
	}
}

func TestMonthShopFlagIsIdempotent(t *testing.T) {
	client := setupMonthShopFlagTest(t)

	payload := &protobuf.CS_16203{Flag: proto.Uint32(555)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if _, _, err := MonthShopFlag(&buf, client); err != nil {
		t.Fatalf("MonthShopFlag first: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := MonthShopFlag(&buf, client); err != nil {
		t.Fatalf("MonthShopFlag second: %v", err)
	}

	count := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM commander_common_flags WHERE commander_id = $1 AND flag_id = $2", int64(client.Commander.CommanderID), int64(555))
	if count != 1 {
		t.Fatalf("expected 1 flag row, got %d", count)
	}
}
