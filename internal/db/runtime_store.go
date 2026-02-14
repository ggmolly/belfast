package db

// DefaultStore is the process-wide sqlc/Postgres store.
//
// Belfast historically relied on a global Gorm DB handle. During the sqlc
// cutover we keep a single global store to avoid threading a handle through
// every packet handler.
var DefaultStore *Store
