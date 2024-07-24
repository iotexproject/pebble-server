package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/pkg/errors"
	"github.com/xoctopus/datatypex"
)

type ClickHouse struct {
	Endpoint datatypex.Endpoint

	driver.Conn
}

func (c *ClickHouse) Init() error {
	options, err := clickhouse.ParseDSN(c.Endpoint.String())
	if err != nil {
		return errors.Wrap(err, "failed to parse dsn")
	}
	conn, err := clickhouse.Open(options)
	if err != nil {
		return errors.Wrap(err, "failed to create connection")
	}
	c.Conn = conn
	return nil
}
