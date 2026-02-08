package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestSendMailList_IndexBeginZeroDoesNotUnderflow(t *testing.T) {
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.MailAttachment{})
	clearTable(t, &orm.Mail{})
	clearTable(t, &orm.Commander{})

	client := newTestClient(t)
	insertMail(t, client.Commander, "mail-1", nil, false)
	insertMail(t, client.Commander, "mail-2", nil, false)

	payload := &protobuf.CS_30002{
		Type:       proto.Uint32(0),
		IndexBegin: proto.Uint32(0),
		IndexEnd:   proto.Uint32(0),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.SendMailList(&buf, client); err != nil {
		t.Fatalf("SendMailList failed: %v", err)
	}

	response := &protobuf.SC_30003{}
	decodePacket(t, client, 30003, response)
	if len(response.MailList) != 2 {
		t.Fatalf("expected 2 mails, got %d", len(response.MailList))
	}
}
