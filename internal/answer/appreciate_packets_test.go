package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func decodeFirstPacket(t *testing.T, client *connection.Client, expectedID int, message proto.Message) {
	t.Helper()
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
}

func ensureCommanderHasSecretary(client *connection.Client) {
	client.Commander.Secretaries = []*orm.OwnedShip{{
		ID:                 1,
		OwnerID:            client.Commander.CommanderID,
		ShipID:             1001,
		SkinID:             0,
		IsSecretary:        true,
		SecretaryPhantomID: 0,
	}}
}

func TestUnlockAppreciateMusic17503Success(t *testing.T) {
	client := setupHandlerCommander(t)
	client.Buffer.Reset()

	payload := protobuf.CS_17503{Id: proto.Uint32(10)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := UnlockAppreciateMusic(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17504
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestUpdateAppreciationMusicPlayerSettings17513Success(t *testing.T) {
	client := setupHandlerCommander(t)
	client.Buffer.Reset()

	payload := protobuf.CS_17513{MusicNo: proto.Uint32(0), MusicMode: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := UpdateAppreciationMusicPlayerSettings(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17514
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestUpdateAppreciationMusicPlayerSettings17513PersistsAndSurfacesInPlayerInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)
	client.Buffer.Reset()

	payload := protobuf.CS_17513{MusicNo: proto.Uint32(999), MusicMode: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := UpdateAppreciationMusicPlayerSettings(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	var response protobuf.SC_17514
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	client.Buffer.Reset()
	buf := []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	var info protobuf.SC_11003
	decodeFirstPacket(t, client, 11003, &info)
	if info.GetAppreciation().GetMusicNo() != 999 {
		t.Fatalf("expected music no 999, got %d", info.GetAppreciation().GetMusicNo())
	}
	if info.GetAppreciation().GetMusicMode() != 2 {
		t.Fatalf("expected music mode 2, got %d", info.GetAppreciation().GetMusicMode())
	}
}

func TestUpdateAppreciationMusicPlayerSettings17513MalformedPayloadErrors(t *testing.T) {
	client := setupHandlerCommander(t)
	client.Buffer.Reset()

	malformed := []byte{0x80}
	if _, _, err := UpdateAppreciationMusicPlayerSettings(&malformed, client); err == nil {
		t.Fatalf("expected error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response on decode error")
	}
}

func TestUnlockAppreciateGallery17501PersistsAndSurfacesInPlayerInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)

	for _, id := range []uint32{1, 32, 33} {
		client.Buffer.Reset()
		payload := protobuf.CS_17501{Id: proto.Uint32(id)}
		buffer, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		if _, _, err := UnlockAppreciateGallery(&buffer, client); err != nil {
			t.Fatalf("handler failed: %v", err)
		}
		var resp protobuf.SC_17502
		decodeResponse(t, client, &resp)
		if resp.GetResult() != 0 {
			t.Fatalf("expected result 0")
		}
	}

	state, err := orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load appreciation state: %v", err)
	}
	unlockMarks := orm.ToUint32List(state.GalleryUnlocks)
	if len(unlockMarks) < 2 {
		t.Fatalf("expected at least 2 unlock buckets, got %d", len(unlockMarks))
	}
	if unlockMarks[0] != (uint32(1) | (uint32(1) << 31)) {
		t.Fatalf("unexpected bucket 0 value %d", unlockMarks[0])
	}
	if unlockMarks[1] != 1 {
		t.Fatalf("unexpected bucket 1 value %d", unlockMarks[1])
	}

	client.Buffer.Reset()
	buf := []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	var info protobuf.SC_11003
	decodeFirstPacket(t, client, 11003, &info)
	if info.GetAppreciation() == nil {
		t.Fatalf("expected appreciation info")
	}
	if len(info.GetAppreciation().GetGallerys()) < 2 {
		t.Fatalf("expected player info gallery unlock marks")
	}
	if info.GetAppreciation().GetGallerys()[0] != unlockMarks[0] {
		t.Fatalf("expected player info bucket 0 %d, got %d", unlockMarks[0], info.GetAppreciation().GetGallerys()[0])
	}
	if info.GetAppreciation().GetGallerys()[1] != unlockMarks[1] {
		t.Fatalf("expected player info bucket 1 %d, got %d", unlockMarks[1], info.GetAppreciation().GetGallerys()[1])
	}
}

func TestMarkMangaRead17509PersistsAndSurfacesInPlayerInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)

	for _, id := range []uint32{1, 32, 33} {
		client.Buffer.Reset()
		payload := protobuf.CS_17509{Id: proto.Uint32(id)}
		buffer, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		if _, _, err := MarkMangaRead(&buffer, client); err != nil {
			t.Fatalf("handler failed: %v", err)
		}
		var resp protobuf.SC_17510
		decodeResponse(t, client, &resp)
		if resp.GetResult() != 0 {
			t.Fatalf("expected result 0")
		}
	}

	state, err := orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load appreciation state: %v", err)
	}
	readMarks := orm.ToUint32List(state.CartoonReadMark)
	if len(readMarks) < 2 {
		t.Fatalf("expected at least 2 mark buckets, got %d", len(readMarks))
	}
	if readMarks[0] != (uint32(1) | (uint32(1) << 31)) {
		t.Fatalf("unexpected bucket 0 value %d", readMarks[0])
	}
	if readMarks[1] != 1 {
		t.Fatalf("unexpected bucket 1 value %d", readMarks[1])
	}

	client.Buffer.Reset()
	buf := []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	var info protobuf.SC_11003
	decodeFirstPacket(t, client, 11003, &info)
	if len(info.GetCartoonReadMark()) < 2 {
		t.Fatalf("expected player info read marks")
	}
	if info.GetCartoonReadMark()[0] != readMarks[0] {
		t.Fatalf("expected player info bucket 0 %d, got %d", readMarks[0], info.GetCartoonReadMark()[0])
	}
	if info.GetCartoonReadMark()[1] != readMarks[1] {
		t.Fatalf("expected player info bucket 1 %d, got %d", readMarks[1], info.GetCartoonReadMark()[1])
	}
}

