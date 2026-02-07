package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChangeMedalDisplayPersistsAndAcks(t *testing.T) {
	client := setupHandlerCommander(t)

	cases := [][]uint32{
		nil,
		{},
		{10},
		{10, 20, 30, 40, 50},
	}
	for _, medals := range cases {
		client.Buffer.Reset()
		payload := protobuf.CS_17401{FixedConst: proto.Uint32(1), MedalId: medals}
		buffer, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		if _, _, err := ChangeMedalDisplay(&buffer, client); err != nil {
			t.Fatalf("handler failed: %v", err)
		}
		var response protobuf.SC_17402
		decodeResponse(t, client, &response)
		if response.GetResult() != 0 {
			t.Fatalf("expected result 0, got %d", response.GetResult())
		}
		stored, err := orm.ListCommanderMedalDisplay(client.Commander.CommanderID)
		if err != nil {
			t.Fatalf("list stored medals: %v", err)
		}
		if len(stored) != len(medals) {
			t.Fatalf("expected %d medals stored, got %d", len(medals), len(stored))
		}
		for i := range stored {
			if stored[i] != medals[i] {
				t.Fatalf("expected medal %d at %d, got %d", medals[i], i, stored[i])
			}
		}
	}
}

func TestChangeMedalDisplayRejectsTooMany(t *testing.T) {
	client := setupHandlerCommander(t)

	seed := protobuf.CS_17401{FixedConst: proto.Uint32(1), MedalId: []uint32{1, 2, 3}}
	seedBuf, err := proto.Marshal(&seed)
	if err != nil {
		t.Fatalf("marshal seed payload: %v", err)
	}
	if _, _, err := ChangeMedalDisplay(&seedBuf, client); err != nil {
		t.Fatalf("seed handler failed: %v", err)
	}
	client.Buffer.Reset()

	payload := protobuf.CS_17401{FixedConst: proto.Uint32(1), MedalId: []uint32{1, 2, 3, 4, 5, 6}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChangeMedalDisplay(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_17402
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
	stored, err := orm.ListCommanderMedalDisplay(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("list stored medals: %v", err)
	}
	if len(stored) != 3 {
		t.Fatalf("expected stored medals unchanged, got %v", stored)
	}
}

func TestChangeMedalDisplayRejectsDuplicates(t *testing.T) {
	client := setupHandlerCommander(t)

	payload := protobuf.CS_17401{FixedConst: proto.Uint32(1), MedalId: []uint32{10, 10}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChangeMedalDisplay(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var response protobuf.SC_17402
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
	stored, err := orm.ListCommanderMedalDisplay(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("list stored medals: %v", err)
	}
	if len(stored) != 0 {
		t.Fatalf("expected no medals stored, got %v", stored)
	}
}

func TestChangeMedalDisplaySurfacesInPlayerInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)

	medals := []uint32{99, 42, 7}
	payload := protobuf.CS_17401{FixedConst: proto.Uint32(1), MedalId: medals}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChangeMedalDisplay(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	client.Buffer.Reset()
	buf := []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	var info protobuf.SC_11003
	decodeFirstPacket(t, client, 11003, &info)
	if len(info.GetMedalId()) != len(medals) {
		t.Fatalf("expected %d medals in player info, got %d", len(medals), len(info.GetMedalId()))
	}
	for i := range medals {
		if info.GetMedalId()[i] != medals[i] {
			t.Fatalf("expected medal %d at %d, got %d", medals[i], i, info.GetMedalId()[i])
		}
	}
}
