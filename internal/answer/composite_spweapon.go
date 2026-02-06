package answer

import (
	"sync/atomic"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

// v1 special-weapon UID allocator.
//
// We do not yet persist crafted spweapons server-side, but the client requires a
// non-zero uid for SpWeapon.CreateByNet to produce a usable instance. A
// process-wide monotonic counter avoids collisions within a session and is easy
// to swap for a persistent allocator later.
var spweaponUIDCounter uint32

func nextSpweaponUID() uint32 {
	uid := atomic.AddUint32(&spweaponUIDCounter, 1)
	if uid == 0 {
		uid = atomic.AddUint32(&spweaponUIDCounter, 1)
	}
	return uid
}

func CompositeSpWeapon(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14209
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14209, err
	}

	templateId := data.GetTemplateId()
	response := protobuf.SC_14210{}
	if templateId == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14210, &response)
	}

	response.Result = proto.Uint32(0)
	response.Spweapon = &protobuf.SPWEAPONINFO{
		Id:         proto.Uint32(nextSpweaponUID()),
		TemplateId: proto.Uint32(templateId),
		Attr_1:     proto.Uint32(0),
		Attr_2:     proto.Uint32(0),
		AttrTemp_1: proto.Uint32(0),
		AttrTemp_2: proto.Uint32(0),
		Effect:     proto.Uint32(0),
		Pt:         proto.Uint32(0),
	}
	return client.SendMessage(14210, &response)
}
