package answer

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

var errThemeTemplateLimit = errors.New("theme template limit exceeded")

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
	if err != nil {
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
	ctx := context.Background()
	if err := db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		return orm.UpsertBackyardCustomThemeTemplateTx(ctx, tx, commanderID, request.GetPos(), request.GetName(), b, request.GetIconImageMd5(), request.GetImageMd5())
	}); err != nil {
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

	ctx := context.Background()
	if err := db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		var entry orm.BackyardCustomThemeTemplate
		row := tx.QueryRow(ctx, `
SELECT commander_id, pos, name, furniture_put_list, icon_image_md5, image_md5, upload_time
FROM backyard_custom_theme_templates
WHERE commander_id = $1 AND pos = $2
`, int64(commanderID), int64(pos))
		if err := row.Scan(&entry.CommanderID, &entry.Pos, &entry.Name, &entry.FurniturePutList, &entry.IconImageMd5, &entry.ImageMd5, &entry.UploadTime); err != nil {
			return err
		}
		if entry.UploadTime == 0 {
			var publishedCount int64
			if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM backyard_custom_theme_templates WHERE commander_id = $1 AND upload_time > 0`, int64(commanderID)).Scan(&publishedCount); err != nil {
				return err
			}
			if publishedCount >= 2 {
				return errThemeTemplateLimit
			}
		}
		version, err := orm.CreateBackyardPublishedThemeVersionTx(ctx, tx, commanderID, pos, entry.Name, entry.FurniturePutList, entry.IconImageMd5, entry.ImageMd5)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE backyard_custom_theme_templates SET upload_time = $3 WHERE commander_id = $1 AND pos = $2`, int64(commanderID), int64(pos), int64(version.UploadTime))
		return err
	}); err != nil {
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
	themeID := orm.BackyardThemeID(commanderID, pos)
	ctx := context.Background()
	if err := db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE backyard_custom_theme_templates SET upload_time = 0 WHERE commander_id = $1 AND pos = $2`, int64(commanderID), int64(pos))
		if err != nil {
			return err
		}
		return orm.DeleteBackyardPublishedThemeVersionsByThemeIDTx(ctx, tx, themeID)
	}); err != nil {
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
	ctx := context.Background()
	if err := db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if err := orm.DeleteBackyardCustomThemeTemplateTx(ctx, tx, commanderID, pos); err != nil {
			return err
		}
		return orm.DeleteBackyardPublishedThemeVersionsByThemeIDTx(ctx, tx, orm.BackyardThemeID(commanderID, pos))
	}); err != nil {
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
	rows, err := db.DefaultStore.Pool.Query(context.Background(), `
SELECT commander_id, theme_id, upload_time
FROM backyard_theme_collections
WHERE commander_id = $1
ORDER BY upload_time DESC
`, int64(commanderID))
	if err != nil {
		resp := protobuf.SC_19116{Result: proto.Int32(1)}
		return client.SendMessage(19116, &resp)
	}
	defer rows.Close()
	entries := make([]orm.BackyardThemeCollection, 0)
	for rows.Next() {
		var entry orm.BackyardThemeCollection
		if err := rows.Scan(&entry.CommanderID, &entry.ThemeID, &entry.UploadTime); err != nil {
			resp := protobuf.SC_19116{Result: proto.Int32(1)}
			return client.SendMessage(19116, &resp)
		}
		entries = append(entries, entry)
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
	_ = db.DefaultStore.Pool.QueryRow(context.Background(), `
SELECT EXISTS(
  SELECT 1 FROM backyard_theme_collections WHERE commander_id = $1 AND theme_id = $2 AND upload_time = $3
)
`, int64(commanderID), themeID, int64(ver.UploadTime)).Scan(&hasFav)
	_ = db.DefaultStore.Pool.QueryRow(context.Background(), `
SELECT EXISTS(
  SELECT 1 FROM backyard_theme_likes WHERE commander_id = $1 AND theme_id = $2 AND upload_time = $3
)
`, int64(commanderID), themeID, int64(ver.UploadTime)).Scan(&hasLike)
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
	ctx := context.Background()
	if err := db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, `
INSERT INTO backyard_theme_likes (commander_id, theme_id, upload_time)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
`, int64(commanderID), request.GetThemeId(), int64(request.GetUploadTime()))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 1 {
			_, err = tx.Exec(ctx, `
UPDATE backyard_published_theme_versions
SET like_count = like_count + 1
WHERE theme_id = $1 AND upload_time = $2
`, request.GetThemeId(), int64(request.GetUploadTime()))
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
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
	ctx := context.Background()
	if err := db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, `
INSERT INTO backyard_theme_collections (commander_id, theme_id, upload_time)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
`, int64(commanderID), request.GetThemeId(), int64(request.GetUploadTime()))
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return nil
		}
		var count int64
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM backyard_theme_collections WHERE commander_id = $1`, int64(commanderID)).Scan(&count); err != nil {
			return err
		}
		if count > 30 {
			return errThemeTemplateLimit
		}
		_, err = tx.Exec(ctx, `
UPDATE backyard_published_theme_versions
SET fav_count = fav_count + 1
WHERE theme_id = $1 AND upload_time = $2
`, request.GetThemeId(), int64(request.GetUploadTime()))
		return err
	}); err != nil {
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
	ctx := context.Background()
	if err := db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
SELECT upload_time
FROM backyard_theme_collections
WHERE commander_id = $1 AND theme_id = $2
`, int64(commanderID), request.GetThemeId())
		if err != nil {
			return err
		}
		uploadTimes := make([]uint32, 0)
		for rows.Next() {
			var uploadTime uint32
			if err := rows.Scan(&uploadTime); err != nil {
				rows.Close()
				return err
			}
			uploadTimes = append(uploadTimes, uploadTime)
		}
		rows.Close()
		if _, err := tx.Exec(ctx, `DELETE FROM backyard_theme_collections WHERE commander_id = $1 AND theme_id = $2`, int64(commanderID), request.GetThemeId()); err != nil {
			return err
		}
		for _, uploadTime := range uploadTimes {
			if _, err := tx.Exec(ctx, `
UPDATE backyard_published_theme_versions
SET fav_count = fav_count - 1
WHERE theme_id = $1 AND upload_time = $2 AND fav_count > 0
`, request.GetThemeId(), int64(uploadTime)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
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
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO backyard_theme_informs (
  reporter_id,
  target_id,
  target_name,
  theme_id,
  theme_name,
  reason,
  created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
`, int64(entry.ReporterID), int64(entry.TargetID), entry.TargetName, entry.ThemeID, entry.ThemeName, int64(entry.Reason), int64(entry.CreatedAt)); err != nil {
		resp := protobuf.SC_19130{Result: proto.Int32(1)}
		return client.SendMessage(19130, &resp)
	}
	resp := protobuf.SC_19130{Result: proto.Int32(0)}
	return client.SendMessage(19130, &resp)
}
