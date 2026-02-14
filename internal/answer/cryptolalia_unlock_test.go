package answer

import (
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupCryptolaliaUnlockTest(t *testing.T, gems uint32) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.CommanderSoundStory{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})

	if err := orm.CreateCommanderRoot(1, 1, "Cryptolalia Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 1}
	// Seed a secretary so PlayerInfo can render.
	pos := uint32(0)
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (template_id) DO NOTHING", int64(202124), "Secretary Ship", "Secretary Ship", int64(1), int64(1), int64(1), int64(1), int64(0))
	ship := orm.OwnedShip{OwnerID: 1, ShipID: 202124, IsSecretary: true, SecretaryPosition: &pos}
	if err := ship.Create(); err != nil {
		t.Fatalf("seed secretary: %v", err)
	}
	execAnswerTestSQLT(t, "UPDATE owned_ships SET is_secretary = $1, secretary_position = $2 WHERE id = $3", true, int64(0), int64(ship.ID))
	commander = orm.Commander{CommanderID: 1}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if err := commander.SetResource(4, gems); err != nil {
		t.Fatalf("seed gems: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedSoundStoryTemplateAlways(t *testing.T, id uint32) {
	t.Helper()
	seedConfigEntry(t, "ShareCfg/soundstory_template.json", "1", `{"id":1,"cost1":[1,4,120],"cost2":[2,15010,3],"time":"always"}`)
}

func TestCryptolaliaUnlockConsumesCurrencyAndPersists(t *testing.T) {
	client := setupCryptolaliaUnlockTest(t, 200)
	seedSoundStoryTemplateAlways(t, 1)

	request := &protobuf.CS_16205{Id: proto.Uint32(1), CostType: proto.Uint32(1)}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := CryptolaliaUnlock(&buf, client); err != nil {
		t.Fatalf("CryptolaliaUnlock: %v", err)
	}
	var resp protobuf.SC_16206
	decodePacketAt(t, client, 0, 16206, &resp)
	client.Buffer.Reset()
	if resp.GetRet() != 0 {
		t.Fatalf("expected ret=0, got %d", resp.GetRet())
	}
	if client.Commander.GetResourceCount(4) != 80 {
		t.Fatalf("expected gems to be consumed")
	}

	ids, err := orm.ListCommanderSoundStoryIDs(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("list sound stories: %v", err)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("expected persisted sound story")
	}
}

func TestCryptolaliaUnlockIsIdempotent(t *testing.T) {
	client := setupCryptolaliaUnlockTest(t, 200)
	seedSoundStoryTemplateAlways(t, 1)

	request := &protobuf.CS_16205{Id: proto.Uint32(1), CostType: proto.Uint32(1)}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := CryptolaliaUnlock(&buf, client); err != nil {
		t.Fatalf("CryptolaliaUnlock first: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := CryptolaliaUnlock(&buf, client); err != nil {
		t.Fatalf("CryptolaliaUnlock second: %v", err)
	}
	var resp protobuf.SC_16206
	decodePacketAt(t, client, 0, 16206, &resp)
	client.Buffer.Reset()
	if resp.GetRet() != 0 {
		t.Fatalf("expected ret=0, got %d", resp.GetRet())
	}
	if client.Commander.GetResourceCount(4) != 80 {
		t.Fatalf("expected currency not to be consumed twice")
	}
}

func TestCryptolaliaUnlockEmitsOnLogin(t *testing.T) {
	client := setupCryptolaliaUnlockTest(t, 200)
	seedSoundStoryTemplateAlways(t, 1)

	request := &protobuf.CS_16205{Id: proto.Uint32(1), CostType: proto.Uint32(1)}
	buf, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, _, err := CryptolaliaUnlock(&buf, client); err != nil {
		t.Fatalf("CryptolaliaUnlock: %v", err)
	}
	client.Buffer.Reset()
	loginBuf := []byte{}
	if _, _, err := PlayerInfo(&loginBuf, client); err != nil {
		t.Fatalf("PlayerInfo: %v", err)
	}
	var info protobuf.SC_11003
	decodePacketAt(t, client, 0, 11003, &info)
	client.Buffer.Reset()
	if len(info.GetSoundstory()) != 1 || info.GetSoundstory()[0] != 1 {
		t.Fatalf("expected soundstory to include unlocked id")
	}
}
