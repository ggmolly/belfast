package authz

const (
	RoleAdmin  = "admin"
	RolePlayer = "player"
)

const (
	PermAdminAuthz      = "admin.authz"
	PermAdminUsers      = "admin.users"
	PermAdminPermission = "admin.permission_policy"
	PermPlayers         = "players"
	PermGameData        = "game_data"
	PermShop            = "shop"
	PermNotices         = "notices"
	PermExchangeCodes   = "exchange_codes"
	PermDorm3D          = "dorm3d"
	PermActivities      = "activities"
	PermJuustagram      = "juustagram"
	PermServer          = "server"
	PermMeResources     = "me.resources"
	PermMeShips         = "me.ships"
	PermMeItems         = "me.items"
	PermMeSkins         = "me.skins"
)

// KnownPermissions returns the permission keys that are enforced by the API.
// Values are human-readable descriptions.
func KnownPermissions() map[string]string {
	return map[string]string{
		PermAdminAuthz:      "Manage roles and permissions",
		PermAdminUsers:      "Manage staff accounts",
		PermAdminPermission: "Manage permission policy",
		PermPlayers:         "Manage player accounts",
		PermGameData:        "Manage game data",
		PermShop:            "Manage shop offers",
		PermNotices:         "Manage notices",
		PermExchangeCodes:   "Manage exchange codes",
		PermDorm3D:          "Manage Dorm3D",
		PermActivities:      "Manage activities",
		PermJuustagram:      "Manage Juustagram",
		PermServer:          "Manage server",
		PermMeResources:     "Self resources read/update",
		PermMeShips:         "Give ships to self",
		PermMeItems:         "Give items to self",
		PermMeSkins:         "Give skins to self",
	}
}
