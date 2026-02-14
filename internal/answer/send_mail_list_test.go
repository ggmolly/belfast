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

func TestSendMailList_RefreshesStaleMailboxCache(t *testing.T) {
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.MailAttachment{})
	clearTable(t, &orm.Mail{})
	clearTable(t, &orm.Commander{})

	client := newTestClient(t)

	otherCommander := orm.Commander{CommanderID: client.Commander.CommanderID}
	outOfBandMail := orm.Mail{
		Title: "Registration PIN",
		Body:  "Your registration PIN is B-123456.",
	}
	if err := otherCommander.SendMail(&outOfBandMail); err != nil {
		t.Fatalf("failed to send out-of-band mail: %v", err)
	}

	if len(client.Commander.Mails) != 0 {
		t.Fatalf("expected stale client mail cache, got %d mails", len(client.Commander.Mails))
	}

	payload := &protobuf.CS_30002{
		Type:       proto.Uint32(1),
		IndexBegin: proto.Uint32(1),
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
	if len(response.MailList) != 1 {
		t.Fatalf("expected 1 mail, got %d", len(response.MailList))
	}
	if response.MailList[0].GetId() != outOfBandMail.ID {
		t.Fatalf("expected mail id %d, got %d", outOfBandMail.ID, response.MailList[0].GetId())
	}
}
