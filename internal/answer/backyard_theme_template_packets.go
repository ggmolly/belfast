package answer

import (
	"encoding/json"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetOSSArgs19103(buffer *[]byte, client *connection.Client) (int, int, error) {
	// TODO: integrate real OSS/S3 creds for publishing previews.
	// For now, return success with empty credentials.
	resp := protobuf.SC_19104{
		Result:        proto.Uint32(0),
		AccessId:      proto.String(""),
		AccessSecret:  proto.String(""),
		ExpireTime:    proto.Uint32(0),
		SecurityToken: proto.String(""),
	}
	return client.SendMessage(19104, &resp)
}

func GetCustomThemeTemplates19105(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19105
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19106, err
	}
	commanderID := client.Commander.CommanderID
	entries, err := orm.ListBackyardCustomThemeTemplates(commanderID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, 19106, err
	}
	themes := make([]*protobuf.DORMTHEME, 0, len(entries))
	for _, e := range entries {
		var stored []storedFurniturePut
		_ = json.Unmarshal(e.FurniturePutList, &stored)
		putList := make([]*protobuf.FURNITUREPUTINFO, 0, len(stored))
		for _, f := range stored {
			children := make([]*protobuf.CHILDINFO, 0, len(f.Child))
			for _, c := range f.Child {
				children = append(children, &protobuf.CHILDINFO{Id: proto.String(c.Id), X: proto.Uint32(c.X), Y: proto.Uint32(c.Y)})
			}
			putList = append(putList, &protobuf.FURNITUREPUTINFO{Id: proto.String(f.Id), X: proto.Uint32(f.X), Y: proto.Uint32(f.Y), Dir: proto.Uint32(f.Dir), Child: children, Parent: proto.Uint64(f.Parent), ShipId: proto.Uint32(f.ShipId)})
		}
		themes = append(themes, &protobuf.DORMTHEME{
			Id:               proto.String(orm.BackyardThemeID(commanderID, e.Pos)),
			Name:             proto.String(e.Name),
			FurniturePutList: putList,
			UserId:           proto.Uint32(commanderID),
			Pos:              proto.Uint32(e.Pos),
			LikeCount:        proto.Uint32(0),
			FavCount:         proto.Uint32(0),
			UploadTime:       proto.Uint32(e.UploadTime),
			IconImageMd5:     proto.String(e.IconImageMd5),
			ImageMd5:         proto.String(e.ImageMd5),
		})
	}
	resp := protobuf.SC_19106{Result: proto.Uint32(0), ThemeList: themes}
	return client.SendMessage(19106, &resp)
}

func SaveCustomThemeTemplate19109(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19109
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19110, err
	}
	commanderID := client.Commander.CommanderID
	state, err := orm.GetOrCreateCommanderDormState(commanderID)
	if err != nil {
		return 0, 19110, err
	}
	mapSize := dormStaticMapSize(state.Level)
	if err := validateFurniturePutList(request.GetFurniturePutList(), 1, mapSize); err != nil {
		resp := protobuf.SC_19110{Result: proto.Int32(1)}
		return client.SendMessage(19110, &resp)
	}
	stored := make([]storedFurniturePut, 0, len(request.GetFurniturePutList()))
	for _, f := range request.GetFurniturePutList() {
		children := make([]storedChild, 0, len(f.GetChild()))
		for _, c := range f.GetChild() {
			children = append(children, storedChild{Id: c.GetId(), X: c.GetX(), Y: c.GetY()})
		}
		stored = append(stored, storedFurniturePut{Id: f.GetId(), X: f.GetX(), Y: f.GetY(), Dir: f.GetDir(), Child: children, Parent: f.GetParent(), ShipId: f.GetShipId()})
	}
	b, err := json.Marshal(stored)
	if err != nil {
		return 0, 19110, err
	}
	tx := orm.GormDB.Begin()
	if err := orm.UpsertBackyardCustomThemeTemplateTx(tx, commanderID, request.GetPos(), request.GetName(), b, request.GetIconImageMd5(), request.GetImageMd5()); err != nil {
		tx.Rollback()
		resp := protobuf.SC_19110{Result: proto.Int32(1)}
		return client.SendMessage(19110, &resp)
	}
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19110{Result: proto.Int32(1)}
		return client.SendMessage(19110, &resp)
	}
	resp := protobuf.SC_19110{Result: proto.Int32(0)}
	return client.SendMessage(19110, &resp)
}

