package orm

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func BackyardThemeID(commanderID uint32, pos uint32) string {
	// Client builds theme id via string concatenation: player.id .. pos
	return fmt.Sprintf("%d%d", commanderID, pos)
}

// BackyardCustomThemeTemplate stores the per-commander editable template (pos 1..N).
// UploadTime==0 means not published.
type BackyardCustomThemeTemplate struct {
	CommanderID      uint32          `gorm:"primaryKey"`
	Pos              uint32          `gorm:"primaryKey"`
	Name             string          `gorm:"size:50;default:'';not_null"`
	FurniturePutList json.RawMessage `gorm:"type:json;not_null"`
	IconImageMd5     string          `gorm:"size:64;default:'';not_null"`
	ImageMd5         string          `gorm:"size:64;default:'';not_null"`
	UploadTime       uint32          `gorm:"not_null;default:0"`
}

func UpsertBackyardCustomThemeTemplateTx(tx *gorm.DB, commanderID uint32, pos uint32, name string, furniturePutList json.RawMessage, iconMd5, imageMd5 string) error {
	entry := BackyardCustomThemeTemplate{
		CommanderID:      commanderID,
		Pos:              pos,
		Name:             name,
		FurniturePutList: furniturePutList,
		IconImageMd5:     iconMd5,
		ImageMd5:         imageMd5,
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "pos"}},
		DoUpdates: clause.Assignments(map[string]any{
			"name":               name,
			"furniture_put_list": furniturePutList,
			"icon_image_md5":     iconMd5,
			"image_md5":          imageMd5,
		}),
	}).Create(&entry).Error
}

func ListBackyardCustomThemeTemplates(commanderID uint32) ([]BackyardCustomThemeTemplate, error) {
	var entries []BackyardCustomThemeTemplate
	if err := GormDB.Where("commander_id = ?", commanderID).Order("pos asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func GetBackyardCustomThemeTemplate(commanderID uint32, pos uint32) (*BackyardCustomThemeTemplate, error) {
	var entry BackyardCustomThemeTemplate
	if err := GormDB.Where("commander_id = ? AND pos = ?", commanderID, pos).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func DeleteBackyardCustomThemeTemplateTx(tx *gorm.DB, commanderID uint32, pos uint32) error {
	return tx.Where("commander_id = ? AND pos = ?", commanderID, pos).Delete(&BackyardCustomThemeTemplate{}).Error
}

// BackyardPublishedThemeVersion stores an uploaded version of a theme.
// ThemeID is the stable id (player.id .. pos), UploadTime disambiguates versions.
type BackyardPublishedThemeVersion struct {
	ThemeID          string          `gorm:"primaryKey;size:64"`
	UploadTime       uint32          `gorm:"primaryKey"`
	OwnerID          uint32          `gorm:"not_null;index"`
	Pos              uint32          `gorm:"not_null"`
	Name             string          `gorm:"size:50;default:'';not_null"`
	FurniturePutList json.RawMessage `gorm:"type:json;not_null"`
	IconImageMd5     string          `gorm:"size:64;default:'';not_null"`
	ImageMd5         string          `gorm:"size:64;default:'';not_null"`
	LikeCount        uint32          `gorm:"not_null;default:0"`
	FavCount         uint32          `gorm:"not_null;default:0"`
}

func CreateBackyardPublishedThemeVersionTx(tx *gorm.DB, commanderID uint32, pos uint32, name string, furniturePutList json.RawMessage, iconMd5, imageMd5 string) (BackyardPublishedThemeVersion, error) {
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
	}
	if err := tx.Create(&entry).Error; err != nil {
		return BackyardPublishedThemeVersion{}, err
	}
	return entry, nil
}

func DeleteBackyardPublishedThemeVersionsByThemeIDTx(tx *gorm.DB, themeID string) error {
	return tx.Where("theme_id = ?", themeID).Delete(&BackyardPublishedThemeVersion{}).Error
}

func LatestBackyardPublishedThemeVersion(themeID string) (*BackyardPublishedThemeVersion, error) {
	var entry BackyardPublishedThemeVersion
	if err := GormDB.Where("theme_id = ?", themeID).Order("upload_time desc").First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func ListBackyardPublishedThemeIDsByPage(page uint32, num uint32) ([]string, error) {
	// Return distinct ThemeID ordered by latest upload_time desc.
	// This is SQLite-friendly and keeps logic in SQL.
	// theme_id is stable per (owner,pos).
	rows, err := GormDB.Raw(`
		SELECT theme_id
		FROM backyard_published_theme_versions
		GROUP BY theme_id
		ORDER BY MAX(upload_time) DESC
		LIMIT ? OFFSET ?
	`, num, (page-1)*num).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

type BackyardThemeLike struct {
	CommanderID uint32 `gorm:"primaryKey"`
	ThemeID     string `gorm:"primaryKey;size:64"`
	UploadTime  uint32 `gorm:"primaryKey"`
}

type BackyardThemeCollection struct {
	CommanderID uint32 `gorm:"primaryKey"`
	ThemeID     string `gorm:"primaryKey;size:64"`
	UploadTime  uint32 `gorm:"primaryKey"`
}

type BackyardThemeInform struct {
	ID         uint64 `gorm:"primaryKey"`
	ReporterID uint32 `gorm:"not_null;index"`
	TargetID   uint32 `gorm:"not_null"`
	TargetName string `gorm:"size:64;default:'';not_null"`
	ThemeID    string `gorm:"size:64;not_null"`
	ThemeName  string `gorm:"size:64;default:'';not_null"`
	Reason     uint32 `gorm:"not_null"`
	CreatedAt  uint32 `gorm:"not_null"`
}
