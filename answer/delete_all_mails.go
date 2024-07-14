package answer

// import (
// 	"github.com/ggmolly/belfast/connection"
// 	"github.com/ggmolly/belfast/orm"

// 	"github.com/ggmolly/belfast/protobuf"
// )

// func DeleteAllMails(buffer *[]byte, client *connection.Client) (int, int, error) {
// 	mailIds := make([]uint32, len(client.Commander.Mails))
// 	for i, mail := range client.Commander.Mails {
// 		mailIds[i] = mail.ID
// 	}
// 	response := protobuf.SC_30007{
// 		IdList: mailIds,
// 	}
// 	if err := client.Commander.CleanMailbox(); err != nil {
// 		return 0, 30007, err
// 	}
// 	client.Commander.Mails = make([]orm.Mail, 0)
// 	client.Commander.MailsMap = make(map[uint32]*orm.Mail)
// 	return client.SendMessage(30007, &response)
// }
