package iaas

import (
	"time"

	"fmt"

	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/open-service-broker-sacloud/service/params"
	"github.com/sacloud/open-service-broker-sacloud/util/mutexkv"
	"github.com/sacloud/open-service-broker-sacloud/version"
)

var mutex = mutexkv.NewMutexKV()

// ClientConfig represents SAKURA Cloud API client config
type ClientConfig struct {
	AccessToken       string
	AccessTokenSecret string
	Zone              string
	AcceptLanguage    string
	RetryMax          int
	RetryIntervalSec  int64
	APIRootURL        string
	TraceMode         bool
}

// Client is SAKURA Cloud API facade interface
type Client interface {
	AuthStatus() (*sacloud.AuthStatus, error)
	MariaDB() DatabaseAPI
	PostgreSQL() DatabaseAPI
}

// DatabaseAPI is SAKURA Cloud Database API interface
type DatabaseAPI interface {
	Read(instanceID string) (*sacloud.Database, error)
	Create(instanceID string, param *params.DatabaseCreateParameter) (*sacloud.Database, error)
	Delete(instanceID string) error
}

const markerTag = "@open-service-broker-sacloud"

type client struct {
	rawClient  *api.Client
	mariaDB    *dbApplianceClient
	postgreSQL *dbApplianceClient
}

// NewClient returns SAKURA Cloud API client
func NewClient(cfg *ClientConfig) Client {
	c := api.NewClient(cfg.AccessToken, cfg.AccessTokenSecret, cfg.Zone)

	c.UserAgent = fmt.Sprintf("oepn-service-broker-sacloud/v%s", version.Version)

	if cfg.AcceptLanguage != "" {
		c.AcceptLanguage = cfg.AcceptLanguage
	}
	if cfg.RetryMax > 0 {
		c.RetryMax = cfg.RetryMax
	}
	if cfg.RetryIntervalSec > 0 {
		c.RetryInterval = time.Duration(cfg.RetryIntervalSec) * time.Second
	}
	if cfg.TraceMode {
		c.TraceMode = true
	}
	if cfg.APIRootURL != "" {
		api.SakuraCloudAPIRoot = cfg.APIRootURL
	}
	client := &client{rawClient: c}
	client.mariaDB = &dbApplianceClient{
		client:          client,
		createParamFunc: sacloud.NewCreateMariaDBDatabaseValue,
	}
	client.postgreSQL = &dbApplianceClient{
		client:          client,
		createParamFunc: sacloud.NewCreatePostgreSQLDatabaseValue,
	}
	return client
}

func (c *client) getRawClient() *api.Client {
	return c.rawClient.Clone()
}

func (c *client) MariaDB() DatabaseAPI {
	return c.mariaDB
}

func (c *client) PostgreSQL() DatabaseAPI {
	return c.postgreSQL
}
