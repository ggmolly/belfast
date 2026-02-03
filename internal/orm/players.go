package orm

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type PlayerQueryParams struct {
	Offset       int
	Limit        int
	FilterBanned bool
	FilterOnline bool
	OnlineIDs    []uint32
	MinLevel     int
	Search       string
}

type PlayerListResult struct {
	Commanders []Commander
	Total      int64
}

type PlayerBanStatus struct {
	Banned   bool
	LiftTime *time.Time
}

func ListCommanders(db *gorm.DB, params PlayerQueryParams) (PlayerListResult, error) {
	query := db.Model(&Commander{})
	query = applyPlayerFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PlayerListResult{}, err
	}

	var commanders []Commander
	query = query.Order("last_login desc")
	query = ApplyPagination(query, params.Offset, params.Limit)
	if err := query.Find(&commanders).Error; err != nil {
		return PlayerListResult{}, err
	}

	return PlayerListResult{Commanders: commanders, Total: total}, nil
}

func SearchCommanders(db *gorm.DB, params PlayerQueryParams) (PlayerListResult, error) {
	query := db.Model(&Commander{})
	query = applyPlayerFilters(query, params)

	if strings.TrimSpace(params.Search) != "" {
		search := strings.ToLower(strings.TrimSpace(params.Search))
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PlayerListResult{}, err
	}

	var commanders []Commander
	query = query.Order("last_login desc")
	query = ApplyPagination(query, params.Offset, params.Limit)
	if err := query.Find(&commanders).Error; err != nil {
		return PlayerListResult{}, err
	}

	return PlayerListResult{Commanders: commanders, Total: total}, nil
}

func applyPlayerFilters(query *gorm.DB, params PlayerQueryParams) *gorm.DB {
	if params.MinLevel > 0 {
		query = query.Where("level >= ?", params.MinLevel)
	}

	if params.FilterBanned {
		now := time.Now()
		query = query.Where("EXISTS (SELECT 1 FROM punishments WHERE punishments.punished_id = commanders.commander_id AND (punishments.lift_timestamp IS NULL OR punishments.lift_timestamp > ?))", now)
	}
	if params.FilterOnline {
		if len(params.OnlineIDs) > 0 {
			query = query.Where("commander_id IN ?", params.OnlineIDs)
		} else {
			query = query.Where("1 = 0")
		}
	}

	return query
}

func LoadCommanderWithDetails(id uint32) (Commander, error) {
	var commander Commander
	if err := GormDB.
		Preload("Ships.Ship").
		Preload("Ships.Equipments").
		Preload("Ships.Strengths").
		Preload("Items.Item").
		Preload("MiscItems.Item").
		Preload("OwnedResources.Resource").
		Preload("Builds.Ship").
		Preload("Mails.Attachments").
		Preload("Compensations.Attachments").
		Preload("OwnedSkins").
		Preload("OwnedEquipments").
		Preload("Fleets").
		First(&commander, id).Error; err != nil {
		return Commander{}, err
	}
	return commander, nil
}

func GetBanStatus(commanderID uint32) (PlayerBanStatus, error) {
	var punishment Punishment
	if err := GormDB.
		Where("punished_id = ?", commanderID).
		Order("id desc").
		First(&punishment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return PlayerBanStatus{Banned: false, LiftTime: nil}, nil
		}
		return PlayerBanStatus{}, err
	}
	if punishment.LiftTimestamp != nil {
		if time.Now().Before(*punishment.LiftTimestamp) {
			return PlayerBanStatus{Banned: true, LiftTime: punishment.LiftTimestamp}, nil
		}
		return PlayerBanStatus{Banned: false, LiftTime: nil}, nil
	}
	return PlayerBanStatus{Banned: true, LiftTime: nil}, nil
}

func ActivePunishment(commanderID uint32) (*Punishment, error) {
	var punishment Punishment
	if err := GormDB.
		Where("punished_id = ?", commanderID).
		Order("id desc").
		First(&punishment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	if punishment.LiftTimestamp != nil && time.Now().After(*punishment.LiftTimestamp) {
		return nil, gorm.ErrRecordNotFound
	}
	return &punishment, nil
}
