package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC20101 protobuf.SC_20101

func WeeklyMissions(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_20101
	response.Info = &protobuf.WEEKLY_INFO{
		Pt:       proto.Uint32(0),
		RewardLv: proto.Uint32(0),
	}
	return client.SendMessage(20101, &response)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC20101)
	task := validSC20101.Info.GetTask()
	for i, _ := range task {
		task[i].Monday_0Clock = proto.Uint32(1606114800)
	}
}
