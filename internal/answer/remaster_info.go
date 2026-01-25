package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func RemasterInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13505
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13506, err
	}
	progress, err := orm.ListRemasterProgress(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 13506, err
	}
	gains, err := listRemasterDropGains()
	if err != nil {
		return 0, 13506, err
	}
	progressMap := make(map[remasterDropKey]orm.RemasterProgress, len(progress))
	for _, entry := range progress {
		key := remasterDropKey{ChapterID: entry.ChapterID, Pos: entry.Pos}
		progressMap[key] = entry
	}
	list := make([]*protobuf.REMAPCOUNT, 0, len(gains))
	for _, gain := range gains {
		entry := progressMap[remasterDropKey{ChapterID: gain.ChapterID, Pos: gain.Pos}]
		flag := uint32(0)
		if entry.Received {
			flag = 1
		}
		list = append(list, &protobuf.REMAPCOUNT{
			ChapterId: proto.Uint32(gain.ChapterID),
			Pos:       proto.Uint32(gain.Pos),
			Count:     proto.Uint32(entry.Count),
			Flag:      proto.Uint32(flag),
		})
	}
	response := protobuf.SC_13506{RemapCountList: list}
	return client.SendMessage(13506, &response)
}

func getRemasterProgress(db *gorm.DB, commanderID uint32, chapterID uint32, pos uint32) (*orm.RemasterProgress, error) {
	entry, err := orm.GetRemasterProgress(db, commanderID, chapterID, pos)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &orm.RemasterProgress{CommanderID: commanderID, ChapterID: chapterID, Pos: pos}, nil
		}
		return nil, err
	}
	return entry, nil
}
