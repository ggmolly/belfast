package answer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func newThemeTestClient(t *testing.T) *connection.Client {
	commanderID := uint32(time.Now().UnixNano())
	name := fmt.Sprintf("Theme Commander %d", commanderID)
	if err := orm.CreateCommanderRoot(commanderID, commanderID, name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func decodeResponse(t *testing.T, client *connection.Client, expectedID int, message proto.Message) {
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
}

func TestLikeThemeIdempotent(t *testing.T) {
	client := newThemeTestClient(t)
	themeID := orm.BackyardThemeID(client.Commander.CommanderID, 1)
	uploadTime := uint32(time.Now().Unix())

	ver := orm.BackyardPublishedThemeVersion{ThemeID: themeID, UploadTime: uploadTime, OwnerID: client.Commander.CommanderID, Pos: 1, Name: "theme", FurniturePutList: []byte(`[]`)}
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_published_theme_versions (theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)", ver.ThemeID, int64(ver.UploadTime), int64(ver.OwnerID), int64(ver.Pos), ver.Name, `[]`, "", "", int64(0), int64(0))

	payload := &protobuf.CS_19121{ThemeId: proto.String(themeID), UploadTime: proto.Uint32(uploadTime)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.LikeTheme19121(&buf, client); err != nil {
		t.Fatalf("LikeTheme19121 failed: %v", err)
	}
	resp1 := &protobuf.SC_19122{}
	decodeResponse(t, client, 19122, resp1)
	if resp1.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp1.GetResult())
	}

	client.Buffer.Reset()
	if _, _, err := answer.LikeTheme19121(&buf, client); err != nil {
		t.Fatalf("LikeTheme19121 (repeat) failed: %v", err)
	}
	resp2 := &protobuf.SC_19122{}
	decodeResponse(t, client, 19122, resp2)
	if resp2.GetResult() != 0 {
		t.Fatalf("expected repeat result 0, got %d", resp2.GetResult())
	}

	stored := queryAnswerExternalTestInt64(t, "SELECT like_count FROM backyard_published_theme_versions WHERE theme_id = $1 AND upload_time = $2", themeID, int64(uploadTime))
	if stored != 1 {
		t.Fatalf("expected like_count=1, got %d", stored)
	}
	likes := queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM backyard_theme_likes WHERE commander_id = $1 AND theme_id = $2 AND upload_time = $3", int64(client.Commander.CommanderID), themeID, int64(uploadTime))
	if likes != 1 {
		t.Fatalf("expected 1 like row, got %d", likes)
	}
}

func TestCollectThemeIdempotent(t *testing.T) {
	client := newThemeTestClient(t)
	themeID := orm.BackyardThemeID(client.Commander.CommanderID, 1)
	uploadTime := uint32(time.Now().Unix())

	ver := orm.BackyardPublishedThemeVersion{ThemeID: themeID, UploadTime: uploadTime, OwnerID: client.Commander.CommanderID, Pos: 1, Name: "theme", FurniturePutList: []byte(`[]`)}
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_published_theme_versions (theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)", ver.ThemeID, int64(ver.UploadTime), int64(ver.OwnerID), int64(ver.Pos), ver.Name, `[]`, "", "", int64(0), int64(0))

	payload := &protobuf.CS_19119{ThemeId: proto.String(themeID), UploadTime: proto.Uint32(uploadTime)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.CollectTheme19119(&buf, client); err != nil {
		t.Fatalf("CollectTheme19119 failed: %v", err)
	}
	resp1 := &protobuf.SC_19120{}
	decodeResponse(t, client, 19120, resp1)
	if resp1.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp1.GetResult())
	}

	client.Buffer.Reset()
	if _, _, err := answer.CollectTheme19119(&buf, client); err != nil {
		t.Fatalf("CollectTheme19119 (repeat) failed: %v", err)
	}
	resp2 := &protobuf.SC_19120{}
	decodeResponse(t, client, 19120, resp2)
	if resp2.GetResult() != 0 {
		t.Fatalf("expected repeat result 0, got %d", resp2.GetResult())
	}

	stored := queryAnswerExternalTestInt64(t, "SELECT fav_count FROM backyard_published_theme_versions WHERE theme_id = $1 AND upload_time = $2", themeID, int64(uploadTime))
	if stored != 1 {
		t.Fatalf("expected fav_count=1, got %d", stored)
	}
	collections := queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM backyard_theme_collections WHERE commander_id = $1 AND theme_id = $2 AND upload_time = $3", int64(client.Commander.CommanderID), themeID, int64(uploadTime))
	if collections != 1 {
		t.Fatalf("expected 1 collection row, got %d", collections)
	}
}

