package mysql

import (
	"database/sql"

	"github.com/openfga/openfga/pkg/storage/sqlcommon"
)

func openRDS(uri string, cfg *sqlcommon.Config) (*sql.DB, error) {
	panic("not implemented")
}