func PublishCustomThemeTemplate19111(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19111
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19112, err
	}
	commanderID := client.Commander.CommanderID
	pos := request.GetPos()

	tx := orm.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	var entry orm.BackyardCustomThemeTemplate
	if err := tx.Where("commander_id = ? AND pos = ?", commanderID, pos).First(&entry).Error; err != nil {
		tx.Rollback()
		resp := protobuf.SC_19112{Result: proto.Int32(1)}
		return client.SendMessage(19112, &resp)
	}
	// Cap applies only when publishing a previously-unpublished template.
	if entry.UploadTime == 0 {
		var publishedCount int64
		if err := tx.Model(&orm.BackyardCustomThemeTemplate{}).
			Where("commander_id = ? AND upload_time > 0", commanderID).
			Count(&publishedCount).Error; err != nil {
			tx.Rollback()
			return 0, 19112, err
		}
		if publishedCount >= 2 {
			tx.Rollback()
			resp := protobuf.SC_19112{Result: proto.Int32(1)}
			return client.SendMessage(19112, &resp)
		}
	}
	version, err := orm.CreateBackyardPublishedThemeVersionTx(tx, commanderID, pos, entry.Name, entry.FurniturePutList, entry.IconImageMd5, entry.ImageMd5)
	if err != nil {
		tx.Rollback()
		resp := protobuf.SC_19112{Result: proto.Int32(1)}
		return client.SendMessage(19112, &resp)
	}
	entry.UploadTime = version.UploadTime
	if err := tx.Save(entry).Error; err != nil {
		tx.Rollback()
		resp := protobuf.SC_19112{Result: proto.Int32(1)}
		return client.SendMessage(19112, &resp)
	}
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19112{Result: proto.Int32(1)}
		return client.SendMessage(19112, &resp)
	}
	resp := protobuf.SC_19112{Result: proto.Int32(0)}
	return client.SendMessage(19112, &resp)
}

func UnpublishCustomThemeTemplate19125(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19125
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19126, err
	}
	commanderID := client.Commander.CommanderID
	pos := request.GetPos()
	entry, err := orm.GetBackyardCustomThemeTemplate(commanderID, pos)
	if err != nil {
		resp := protobuf.SC_19126{Result: proto.Int32(1)}
		return client.SendMessage(19126, &resp)
	}
	if entry.UploadTime == 0 {
		resp := protobuf.SC_19126{Result: proto.Int32(0)}
		return client.SendMessage(19126, &resp)
	}
	tx := orm.GormDB.Begin()
	themeID := orm.BackyardThemeID(commanderID, pos)
	entry.UploadTime = 0
	if err := tx.Save(entry).Error; err != nil {
		tx.Rollback()
		resp := protobuf.SC_19126{Result: proto.Int32(1)}
		return client.SendMessage(19126, &resp)
	}
	_ = orm.DeleteBackyardPublishedThemeVersionsByThemeIDTx(tx, themeID)
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19126{Result: proto.Int32(1)}
		return client.SendMessage(19126, &resp)
	}
	resp := protobuf.SC_19126{Result: proto.Int32(0)}
	return client.SendMessage(19126, &resp)
}

func DeleteCustomThemeTemplate19123(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19123
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19124, err
	}
	commanderID := client.Commander.CommanderID
	pos := request.GetPos()
	tx := orm.GormDB.Begin()
	_ = orm.DeleteBackyardCustomThemeTemplateTx(tx, commanderID, pos)
	_ = orm.DeleteBackyardPublishedThemeVersionsByThemeIDTx(tx, orm.BackyardThemeID(commanderID, pos))
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19124{Result: proto.Int32(1)}
		return client.SendMessage(19124, &resp)
	}
	resp := protobuf.SC_19124{Result: proto.Int32(0)}
	return client.SendMessage(19124, &resp)
}

func GetThemeShopList19117(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19117
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19118, err
	}
	ids, err := orm.ListBackyardPublishedThemeIDsByPage(uint32(request.GetPage()), uint32(request.GetNum()))
	if err != nil {
		resp := protobuf.SC_19118{Result: proto.Int32(1)}
		return client.SendMessage(19118, &resp)
	}
	resp := protobuf.SC_19118{Result: proto.Int32(0), ThemeIdList: ids}
	return client.SendMessage(19118, &resp)
}

