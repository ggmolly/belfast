package consts

// NOTE: All of the following consts are used in the update packet and game update
// functions to re-emit a SC_10801 packet with the game's update information
// and to automatically fetch the game's updates
// Upon implementing a new region, the gateway and proxy must be added here
// check answer/update_packet.go for more information
var (
	RegionGateways = map[string]string{
		"EN": "blhxusgate.yo-star.com",
	}
	RegionProxies = map[string]string{
		"EN": "blhxusproxy.yo-star.com",
	}
	// This might not change between regions, but it's here just in case
	Monday_0OclockTimestamps = map[string]uint32{
		"EN": 1606114800,
	}

	// This map represents the game's url on the respective platforms
	GamePlatformUrl = map[string]map[string]string{
		"EN": {
			"0": "https://play.google.com/store/apps/details?id=com.YoStarEN.AzurLane",
			"1": "https://itunes.apple.com/us/app/azur-lane/id1411126549",
		},
	}
)
