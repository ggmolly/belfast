package orm

// QualifiedTable returns the DB table name using the active NamingStrategy.
// When using Postgres with database.schema_name, tables are created under that
// schema via gorm's TablePrefix (e.g. "belfast.account_roles").
func QualifiedTable(name string) string {
	if GormDB == nil || GormDB.Config == nil || GormDB.Config.NamingStrategy == nil {
		return name
	}
	return GormDB.Config.NamingStrategy.TableName(name)
}
