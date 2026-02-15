package orm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

func BackyardThemeID(commanderID uint32, pos uint32) string {
	// Client builds theme id via string concatenation: player.id .. pos
	return fmt.Sprintf("%d%d", commanderID, pos)
}

// BackyardCustomThemeTemplate stores the per-commander editable template (pos 1..N).
// UploadTime==0 means not published.
type BackyardCustomThemeTemplate struct {
	CommanderID      uint32          `json:"commander_id"`
	Pos              uint32          `json:"pos"`
	Name             string          `json:"name"`
	FurniturePutList json.RawMessage `json:"furniture_put_list"`
	IconImageMd5     string          `json:"icon_image_md5"`
	ImageMd5         string          `json:"image_md5"`
	UploadTime       uint32          `json:"upload_time"`
}

func UpsertBackyardCustomThemeTemplateTx(ctx context.Context, tx pgx.Tx, commanderID uint32, pos uint32, name string, furniturePutList json.RawMessage, iconMd5, imageMd5 string) error {
	_, err := tx.Exec(ctx, `
INSERT INTO backyard_custom_theme_templates (
  commander_id,
  pos,
  name,
  furniture_put_list,
  icon_image_md5,
  image_md5,
  upload_time
) VALUES (
  $1, $2, $3, $4, $5, $6, 0
)
ON CONFLICT (commander_id, pos)
DO UPDATE SET
  name = EXCLUDED.name,
  furniture_put_list = EXCLUDED.furniture_put_list,
  icon_image_md5 = EXCLUDED.icon_image_md5,
  image_md5 = EXCLUDED.image_md5
`, int64(commanderID), int64(pos), name, furniturePutList, iconMd5, imageMd5)
	return err
}

func ListBackyardCustomThemeTemplates(commanderID uint32) ([]BackyardCustomThemeTemplate, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT commander_id, pos, name, furniture_put_list, icon_image_md5, image_md5, upload_time
FROM backyard_custom_theme_templates
WHERE commander_id = $1
ORDER BY pos ASC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]BackyardCustomThemeTemplate, 0)
	for rows.Next() {
		var entry BackyardCustomThemeTemplate
		if err := rows.Scan(&entry.CommanderID, &entry.Pos, &entry.Name, &entry.FurniturePutList, &entry.IconImageMd5, &entry.ImageMd5, &entry.UploadTime); err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, rows.Err()
}

func GetBackyardCustomThemeTemplate(commanderID uint32, pos uint32) (*BackyardCustomThemeTemplate, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, pos, name, furniture_put_list, icon_image_md5, image_md5, upload_time
FROM backyard_custom_theme_templates
WHERE commander_id = $1 AND pos = $2
`, int64(commanderID), int64(pos))
	var entry BackyardCustomThemeTemplate
	if err := row.Scan(&entry.CommanderID, &entry.Pos, &entry.Name, &entry.FurniturePutList, &entry.IconImageMd5, &entry.ImageMd5, &entry.UploadTime); err != nil {
		return nil, err
	}
	return &entry, nil
}

func DeleteBackyardCustomThemeTemplateTx(ctx context.Context, tx pgx.Tx, commanderID uint32, pos uint32) error {
	_, err := tx.Exec(ctx, `DELETE FROM backyard_custom_theme_templates WHERE commander_id = $1 AND pos = $2`, int64(commanderID), int64(pos))
	return err
}

// BackyardPublishedThemeVersion stores an uploaded version of a theme.
// ThemeID is the stable id (player.id .. pos), UploadTime disambiguates versions.
type BackyardPublishedThemeVersion struct {
	ThemeID          string          `json:"theme_id"`
	UploadTime       uint32          `json:"upload_time"`
	OwnerID          uint32          `json:"owner_id"`
	Pos              uint32          `json:"pos"`
	Name             string          `json:"name"`
	FurniturePutList json.RawMessage `json:"furniture_put_list"`
	IconImageMd5     string          `json:"icon_image_md5"`
	ImageMd5         string          `json:"image_md5"`
	LikeCount        uint32          `json:"like_count"`
	FavCount         uint32          `json:"fav_count"`
}

func CreateBackyardPublishedThemeVersionTx(ctx context.Context, tx pgx.Tx, commanderID uint32, pos uint32, name string, furniturePutList json.RawMessage, iconMd5, imageMd5 string) (BackyardPublishedThemeVersion, error) {
	uploadTime := uint32(time.Now().Unix())
	entry := BackyardPublishedThemeVersion{
		ThemeID:          BackyardThemeID(commanderID, pos),
		UploadTime:       uploadTime,
		OwnerID:          commanderID,
		Pos:              pos,
		Name:             name,
		FurniturePutList: furniturePutList,
		IconImageMd5:     iconMd5,
		ImageMd5:         imageMd5,
		LikeCount:        0,
		FavCount:         0,
	}
	_, err := tx.Exec(ctx, `
INSERT INTO backyard_published_theme_versions (
  theme_id,
  upload_time,
  owner_id,
  pos,
  name,
  furniture_put_list,
  icon_image_md5,
  image_md5,
  like_count,
  fav_count
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, 0, 0
)
`, entry.ThemeID, int64(entry.UploadTime), int64(entry.OwnerID), int64(entry.Pos), entry.Name, entry.FurniturePutList, entry.IconImageMd5, entry.ImageMd5)
	if err != nil {
		return BackyardPublishedThemeVersion{}, err
	}
	return entry, nil
}

func DeleteBackyardPublishedThemeVersionsByThemeIDTx(ctx context.Context, tx pgx.Tx, themeID string) error {
	_, err := tx.Exec(ctx, `DELETE FROM backyard_published_theme_versions WHERE theme_id = $1`, themeID)
	return err
}

func LatestBackyardPublishedThemeVersion(themeID string) (*BackyardPublishedThemeVersion, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count
FROM backyard_published_theme_versions
WHERE theme_id = $1
ORDER BY upload_time DESC
LIMIT 1
`, themeID)
	var entry BackyardPublishedThemeVersion
	var uploadTime int64
	if err := row.Scan(&entry.ThemeID, &uploadTime, &entry.OwnerID, &entry.Pos, &entry.Name, &entry.FurniturePutList, &entry.IconImageMd5, &entry.ImageMd5, &entry.LikeCount, &entry.FavCount); err != nil {
		return nil, err
	}
	entry.UploadTime = uint32(uploadTime)
	return &entry, nil
}