func TestCollectThemeCapAllowsExistingEntry(t *testing.T) {
	client := newThemeTestClient(t)
	commanderID := client.Commander.CommanderID
	uploadTime := uint32(time.Now().Unix())
	themeID := orm.BackyardThemeID(commanderID, 1)

	ver := orm.BackyardPublishedThemeVersion{ThemeID: themeID, UploadTime: uploadTime, OwnerID: commanderID, Pos: 1, Name: "theme", FurniturePutList: []byte(`[]`), FavCount: 1}
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_published_theme_versions (theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)", ver.ThemeID, int64(ver.UploadTime), int64(ver.OwnerID), int64(ver.Pos), ver.Name, `[]`, "", "", int64(0), int64(ver.FavCount))

	// Seed exactly 30 collections including the target.
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_theme_collections (commander_id, theme_id, upload_time) VALUES ($1, $2, $3)", int64(commanderID), themeID, int64(uploadTime))
	for i := uint32(2); i <= 30; i++ {
		otherID := orm.BackyardThemeID(commanderID, i)
		execAnswerExternalTestSQLT(t, "INSERT INTO backyard_theme_collections (commander_id, theme_id, upload_time) VALUES ($1, $2, $3)", int64(commanderID), otherID, int64(uploadTime))
	}

	count := queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM backyard_theme_collections WHERE commander_id = $1", int64(commanderID))
	if count != 30 {
		t.Fatalf("expected seeded collection count 30, got %d", count)
	}

	payload := &protobuf.CS_19119{ThemeId: proto.String(themeID), UploadTime: proto.Uint32(uploadTime)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.CollectTheme19119(&buf, client); err != nil {
		t.Fatalf("CollectTheme19119 failed: %v", err)
	}
	resp := &protobuf.SC_19120{}
	decodeResponse(t, client, 19120, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}

	stored := queryAnswerExternalTestInt64(t, "SELECT fav_count FROM backyard_published_theme_versions WHERE theme_id = $1 AND upload_time = $2", themeID, int64(uploadTime))
	if stored != 1 {
		t.Fatalf("expected fav_count to remain 1, got %d", stored)
	}

	count = queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM backyard_theme_collections WHERE commander_id = $1", int64(commanderID))
	if count != 30 {
		t.Fatalf("expected collection count to remain 30, got %d", count)
	}
}

func TestPublishThemeCapAllowsRepublish(t *testing.T) {
	client := newThemeTestClient(t)
	commanderID := client.Commander.CommanderID
	oldUpload := uint32(time.Now().Unix() - 10)

	// Seed two already-published templates.
	t1 := orm.BackyardCustomThemeTemplate{CommanderID: commanderID, Pos: 1, Name: "t1", FurniturePutList: []byte(`[]`), IconImageMd5: "", ImageMd5: "", UploadTime: oldUpload}
	t2 := orm.BackyardCustomThemeTemplate{CommanderID: commanderID, Pos: 2, Name: "t2", FurniturePutList: []byte(`[]`), IconImageMd5: "", ImageMd5: "", UploadTime: oldUpload}
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_custom_theme_templates (commander_id, pos, name, furniture_put_list, icon_image_md5, image_md5, upload_time) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7)", int64(t1.CommanderID), int64(t1.Pos), t1.Name, `[]`, t1.IconImageMd5, t1.ImageMd5, int64(t1.UploadTime))
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_custom_theme_templates (commander_id, pos, name, furniture_put_list, icon_image_md5, image_md5, upload_time) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7)", int64(t2.CommanderID), int64(t2.Pos), t2.Name, `[]`, t2.IconImageMd5, t2.ImageMd5, int64(t2.UploadTime))

	// Keep published versions consistent with the stored UploadTime.
	ver1 := orm.BackyardPublishedThemeVersion{ThemeID: orm.BackyardThemeID(commanderID, 1), UploadTime: oldUpload, OwnerID: commanderID, Pos: 1, Name: "t1", FurniturePutList: []byte(`[]`)}
	ver2 := orm.BackyardPublishedThemeVersion{ThemeID: orm.BackyardThemeID(commanderID, 2), UploadTime: oldUpload, OwnerID: commanderID, Pos: 2, Name: "t2", FurniturePutList: []byte(`[]`)}
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_published_theme_versions (theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)", ver1.ThemeID, int64(ver1.UploadTime), int64(ver1.OwnerID), int64(ver1.Pos), ver1.Name, `[]`, "", "", int64(0), int64(0))
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_published_theme_versions (theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)", ver2.ThemeID, int64(ver2.UploadTime), int64(ver2.OwnerID), int64(ver2.Pos), ver2.Name, `[]`, "", "", int64(0), int64(0))

	payload := &protobuf.CS_19111{Pos: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.PublishCustomThemeTemplate19111(&buf, client); err != nil {
		t.Fatalf("PublishCustomThemeTemplate19111 failed: %v", err)
	}
	resp := &protobuf.SC_19112{}
	decodeResponse(t, client, 19112, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}

	updated := queryAnswerExternalTestInt64(t, "SELECT upload_time FROM backyard_custom_theme_templates WHERE commander_id = $1 AND pos = $2", int64(commanderID), int64(1))
	if updated == 0 || uint32(updated) == oldUpload {
		t.Fatalf("expected upload_time to update on republish, got %d", updated)
	}

	publishedCount := queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM backyard_custom_theme_templates WHERE commander_id = $1 AND upload_time > 0", int64(commanderID))
	if publishedCount != 2 {
		t.Fatalf("expected published template count 2, got %d", publishedCount)
	}

	latest := queryAnswerExternalTestInt64(t, "SELECT upload_time FROM backyard_published_theme_versions WHERE theme_id = $1 ORDER BY upload_time DESC LIMIT 1", orm.BackyardThemeID(commanderID, 1))
	if latest != updated {
		t.Fatalf("expected latest version upload_time %d to match template upload_time %d", latest, updated)
	}
}

