package consts

// NOTE: All of the following consts are used in the update packet and game update
// functions to re-emit a SC_10801 packet with the game's update information
// and to automatically fetch the game's updates
// Upon implementing a new region, the gateway and proxy must be added here
// check answer/update_packet.go for more information
var (
	RegionGateways = map[string]string{
		"EN": "blhxusgate.yo-star.com",
		"CN": "line1-login-bili-blhx.bilibiligame.net",
	}
	RegionProxies = map[string]string{
		"EN": "blhxusproxy.yo-star.com",
		// Its port in the apk should be 8080, but we don't need to care about it at the moment
		"CN": "line1-bak-login-bili-blhx.bilibiligame.net",
	}
	// This might not change between regions, but it's here just in case
	Monday_0OclockTimestamps = map[string]uint32{
		"EN": 1606114800,
		// No one knows why, but according to version rules, this date may indeed be earlier than EN
		"CN": 1606060800,
	}

	// This map represents the game's url on the respective platforms
	GamePlatformUrl = map[string]map[string]string{
		"EN": {
			"0": "https://play.google.com/store/apps/details?id=com.YoStarEN.AzurLane",
			"1": "https://itunes.apple.com/us/app/azur-lane/id1411126549",
		},
		"CN": {
			// On Android, it should point to the latest apk address,
			// but we can currently point directly to the official website
			"0": "https://blhx.biligame.com/",
			"1": "https://blhx.biligame.com/",
		},
	}
)
