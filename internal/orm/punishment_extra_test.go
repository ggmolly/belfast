package orm

import (
	"testing"
	"time"
)

func TestPunishmentCRUDAndRetrieve(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Punishment{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 90, AccountID: 90, Name: "Pun"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	lift := time.Now().Add(time.Hour)
	punish := Punishment{PunishedID: commander.CommanderID, LiftTimestamp: &lift}
	if err := punish.Create(); err != nil {
		t.Fatalf("create punishment: %v", err)
	}
	punish.IsPermanent = true
	if err := punish.Update(); err != nil {
		t.Fatalf("update punishment: %v", err)
	}
	loaded := Punishment{ID: punish.ID}
	if err := loaded.Retrieve(false); err != nil {
		t.Fatalf("retrieve punishment: %v", err)
	}
	if err := loaded.Retrieve(true); err != nil {
		t.Fatalf("retrieve punishment greedy: %v", err)
	}
	if err := loaded.Delete(); err != nil {
		t.Fatalf("delete punishment: %v", err)
	}
}
