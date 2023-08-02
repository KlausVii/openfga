package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/jackc/pgx/v5"
	pgxstd "github.com/jackc/pgx/v5/stdlib"

	"github.com/openfga/openfga/pkg/storage/sqlcommon"
)

func openRDS(uri *url.URL, cfg *sqlcommon.Config) (*sql.DB, error) {
	ctx := context.Background()
	iam, err := newIAM(ctx, cfg.Username, uri.Hostname(), uri.Port())
	if err != nil {
		return nil, err
	}
	conf, err := pgx.ParseConfig(uri.String())
	if err != nil {
		return nil, err
	}
	connector := pgxstd.GetConnector(*conf, pgxstd.OptionBeforeConnect(iam.BeforeConnect))
	db := sql.OpenDB(connector)
	return db, nil
}

type iam struct {
	lock sync.RWMutex

	username, host, port string
	awsCfg               aws.Config
	token                string
	expiresAt            time.Time
}

func newIAM(ctx context.Context, username, host, port string) (*iam, error) {
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &iam{
		username: username,
		host:     host,
		port:     port,
		awsCfg:   awsCfg,
	}, nil
}

func (i *iam) BeforeConnect(ctx context.Context, conn *pgx.ConnConfig) error {
	token, err := i.getRDSAuthToken(ctx)
	if err != nil {
		return err
	}
	conn.Password = token
	return nil
}

func (i *iam) getRDSAuthToken(ctx context.Context) (string, error) {
	i.lock.RLock()
	if i.token != "" && i.expiresAt.After(time.Now()) {
		defer i.lock.RUnlock()
		return i.token, nil
	}
	i.lock.RUnlock()

	endpoint := fmt.Sprintf("%s:%d", i.host, i.port)
	token, err := auth.BuildAuthToken(ctx, endpoint, i.awsCfg.Region, i.username, i.awsCfg.Credentials)
	if err != nil {
		return "", err
	}
	i.lock.Lock()
	defer i.lock.Unlock()
	i.token = token
	i.expiresAt = time.Now().Add(14 * time.Minute)
	return token, nil
}