func TestAppreciationMarkDoesNotGrowUnbounded(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)

	client.Buffer.Reset()
	huge := uint32(^uint32(0))
	readPayload := protobuf.CS_17509{Id: proto.Uint32(huge)}
	buffer, err := proto.Marshal(&readPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := MarkMangaRead(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	state, err := orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load appreciation state: %v", err)
	}
	if len(orm.ToUint32List(state.CartoonReadMark)) != 0 {
		t.Fatalf("expected read mark to remain empty")
	}

	client.Buffer.Reset()
	likePayload := protobuf.CS_17511{Id: proto.Uint32(huge), Action: proto.Uint32(0)}
	buffer, err = proto.Marshal(&likePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleMangaLike(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	state, err = orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load appreciation state: %v", err)
	}
	if len(orm.ToUint32List(state.CartoonCollectMark)) != 0 {
		t.Fatalf("expected collect mark to remain empty")
	}

	client.Buffer.Reset()
	unlockPayload := protobuf.CS_17501{Id: proto.Uint32(huge)}
	buffer, err = proto.Marshal(&unlockPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := UnlockAppreciateGallery(&buffer, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	state, err = orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load appreciation state: %v", err)
	}
	if len(orm.ToUint32List(state.GalleryUnlocks)) != 0 {
		t.Fatalf("expected gallery unlocks to remain empty")
	}
}

func TestToggleMangaLike17511SetsAndClearsCollectMark(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)

	client.Buffer.Reset()
	likePayload := protobuf.CS_17511{Id: proto.Uint32(33), Action: proto.Uint32(0)}
	buf, err := proto.Marshal(&likePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleMangaLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var likeResp protobuf.SC_17512
	decodeResponse(t, client, &likeResp)
	if likeResp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	state, err := orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load appreciation state: %v", err)
	}
	collectMarks := orm.ToUint32List(state.CartoonCollectMark)
	if len(collectMarks) < 2 || (collectMarks[1]&1) == 0 {
		t.Fatalf("expected collect mark bit set")
	}

	client.Buffer.Reset()
	unlikePayload := protobuf.CS_17511{Id: proto.Uint32(33), Action: proto.Uint32(1)}
	buf, err = proto.Marshal(&unlikePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleMangaLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var unlikeResp protobuf.SC_17512
	decodeResponse(t, client, &unlikeResp)
	if unlikeResp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	state, err = orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load appreciation state: %v", err)
	}
	collectMarks = orm.ToUint32List(state.CartoonCollectMark)
	if len(collectMarks) < 2 {
		t.Fatalf("expected collect marks buckets")
	}
	if (collectMarks[1] & 1) != 0 {
		t.Fatalf("expected collect mark bit cleared")
	}
}

func TestToggleAppreciationGalleryLike17505PersistsAndSurfacesInPlayerInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)
	seedConfigEntry(t, "ShareCfg/gallery_config.json", "123", `{"id":123}`)

	client.Buffer.Reset()
	likePayload := protobuf.CS_17505{Id: proto.Uint32(123), Action: proto.Uint32(0)}
	buf, err := proto.Marshal(&likePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationGalleryLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var likeResp protobuf.SC_17506
	decodeResponse(t, client, &likeResp)
	if likeResp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	client.Buffer.Reset()
	buf = []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	var info protobuf.SC_11003
	decodeFirstPacket(t, client, 11003, &info)
	if !containsUint32(info.GetAppreciation().GetFavorGallerys(), 123) {
		t.Fatalf("expected favor_gallerys to contain 123")
	}

	client.Buffer.Reset()
	buf, err = proto.Marshal(&likePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationGalleryLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	client.Buffer.Reset()
	buf = []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	info = protobuf.SC_11003{}
	decodeFirstPacket(t, client, 11003, &info)
	if len(info.GetAppreciation().GetFavorGallerys()) != 1 {
		t.Fatalf("expected favor_gallerys to dedupe")
	}

	client.Buffer.Reset()
	unlikePayload := protobuf.CS_17505{Id: proto.Uint32(123), Action: proto.Uint32(1)}
	buf, err = proto.Marshal(&unlikePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationGalleryLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var unlikeResp protobuf.SC_17506
	decodeResponse(t, client, &unlikeResp)
	if unlikeResp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	client.Buffer.Reset()
	buf = []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	info = protobuf.SC_11003{}
	decodeFirstPacket(t, client, 11003, &info)
	if containsUint32(info.GetAppreciation().GetFavorGallerys(), 123) {
		t.Fatalf("expected favor_gallerys to not contain 123")
	}
}

func TestToggleAppreciationGalleryLike17505Guardrails(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)

	client.Buffer.Reset()
	unknownPayload := protobuf.CS_17505{Id: proto.Uint32(999999), Action: proto.Uint32(0)}
	buf, err := proto.Marshal(&unknownPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationGalleryLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var resp protobuf.SC_17506
	decodeResponse(t, client, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	client.Buffer.Reset()
	buf = []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	var info protobuf.SC_11003
	decodeFirstPacket(t, client, 11003, &info)
	if len(info.GetAppreciation().GetFavorGallerys()) != 0 {
		t.Fatalf("expected favor_gallerys to remain empty")
	}

	client.Buffer.Reset()
	noOpPayload := protobuf.CS_17505{Id: proto.Uint32(0), Action: proto.Uint32(0)}
	buf, err = proto.Marshal(&noOpPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationGalleryLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	decodeResponse(t, client, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	client.Buffer.Reset()
	unsupportedAction := protobuf.CS_17505{Id: proto.Uint32(999999), Action: proto.Uint32(2)}
	buf, err = proto.Marshal(&unsupportedAction)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationGalleryLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	decodeResponse(t, client, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestToggleAppreciationGalleryLike17505MalformedPayloadErrors(t *testing.T) {
	client := setupHandlerCommander(t)
	client.Buffer.Reset()

	malformed := []byte{0x80}
	if _, _, err := ToggleAppreciationGalleryLike(&malformed, client); err == nil {
		t.Fatalf("expected error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response on decode error")
	}
}

func TestToggleAppreciationMusicLike17507PersistsAndSurfacesInPlayerInfo(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)
	seedConfigEntry(t, "ShareCfg/music_collect_config.json", "10", `{"id":10}`)

	client.Buffer.Reset()
	likePayload := protobuf.CS_17507{Id: proto.Uint32(10), Action: proto.Uint32(0)}
	buf, err := proto.Marshal(&likePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationMusicLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var likeResp protobuf.SC_17508
	decodeResponse(t, client, &likeResp)
	if likeResp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	client.Buffer.Reset()
	buf = []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	var info protobuf.SC_11003
	decodeFirstPacket(t, client, 11003, &info)
	if !containsUint32(info.GetAppreciation().GetFavorMusics(), 10) {
		t.Fatalf("expected favor_musics to contain 10")
	}

	client.Buffer.Reset()
	unlikePayload := protobuf.CS_17507{Id: proto.Uint32(10), Action: proto.Uint32(1)}
	buf, err = proto.Marshal(&unlikePayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationMusicLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var unlikeResp protobuf.SC_17508
	decodeResponse(t, client, &unlikeResp)
	if unlikeResp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	client.Buffer.Reset()
	buf = []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	info = protobuf.SC_11003{}
	decodeFirstPacket(t, client, 11003, &info)
	if containsUint32(info.GetAppreciation().GetFavorMusics(), 10) {
		t.Fatalf("expected favor_musics to not contain 10")
	}
}

func TestToggleAppreciationMusicLike17507Guardrails(t *testing.T) {
	client := setupHandlerCommander(t)
	ensureCommanderHasSecretary(client)

	client.Buffer.Reset()
	unknownPayload := protobuf.CS_17507{Id: proto.Uint32(999999), Action: proto.Uint32(0)}
	buf, err := proto.Marshal(&unknownPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationMusicLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	var resp protobuf.SC_17508
	decodeResponse(t, client, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	client.Buffer.Reset()
	buf = []byte{}
	if _, _, err := PlayerInfo(&buf, client); err != nil {
		t.Fatalf("player info failed: %v", err)
	}
	var info protobuf.SC_11003
	decodeFirstPacket(t, client, 11003, &info)
	if len(info.GetAppreciation().GetFavorMusics()) != 0 {
		t.Fatalf("expected favor_musics to remain empty")
	}

	client.Buffer.Reset()
	unsupportedAction := protobuf.CS_17507{Id: proto.Uint32(999999), Action: proto.Uint32(2)}
	buf, err = proto.Marshal(&unsupportedAction)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ToggleAppreciationMusicLike(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	decodeResponse(t, client, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestToggleAppreciationMusicLike17507MalformedPayloadErrors(t *testing.T) {
	client := setupHandlerCommander(t)
	client.Buffer.Reset()

	malformed := []byte{0x80}
	if _, _, err := ToggleAppreciationMusicLike(&malformed, client); err == nil {
		t.Fatalf("expected error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response on decode error")
	}
}
