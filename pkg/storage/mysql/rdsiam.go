package mysql

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"

	"github.com/openfga/openfga/pkg/storage/sqlcommon"
)

func openRDS(uri *mysql.Config, cfg *sqlcommon.Config) (*sql.DB, error) {
	panic("not implemented")
}
