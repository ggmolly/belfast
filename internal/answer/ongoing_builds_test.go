package answer

import (
	"os"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestOngoingBuildsSnapshot(t *testing.T) {
	os.Setenv("MODE", "test")
	orm.InitDatabase()

	now := time.Now()
	commander := &orm.Commander{
		DrawCount1:    2,
		DrawCount10:   3,
		ExchangeCount: 5,
		Builds: []orm.Build{
			{ID: 2, PoolID: 3, FinishesAt: now.Add(2 * time.Minute)},
			{ID: 1, PoolID: 1, FinishesAt: now.Add(-1 * time.Minute)},
		},
	}
	client := &connection.Client{Commander: commander}
	buffer := []byte{}

	_, packetID, err := OngoingBuilds(&buffer, client)
	if err != nil {
		t.Fatalf("ongoing builds failed: %v", err)
	}
	if packetID != 12024 {
		t.Fatalf("expected packet 12024, got %d", packetID)
	}

	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected buffer to include header and payload")
	}
	data = data[7:]

	var response protobuf.SC_12024
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if response.GetWorklistCount() != consts.MaxBuildWorkCount {
		t.Fatalf("expected worklist count %d, got %d", consts.MaxBuildWorkCount, response.GetWorklistCount())
	}
	if response.GetDrawCount_1() != 2 || response.GetDrawCount_10() != 3 || response.GetExchangeCount() != 5 {
		t.Fatalf("unexpected counters")
	}

	list := response.GetWorklistList()
	if len(list) != 2 {
		t.Fatalf("expected 2 build entries, got %d", len(list))
	}
	if list[0].GetBuildId() != 1 {
		t.Fatalf("expected pool id 1 first, got %d", list[0].GetBuildId())
	}
	if list[0].GetTime() != 0 {
		t.Fatalf("expected finished build to have 0 remaining time")
	}
	if list[1].GetBuildId() != 3 {
		t.Fatalf("expected pool id 3 second, got %d", list[1].GetBuildId())
	}
	if list[1].GetTime() == 0 {
		t.Fatalf("expected active build to have remaining time")
	}
}
