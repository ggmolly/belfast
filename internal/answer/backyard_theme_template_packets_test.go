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
	commander := orm.Commander{CommanderID: commanderID, AccountID: commanderID, Name: fmt.Sprintf("Theme Commander %d", commanderID)}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
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
	if err := orm.GormDB.Create(&ver).Error; err != nil {
		t.Fatalf("failed to create published theme version: %v", err)
	}

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

	var stored orm.BackyardPublishedThemeVersion
	if err := orm.GormDB.Where("theme_id = ? AND upload_time = ?", themeID, uploadTime).First(&stored).Error; err != nil {
		t.Fatalf("failed to reload published theme version: %v", err)
	}
	if stored.LikeCount != 1 {
		t.Fatalf("expected like_count=1, got %d", stored.LikeCount)
	}
	var likes int64
	if err := orm.GormDB.Model(&orm.BackyardThemeLike{}).Where("commander_id = ? AND theme_id = ? AND upload_time = ?", client.Commander.CommanderID, themeID, uploadTime).Count(&likes).Error; err != nil {
		t.Fatalf("failed to count likes: %v", err)
	}
	if likes != 1 {
		t.Fatalf("expected 1 like row, got %d", likes)
	}
}

func TestCollectThemeIdempotent(t *testing.T) {
	client := newThemeTestClient(t)
	themeID := orm.BackyardThemeID(client.Commander.CommanderID, 1)
	uploadTime := uint32(time.Now().Unix())

	ver := orm.BackyardPublishedThemeVersion{ThemeID: themeID, UploadTime: uploadTime, OwnerID: client.Commander.CommanderID, Pos: 1, Name: "theme", FurniturePutList: []byte(`[]`)}
	if err := orm.GormDB.Create(&ver).Error; err != nil {
		t.Fatalf("failed to create published theme version: %v", err)
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

	var stored orm.BackyardPublishedThemeVersion
	if err := orm.GormDB.Where("theme_id = ? AND upload_time = ?", themeID, uploadTime).First(&stored).Error; err != nil {
		t.Fatalf("failed to reload published theme version: %v", err)
	}
	if stored.FavCount != 1 {
		t.Fatalf("expected fav_count=1, got %d", stored.FavCount)
	}
	var collections int64
	if err := orm.GormDB.Model(&orm.BackyardThemeCollection{}).Where("commander_id = ? AND theme_id = ? AND upload_time = ?", client.Commander.CommanderID, themeID, uploadTime).Count(&collections).Error; err != nil {
		t.Fatalf("failed to count collections: %v", err)
	}
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
	if err := orm.GormDB.Create(&ver).Error; err != nil {
		t.Fatalf("failed to create published theme version: %v", err)
	}

	// Seed exactly 30 collections including the target.
	if err := orm.GormDB.Create(&orm.BackyardThemeCollection{CommanderID: commanderID, ThemeID: themeID, UploadTime: uploadTime}).Error; err != nil {
		t.Fatalf("failed to create existing collection row: %v", err)
	}
	for i := uint32(2); i <= 30; i++ {
		otherID := orm.BackyardThemeID(commanderID, i)
		if err := orm.GormDB.Create(&orm.BackyardThemeCollection{CommanderID: commanderID, ThemeID: otherID, UploadTime: uploadTime}).Error; err != nil {
			t.Fatalf("failed to create collection seed row %d: %v", i, err)
		}
	}

	var count int64
	if err := orm.GormDB.Model(&orm.BackyardThemeCollection{}).Where("commander_id = ?", commanderID).Count(&count).Error; err != nil {
		t.Fatalf("failed to count collections: %v", err)
	}
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

	var stored orm.BackyardPublishedThemeVersion
	if err := orm.GormDB.Where("theme_id = ? AND upload_time = ?", themeID, uploadTime).First(&stored).Error; err != nil {
		t.Fatalf("failed to reload published theme version: %v", err)
	}
	if stored.FavCount != 1 {
		t.Fatalf("expected fav_count to remain 1, got %d", stored.FavCount)
	}

	if err := orm.GormDB.Model(&orm.BackyardThemeCollection{}).Where("commander_id = ?", commanderID).Count(&count).Error; err != nil {
		t.Fatalf("failed to recount collections: %v", err)
	}
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
	if err := orm.GormDB.Create(&t1).Error; err != nil {
		t.Fatalf("failed to create template 1: %v", err)
	}
	if err := orm.GormDB.Create(&t2).Error; err != nil {
		t.Fatalf("failed to create template 2: %v", err)
	}

	// Keep published versions consistent with the stored UploadTime.
	ver1 := orm.BackyardPublishedThemeVersion{ThemeID: orm.BackyardThemeID(commanderID, 1), UploadTime: oldUpload, OwnerID: commanderID, Pos: 1, Name: "t1", FurniturePutList: []byte(`[]`)}
	ver2 := orm.BackyardPublishedThemeVersion{ThemeID: orm.BackyardThemeID(commanderID, 2), UploadTime: oldUpload, OwnerID: commanderID, Pos: 2, Name: "t2", FurniturePutList: []byte(`[]`)}
	if err := orm.GormDB.Create(&ver1).Error; err != nil {
		t.Fatalf("failed to create published version 1: %v", err)
	}
	if err := orm.GormDB.Create(&ver2).Error; err != nil {
		t.Fatalf("failed to create published version 2: %v", err)
	}

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

	var updated orm.BackyardCustomThemeTemplate
	if err := orm.GormDB.Where("commander_id = ? AND pos = ?", commanderID, 1).First(&updated).Error; err != nil {
		t.Fatalf("failed to reload updated template: %v", err)
	}
	if updated.UploadTime == 0 || updated.UploadTime == oldUpload {
		t.Fatalf("expected upload_time to update on republish, got %d", updated.UploadTime)
	}

	var publishedCount int64
	if err := orm.GormDB.Model(&orm.BackyardCustomThemeTemplate{}).
		Where("commander_id = ? AND upload_time > 0", commanderID).
		Count(&publishedCount).Error; err != nil {
		t.Fatalf("failed to count published templates: %v", err)
	}
	if publishedCount != 2 {
		t.Fatalf("expected published template count 2, got %d", publishedCount)
	}

	var latest orm.BackyardPublishedThemeVersion
	if err := orm.GormDB.Where("theme_id = ?", orm.BackyardThemeID(commanderID, 1)).Order("upload_time desc").First(&latest).Error; err != nil {
		t.Fatalf("failed to reload latest published version: %v", err)
	}
	if latest.UploadTime != updated.UploadTime {
		t.Fatalf("expected latest version upload_time %d to match template upload_time %d", latest.UploadTime, updated.UploadTime)
	}
}

func TestCancelCollectThemeDecrements(t *testing.T) {
	client := newThemeTestClient(t)
	themeID := orm.BackyardThemeID(client.Commander.CommanderID, 1)
	uploadTime := uint32(time.Now().Unix())

	ver := orm.BackyardPublishedThemeVersion{ThemeID: themeID, UploadTime: uploadTime, OwnerID: client.Commander.CommanderID, Pos: 1, Name: "theme", FurniturePutList: []byte(`[]`)}
	if err := orm.GormDB.Create(&ver).Error; err != nil {
		t.Fatalf("failed to create published theme version: %v", err)
	}

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

	var stored orm.BackyardPublishedThemeVersion
	if err := orm.GormDB.Where("theme_id = ? AND upload_time = ?", themeID, uploadTime).First(&stored).Error; err != nil {
		t.Fatalf("failed to reload published theme version: %v", err)
	}
	if stored.FavCount != 0 {
		t.Fatalf("expected fav_count=0 after cancel, got %d", stored.FavCount)
	}
	var collections int64
	if err := orm.GormDB.Model(&orm.BackyardThemeCollection{}).Where("commander_id = ? AND theme_id = ?", client.Commander.CommanderID, themeID).Count(&collections).Error; err != nil {
		t.Fatalf("failed to count collections: %v", err)
	}
	if collections != 0 {
		t.Fatalf("expected 0 collection rows after cancel, got %d", collections)
	}
}