func GetCollectionList19115(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19115
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19116, err
	}
	_ = request
	commanderID := client.Commander.CommanderID
	var entries []orm.BackyardThemeCollection
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Order("upload_time desc").Find(&entries).Error; err != nil {
		resp := protobuf.SC_19116{Result: proto.Int32(1)}
		return client.SendMessage(19116, &resp)
	}
	profiles := make([]*protobuf.DORMTHEME_PROFILE, 0, len(entries))
	for _, e := range entries {
		profiles = append(profiles, &protobuf.DORMTHEME_PROFILE{Id: proto.String(e.ThemeID), UploadTime: proto.Uint32(e.UploadTime)})
	}
	resp := protobuf.SC_19116{Result: proto.Int32(0), ThemeProfileList: profiles}
	return client.SendMessage(19116, &resp)
}

func GetPreviewMd5s19131(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19131
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19132, err
	}
	list := make([]*protobuf.THEME_MD5, 0, len(request.GetIdList()))
	for _, id := range request.GetIdList() {
		ver, err := orm.LatestBackyardPublishedThemeVersion(id)
		if err != nil {
			continue
		}
		list = append(list, &protobuf.THEME_MD5{Id: proto.String(id), Md5: proto.String(ver.IconImageMd5)})
	}
	resp := protobuf.SC_19132{List: list}
	return client.SendMessage(19132, &resp)
}

func SearchTheme19113(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19113
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19114, err
	}
	commanderID := client.Commander.CommanderID
	themeID := request.GetThemeId()
	ver, err := orm.LatestBackyardPublishedThemeVersion(themeID)
	if err != nil {
		resp := protobuf.SC_19114{Result: proto.Int32(20)}
		return client.SendMessage(19114, &resp)
	}
	var stored []storedFurniturePut
	_ = json.Unmarshal(ver.FurniturePutList, &stored)
	putList := make([]*protobuf.FURNITUREPUTINFO, 0, len(stored))
	for _, f := range stored {
		children := make([]*protobuf.CHILDINFO, 0, len(f.Child))
		for _, c := range f.Child {
			children = append(children, &protobuf.CHILDINFO{Id: proto.String(c.Id), X: proto.Uint32(c.X), Y: proto.Uint32(c.Y)})
		}
		putList = append(putList, &protobuf.FURNITUREPUTINFO{Id: proto.String(f.Id), X: proto.Uint32(f.X), Y: proto.Uint32(f.Y), Dir: proto.Uint32(f.Dir), Child: children, Parent: proto.Uint64(f.Parent), ShipId: proto.Uint32(f.ShipId)})
	}
	var hasFav bool
	var hasLike bool
	var fav orm.BackyardThemeCollection
	if err := orm.GormDB.Where("commander_id = ? AND theme_id = ? AND upload_time = ?", commanderID, themeID, ver.UploadTime).First(&fav).Error; err == nil {
		hasFav = true
	}
	var like orm.BackyardThemeLike
	if err := orm.GormDB.Where("commander_id = ? AND theme_id = ? AND upload_time = ?", commanderID, themeID, ver.UploadTime).First(&like).Error; err == nil {
		hasLike = true
	}
	resp := protobuf.SC_19114{
		Result: proto.Int32(0),
		Theme: &protobuf.DORMTHEME{
			Id:               proto.String(ver.ThemeID),
			Name:             proto.String(ver.Name),
			FurniturePutList: putList,
			UserId:           proto.Uint32(ver.OwnerID),
			Pos:              proto.Uint32(ver.Pos),
			LikeCount:        proto.Uint32(ver.LikeCount),
			FavCount:         proto.Uint32(ver.FavCount),
			UploadTime:       proto.Uint32(ver.UploadTime),
			IconImageMd5:     proto.String(ver.IconImageMd5),
			ImageMd5:         proto.String(ver.ImageMd5),
		},
		HasFav:  proto.Bool(hasFav),
		HasLike: proto.Bool(hasLike),
	}
	return client.SendMessage(19114, &resp)
}

