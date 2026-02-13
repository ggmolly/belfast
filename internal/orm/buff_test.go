package orm

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var buffTestOnce sync.Once

func initBuffTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	buffTestOnce.Do(func() {
		InitDatabase()
	})
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM buffs`); err != nil {
		t.Fatalf("clear buffs: %v", err)
	}
}

func TestBuffCreate(t *testing.T) {
	initBuffTest(t)

	buff := Buff{
		ID:          1001,
		Name:        "Experience Boost",
		Description: "Increases experience gain by 50%",
		MaxTime:     3600,
		BenefitType: "exp_rate",
	}

	if err := CreateBuffRecord(&buff); err != nil {
		t.Fatalf("create buff: %v", err)
	}

	if buff.ID != 1001 {
		t.Fatalf("expected id 1001, got %d", buff.ID)
	}
	if buff.Name != "Experience Boost" {
		t.Fatalf("expected name 'Experience Boost', got %s", buff.Name)
	}
	if buff.BenefitType != "exp_rate" {
		t.Fatalf("expected benefit type 'exp_rate', got %s", buff.BenefitType)
	}
}

func TestBuffFind(t *testing.T) {
	initBuffTest(t)

	buff := Buff{
		ID:          1002,
		Name:        "Gold Boost",
		Description: "Increases gold gain by 100%",
		MaxTime:     7200,
		BenefitType: "gold_rate",
	}
	if err := CreateBuffRecord(&buff); err != nil {
		t.Fatalf("create buff: %v", err)
	}

	found, err := GetBuffByID(buff.ID)
	if err != nil {
		t.Fatalf("find buff: %v", err)
	}

	if found.ID != 1002 {
		t.Fatalf("expected id 1002, got %d", found.ID)
	}
	if found.Name != "Gold Boost" {
		t.Fatalf("expected name 'Gold Boost', got %s", found.Name)
	}
}

func TestBuffUpdate(t *testing.T) {
	initBuffTest(t)

	buff := Buff{
		ID:          1003,
		Name:        "Drop Rate Boost",
		Description: "Increases drop rate",
		MaxTime:     1800,
		BenefitType: "drop_rate",
	}
	if err := CreateBuffRecord(&buff); err != nil {
		t.Fatalf("create buff: %v", err)
	}

	buff.Name = "Enhanced Drop Rate"
	buff.Description = "Greatly increases drop rate"
	if err := UpdateBuffRecord(&buff); err != nil {
		t.Fatalf("update buff: %v", err)
	}

	found, err := GetBuffByID(buff.ID)
	if err != nil {
		t.Fatalf("find updated buff: %v", err)
	}

	if found.Name != "Enhanced Drop Rate" {
		t.Fatalf("expected name 'Enhanced Drop Rate', got %s", found.Name)
	}
	if found.Description != "Greatly increases drop rate" {
		t.Fatalf("expected description 'Greatly increases drop rate', got %s", found.Description)
	}
}

func TestBuffDelete(t *testing.T) {
	initBuffTest(t)

	buff := Buff{
		ID:          1004,
		Name:        "Test Buff",
		Description: "Test description",
		MaxTime:     60,
		BenefitType: "test_type",
	}
	if err := CreateBuffRecord(&buff); err != nil {
		t.Fatalf("create buff: %v", err)
	}

	if err := DeleteBuffRecord(buff.ID); err != nil {
		t.Fatalf("delete buff: %v", err)
	}

	_, err := GetBuffByID(buff.ID)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestBuffMaxTime(t *testing.T) {
	initBuffTest(t)

	tests := []struct {
		name    string
		buff    Buff
		wantMax int
	}{
		{
			name: "default max time zero",
			buff: Buff{
				ID:          2001,
				Name:        "Zero Max Time",
				Description: "Test",
				BenefitType: "test",
			},
			wantMax: 0,
		},
		{
			name: "positive max time",
			buff: Buff{
				ID:          2002,
				Name:        "One Hour",
				Description: "Test",
				MaxTime:     3600,
				BenefitType: "test",
			},
			wantMax: 3600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateBuffRecord(&tt.buff); err != nil {
				t.Fatalf("create buff: %v", err)
			}

			found, err := GetBuffByID(tt.buff.ID)
			if err != nil {
				t.Fatalf("find buff: %v", err)
			}

			if found.MaxTime != tt.wantMax {
				t.Fatalf("expected max time %d, got %d", tt.wantMax, found.MaxTime)
			}
		})
	}
}

func TestBuffNameAndDescriptionLimits(t *testing.T) {
	initBuffTest(t)

	buff := Buff{
		ID:          3001,
		Name:        strings.Repeat("X", 50),
		Description: strings.Repeat("Y", 170),
		MaxTime:     100,
		BenefitType: "test",
	}

	if err := CreateBuffRecord(&buff); err != nil {
		t.Fatalf("create buff: %v", err)
	}

	found, err := GetBuffByID(buff.ID)
	if err != nil {
		t.Fatalf("find buff: %v", err)
	}

	if len(found.Name) > 50 {
		t.Fatalf("expected name length <= 50, got %d", len(found.Name))
	}
	if len(found.Description) > 170 {
		t.Fatalf("expected description length <= 170, got %d", len(found.Description))
	}
}

func TestBuffBenefitTypes(t *testing.T) {
	initBuffTest(t)

	benefitTypes := []string{"exp_rate", "gold_rate", "drop_rate", "oil_rate", "build_speed"}

	for i, benefitType := range benefitTypes {
		buff := Buff{
			ID:          uint32(4001 + i),
			Name:        "Buff " + benefitType,
			Description: "Description for " + benefitType,
			MaxTime:     3600,
			BenefitType: benefitType,
		}
		if err := CreateBuffRecord(&buff); err != nil {
			t.Fatalf("create buff %s: %v", benefitType, err)
		}
	}

	found, _, err := ListBuffsPage(0, 100)
	if err != nil {
		t.Fatalf("find buffs: %v", err)
	}

	if len(found) != len(benefitTypes) {
		t.Fatalf("expected %d buffs, got %d", len(benefitTypes), len(found))
	}
}
