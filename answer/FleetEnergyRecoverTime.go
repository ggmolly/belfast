package answer

import (
	"time"

	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func FleetEnergyRecoverTime(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_12031
	// let's assert every hour
	now := time.Now()
	// set seconds, minutes and nanoseconds to 0
	now = now.Add(-time.Duration(now.Second()) * time.Second)
	now = now.Add(-time.Duration(now.Minute()) * time.Minute)
	now = now.Add(-time.Duration(now.Nanosecond()) * time.Nanosecond)
	// Check if the user logged in in the last hour, if so, add 1 hour to the recovery time
	if now.Sub(client.Commander.LastLogin).Hours() < 1 {
		response.EnergyAutoIncreaseTime = proto.Uint32(uint32(now.Add(time.Hour).Unix()))
	} else {
		response.EnergyAutoIncreaseTime = proto.Uint32(uint32(now.Unix()))
	}
	return client.SendMessage(12031, &response)
}
