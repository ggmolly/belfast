package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type tutorialHandbook struct {
	ID     uint32   `json:"id"`
	TagIDs []uint32 `json:"tag_list"`
}

type tutorialHandbookTask struct {
	ID uint32 `json:"id"`
	PT uint32 `json:"pt"`
}

func CommanderManualInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	handbookEntries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/tutorial_handbook.json")
	if err != nil {
		return 0, 22300, err
	}
	taskEntries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/tutorial_handbook_task.json")
	if err != nil {
		return 0, 22300, err
	}
	ptByTask := make(map[uint32]uint32, len(taskEntries))
	for _, entry := range taskEntries {
		var task tutorialHandbookTask
		if err := json.Unmarshal(entry.Data, &task); err != nil {
			return 0, 22300, err
		}
		ptByTask[task.ID] = task.PT
	}
	response := protobuf.SC_22300{
		Handbooks:       []*protobuf.TUTHANDBOOK{},
		FinishedTaskIds: []uint32{},
	}
	for _, entry := range handbookEntries {
		var handbook tutorialHandbook
		if err := json.Unmarshal(entry.Data, &handbook); err != nil {
			return 0, 22300, err
		}
		for _, taskID := range handbook.TagIDs {
			response.Handbooks = append(response.Handbooks, &protobuf.TUTHANDBOOK{
				Id:    proto.Uint32(taskID),
				Pt:    proto.Uint32(ptByTask[taskID]),
				Award: proto.Uint32(0),
			})
		}
	}
	return client.SendMessage(22300, &response)
}

func CommanderManualGetTask(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_22302
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 22303, err
	}
	response := protobuf.SC_22303{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(22303, &response)
}

func CommanderManualGetPtAward(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_22304
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 22305, err
	}
	response := protobuf.SC_22305{
		Result:   proto.Uint32(0),
		DropList: []*protobuf.DROPINFO{},
	}
	return client.SendMessage(22305, &response)
}
