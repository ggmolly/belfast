package answer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"google.golang.org/protobuf/proto"
)

func TestBuildLegacyThemeList19108ResponseBoundsPayloadSize(t *testing.T) {
	versions := make([]orm.BackyardPublishedThemeVersion, 0, 200)
	for i := 0; i < 200; i++ {
		versions = append(versions, orm.BackyardPublishedThemeVersion{
			ThemeID:          fmt.Sprintf("theme-%03d", i),
			UploadTime:       uint32(i + 1),
			OwnerID:          100,
			Pos:              uint32(i + 1),
			Name:             strings.Repeat("n", 512),
			FurniturePutList: []byte(`[]`),
		})
	}

	resp := buildLegacyThemeList19108Response(versions)
	if len(resp.GetThemeList()) >= len(versions) {
		t.Fatalf("expected bounded theme list to truncate input")
	}
	if size := proto.Size(&resp); size > maxLegacyThemeListPayloadBytes {
		t.Fatalf("expected response size <= %d, got %d", maxLegacyThemeListPayloadBytes, size)
	}
}

func TestBuildLegacyThemeList19108ResponseKeepsSmallPayload(t *testing.T) {
	versions := []orm.BackyardPublishedThemeVersion{
		{ThemeID: "theme-1", UploadTime: 1, OwnerID: 10, Pos: 1, Name: "one", FurniturePutList: []byte(`[]`)},
		{ThemeID: "theme-2", UploadTime: 2, OwnerID: 11, Pos: 2, Name: "two", FurniturePutList: []byte(`[]`)},
	}

	resp := buildLegacyThemeList19108Response(versions)
	if got := len(resp.GetThemeList()); got != len(versions) {
		t.Fatalf("expected %d themes, got %d", len(versions), got)
	}
	if size := proto.Size(&resp); size > maxLegacyThemeListPayloadBytes {
		t.Fatalf("expected response size <= %d, got %d", maxLegacyThemeListPayloadBytes, size)
	}
}