func LikeTheme19121(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19121
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19122, err
	}
	commanderID := client.Commander.CommanderID
	tx := orm.GormDB.Begin()
	entry := orm.BackyardThemeLike{CommanderID: commanderID, ThemeID: request.GetThemeId(), UploadTime: request.GetUploadTime()}
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry)
	if err := res.Error; err != nil {
		tx.Rollback()
		resp := protobuf.SC_19122{Result: proto.Int32(1)}
		return client.SendMessage(19122, &resp)
	}
	// Only increment if we inserted a new like.
	if res.RowsAffected == 1 {
		if err := tx.Model(&orm.BackyardPublishedThemeVersion{}).
			Where("theme_id = ? AND upload_time = ?", request.GetThemeId(), request.GetUploadTime()).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			tx.Rollback()
			resp := protobuf.SC_19122{Result: proto.Int32(1)}
			return client.SendMessage(19122, &resp)
		}
	}
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19122{Result: proto.Int32(1)}
		return client.SendMessage(19122, &resp)
	}
	resp := protobuf.SC_19122{Result: proto.Int32(0)}
	return client.SendMessage(19122, &resp)
}

func CollectTheme19119(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19119
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19120, err
	}
	commanderID := client.Commander.CommanderID
	tx := orm.GormDB.Begin()
	entry := orm.BackyardThemeCollection{CommanderID: commanderID, ThemeID: request.GetThemeId(), UploadTime: request.GetUploadTime()}
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry)
	if err := res.Error; err != nil {
		tx.Rollback()
		resp := protobuf.SC_19120{Result: proto.Int32(1)}
		return client.SendMessage(19120, &resp)
	}
	// If the row already exists, treat as success (idempotent) and do not apply the cap.
	if res.RowsAffected == 0 {
		if err := tx.Commit().Error; err != nil {
			resp := protobuf.SC_19120{Result: proto.Int32(1)}
			return client.SendMessage(19120, &resp)
		}
		resp := protobuf.SC_19120{Result: proto.Int32(0)}
		return client.SendMessage(19120, &resp)
	}
	// Cap 30 (only enforced for a new collection row).
	var count int64
	if err := tx.Model(&orm.BackyardThemeCollection{}).Where("commander_id = ?", commanderID).Count(&count).Error; err != nil {
		tx.Rollback()
		resp := protobuf.SC_19120{Result: proto.Int32(1)}
		return client.SendMessage(19120, &resp)
	}
	if count > 30 {
		tx.Rollback()
		resp := protobuf.SC_19120{Result: proto.Int32(1)}
		return client.SendMessage(19120, &resp)
	}
	if res.RowsAffected == 1 {
		if err := tx.Model(&orm.BackyardPublishedThemeVersion{}).
			Where("theme_id = ? AND upload_time = ?", request.GetThemeId(), request.GetUploadTime()).
			UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error; err != nil {
			tx.Rollback()
			resp := protobuf.SC_19120{Result: proto.Int32(1)}
			return client.SendMessage(19120, &resp)
		}
	}
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19120{Result: proto.Int32(1)}
		return client.SendMessage(19120, &resp)
	}
	resp := protobuf.SC_19120{Result: proto.Int32(0)}
	return client.SendMessage(19120, &resp)
}

func CancelCollectTheme19127(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19127
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19128, err
	}
	commanderID := client.Commander.CommanderID
	tx := orm.GormDB.Begin()
	var entries []orm.BackyardThemeCollection
	_ = tx.Where("commander_id = ? AND theme_id = ?", commanderID, request.GetThemeId()).Find(&entries).Error
	_ = tx.Where("commander_id = ? AND theme_id = ?", commanderID, request.GetThemeId()).Delete(&orm.BackyardThemeCollection{}).Error
	for _, e := range entries {
		_ = tx.Model(&orm.BackyardPublishedThemeVersion{}).
			Where("theme_id = ? AND upload_time = ? AND fav_count > 0", e.ThemeID, e.UploadTime).
			UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
	}
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19128{Result: proto.Int32(1)}
		return client.SendMessage(19128, &resp)
	}
	resp := protobuf.SC_19128{Result: proto.Int32(0)}
	return client.SendMessage(19128, &resp)
}

func InformTheme19129(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19129
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19130, err
	}
	entry := orm.BackyardThemeInform{
		ReporterID: client.Commander.CommanderID,
		TargetID:   request.GetTargetId(),
		TargetName: request.GetTargetName(),
		ThemeID:    request.GetThemeId(),
		ThemeName:  request.GetThemeName(),
		Reason:     request.GetReason(),
		CreatedAt:  uint32(time.Now().Unix()),
	}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		resp := protobuf.SC_19130{Result: proto.Int32(1)}
		return client.SendMessage(19130, &resp)
	}
	resp := protobuf.SC_19130{Result: proto.Int32(0)}
	return client.SendMessage(19130, &resp)
}