func TestCancelCollectThemeDecrements(t *testing.T) {
	client := newThemeTestClient(t)
	themeID := orm.BackyardThemeID(client.Commander.CommanderID, 1)
	uploadTime := uint32(time.Now().Unix())

	ver := orm.BackyardPublishedThemeVersion{ThemeID: themeID, UploadTime: uploadTime, OwnerID: client.Commander.CommanderID, Pos: 1, Name: "theme", FurniturePutList: []byte(`[]`)}
	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_published_theme_versions (theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)", ver.ThemeID, int64(ver.UploadTime), int64(ver.OwnerID), int64(ver.Pos), ver.Name, `[]`, "", "", int64(0), int64(0))

	collectPayload := &protobuf.CS_19119{ThemeId: proto.String(themeID), UploadTime: proto.Uint32(uploadTime)}
	collectBuf, err := proto.Marshal(collectPayload)
	if err != nil {
		t.Fatalf("failed to marshal collect payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.CollectTheme19119(&collectBuf, client); err != nil {
		t.Fatalf("CollectTheme19119 failed: %v", err)
	}
	decodeResponse(t, client, 19120, &protobuf.SC_19120{})

	cancelPayload := &protobuf.CS_19127{ThemeId: proto.String(themeID)}
	cancelBuf, err := proto.Marshal(cancelPayload)
	if err != nil {
		t.Fatalf("failed to marshal cancel payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.CancelCollectTheme19127(&cancelBuf, client); err != nil {
		t.Fatalf("CancelCollectTheme19127 failed: %v", err)
	}
	decodeResponse(t, client, 19128, &protobuf.SC_19128{})

	stored := queryAnswerExternalTestInt64(t, "SELECT fav_count FROM backyard_published_theme_versions WHERE theme_id = $1 AND upload_time = $2", themeID, int64(uploadTime))
	if stored != 0 {
		t.Fatalf("expected fav_count=0 after cancel, got %d", stored)
	}
	collections := queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM backyard_theme_collections WHERE commander_id = $1 AND theme_id = $2", int64(client.Commander.CommanderID), themeID)
	if collections != 0 {
		t.Fatalf("expected 0 collection rows after cancel, got %d", collections)
	}
}

func TestGetThemeListLegacy19107Success(t *testing.T) {
	client := newThemeTestClient(t)
	commanderID := client.Commander.CommanderID
	themeID := orm.BackyardThemeID(commanderID, 1)
	uploadTime := uint32(time.Now().Unix())

	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_published_theme_versions (theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)", themeID, int64(uploadTime), int64(commanderID), int64(1), "legacy-theme", `[{"id":"furn_1","x":1,"y":2,"dir":3,"child":[],"parent":0,"shipId":0}]`, "icon-md5", "image-md5", int64(4), int64(5))

	payload := &protobuf.CS_19107{Typ: proto.Int32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.GetThemeListLegacy19107(&buf, client); err != nil {
		t.Fatalf("GetThemeListLegacy19107 failed: %v", err)
	}
	resp := &protobuf.SC_19108{}
	decodeResponse(t, client, 19108, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}
	if len(resp.GetThemeList()) == 0 {
		t.Fatalf("expected at least one theme")
	}
	var theme *protobuf.DORMTHEME
	for _, candidate := range resp.GetThemeList() {
		if candidate.GetId() == themeID {
			theme = candidate
			break
		}
	}
	if theme == nil {
		t.Fatalf("expected theme id %s to be present", themeID)
	}
	if theme.GetId() != themeID {
		t.Fatalf("expected id %s, got %s", themeID, theme.GetId())
	}
	if theme.GetName() != "legacy-theme" {
		t.Fatalf("expected theme name legacy-theme, got %s", theme.GetName())
	}
	if theme.GetUserId() != commanderID {
		t.Fatalf("expected user id %d, got %d", commanderID, theme.GetUserId())
	}
	if theme.GetPos() != 1 {
		t.Fatalf("expected pos 1, got %d", theme.GetPos())
	}
	if theme.GetUploadTime() != uploadTime {
		t.Fatalf("expected upload_time %d, got %d", uploadTime, theme.GetUploadTime())
	}
	if theme.GetIconImageMd5() != "icon-md5" {
		t.Fatalf("expected icon md5 icon-md5, got %s", theme.GetIconImageMd5())
	}
	if len(theme.GetFurniturePutList()) != 1 {
		t.Fatalf("expected 1 furniture item, got %d", len(theme.GetFurniturePutList()))
	}
}

func TestGetThemeListLegacy19107UnsupportedType(t *testing.T) {
	client := newThemeTestClient(t)
	payload := &protobuf.CS_19107{Typ: proto.Int32(999)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.GetThemeListLegacy19107(&buf, client); err != nil {
		t.Fatalf("GetThemeListLegacy19107 failed: %v", err)
	}
	resp := &protobuf.SC_19108{}
	decodeResponse(t, client, 19108, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result for unsupported type")
	}
	if len(resp.GetThemeList()) != 0 {
		t.Fatalf("expected empty theme list for unsupported type, got %d", len(resp.GetThemeList()))
	}
}

func TestGetThemeListLegacy19107UnmarshalError(t *testing.T) {
	client := newThemeTestClient(t)
	buf := []byte{0xff, 0x00, 0x42}
	_, outID, err := answer.GetThemeListLegacy19107(&buf, client)
	if err == nil {
		t.Fatalf("expected error")
	}
	if outID != 19108 {
		t.Fatalf("expected outgoing packet id 19108, got %d", outID)
	}
}

func TestGetThemeShopList19117Regression(t *testing.T) {
	client := newThemeTestClient(t)
	commanderID := client.Commander.CommanderID
	uploadTime := uint32(time.Now().Unix())
	themeID := orm.BackyardThemeID(commanderID, 2)

	execAnswerExternalTestSQLT(t, "INSERT INTO backyard_published_theme_versions (theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)", themeID, int64(uploadTime), int64(commanderID), int64(2), "shop-theme", `[]`, "", "", int64(0), int64(0))

	payload := &protobuf.CS_19117{Typ: proto.Int32(1), Page: proto.Int32(1), Num: proto.Int32(10)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.GetThemeShopList19117(&buf, client); err != nil {
		t.Fatalf("GetThemeShopList19117 failed: %v", err)
	}
	resp := &protobuf.SC_19118{}
	decodeResponse(t, client, 19118, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}
	found := false
	for _, id := range resp.GetThemeIdList() {
		if id == themeID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected theme id %s in list, got %v", themeID, resp.GetThemeIdList())
	}
}