func ListBackyardPublishedThemeIDsByPage(page uint32, num uint32) ([]string, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	offset := int64(0)
	if page > 1 {
		offset = int64(page-1) * int64(num)
	}
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT theme_id
FROM backyard_published_theme_versions
GROUP BY theme_id
ORDER BY MAX(upload_time) DESC
LIMIT $1 OFFSET $2
`, int64(num), offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func ListLatestBackyardPublishedThemeVersions() ([]BackyardPublishedThemeVersion, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT theme_id, upload_time, owner_id, pos, name, furniture_put_list, icon_image_md5, image_md5, like_count, fav_count
FROM (
  SELECT DISTINCT ON (theme_id)
    theme_id,
    upload_time,
    owner_id,
    pos,
    name,
    furniture_put_list,
    icon_image_md5,
    image_md5,
    like_count,
    fav_count
  FROM backyard_published_theme_versions
  ORDER BY theme_id, upload_time DESC
) latest
ORDER BY upload_time DESC, theme_id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]BackyardPublishedThemeVersion, 0)
	for rows.Next() {
		var entry BackyardPublishedThemeVersion
		var uploadTime int64
		if err := rows.Scan(
			&entry.ThemeID,
			&uploadTime,
			&entry.OwnerID,
			&entry.Pos,
			&entry.Name,
			&entry.FurniturePutList,
			&entry.IconImageMd5,
			&entry.ImageMd5,
			&entry.LikeCount,
			&entry.FavCount,
		); err != nil {
			return nil, err
		}
		entry.UploadTime = uint32(uploadTime)
		result = append(result, entry)
	}
	return result, rows.Err()
}

type BackyardThemeLike struct {
	CommanderID uint32 `json:"commander_id"`
	ThemeID     string `json:"theme_id"`
	UploadTime  uint32 `json:"upload_time"`
}

type BackyardThemeCollection struct {
	CommanderID uint32 `json:"commander_id"`
	ThemeID     string `json:"theme_id"`
	UploadTime  uint32 `json:"upload_time"`
}

type BackyardThemeInform struct {
	ID         uint64 `json:"id"`
	ReporterID uint32 `json:"reporter_id"`
	TargetID   uint32 `json:"target_id"`
	TargetName string `json:"target_name"`
	ThemeID    string `json:"theme_id"`
	ThemeName  string `json:"theme_name"`
	Reason     uint32 `json:"reason"`
	CreatedAt  uint32 `json:"created_at"`
}
