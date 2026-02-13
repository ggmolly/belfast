package answer_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func clearTable(t *testing.T, model any) {
	t.Helper()
	tableName, err := tableNameFromModel(model)
	if err != nil {
		t.Fatalf("failed to resolve table name: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", quoteIdentifier(tableName))); err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}

func tableNameFromModel(model any) (string, error) {
	t := reflect.TypeOf(model)
	if t == nil {
		return "", fmt.Errorf("model is nil")
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("model must be struct or pointer to struct")
	}

	name := t.Name()
	if name == "" {
		return "", fmt.Errorf("model type has no name")
	}

	if name == "OwnedSpWeapon" {
		return "owned_spweapons", nil
	}

	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if unicode.IsLower(prev) || nextIsLower {
					b.WriteRune('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}

	snake := b.String()
	if strings.HasSuffix(snake, "y") {
		return snake[:len(snake)-1] + "ies", nil
	}
	if strings.HasSuffix(snake, "s") {
		return snake + "es", nil
	}
	return snake + "s", nil
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func TestCompositeSpWeaponSuccess(t *testing.T) {
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.OwnedSpWeapon{})
	clearTable(t, &orm.Commander{})
	if err := orm.CreateCommanderRoot(1, 1, "SpWeapon Commander", 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 1}
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
	if err := orm.CreateCommanderRoot(1, 1, "SpWeapon Commander", 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 1}
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
