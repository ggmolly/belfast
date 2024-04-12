package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateMailImpFlag(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_30010
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 30011, err
	}
	var response protobuf.SC_30011
	mail, ok := client.Commander.MailsMap[data.GetId()]
	if !ok {
		response.Result = proto.Uint32(1) // 1 = mail not found
	} else {
		mail.IsImportant = data.GetFlag() == 1
		err = mail.Update()
		if err != nil {
			return 0, 30011, err
		}
		response.Result = proto.Uint32(0) // 0 = success
	}
	return client.SendMessage(30011, &response)
}
