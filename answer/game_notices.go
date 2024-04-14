package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/orm"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func GameNotices(buffer *[]byte, client *connection.Client) (int, int, error) {
	var notices []orm.Notice
	if err := orm.GormDB.Order("id desc").Limit(10).Find(&notices).Error; err != nil {
		return 0, 11300, fmt.Errorf("failed to get notices: %w", err)
	}
	response := protobuf.SC_11300{
		NoticeList: make([]*protobuf.NOTICEINFO, len(notices)),
	}

	for i, notice := range notices {
		response.NoticeList[i] = &protobuf.NOTICEINFO{
			Id:      proto.Uint32(uint32(notice.ID)),
			Title:   proto.String(notice.Title),
			Content: proto.String(notice.Content),
		}
	}
	return client.SendMessage(11300, &response)
}
