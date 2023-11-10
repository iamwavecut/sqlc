package compiler

import (
	"context"
	"fmt"

	"github.com/iamwavecut/sqlc/internal/analyzer"
	"github.com/iamwavecut/sqlc/internal/config"
	"github.com/iamwavecut/sqlc/internal/engine/dolphin"
	"github.com/iamwavecut/sqlc/internal/engine/postgresql"
	pganalyze "github.com/iamwavecut/sqlc/internal/engine/postgresql/analyzer"
	"github.com/iamwavecut/sqlc/internal/engine/sqlite"
	"github.com/iamwavecut/sqlc/internal/opts"
	"github.com/iamwavecut/sqlc/internal/quickdb"
	pb "github.com/iamwavecut/sqlc/internal/quickdb/v1"
	"github.com/iamwavecut/sqlc/internal/sql/catalog"
)

type Compiler struct {
	conf     config.SQL
	combo    config.CombinedSettings
	catalog  *catalog.Catalog
	parser   Parser
	result   *Result
	analyzer analyzer.Analyzer
	client   pb.QuickClient

	schema []string
}

func NewCompiler(conf config.SQL, combo config.CombinedSettings) (*Compiler, error) {
	c := &Compiler{conf: conf, combo: combo}

	if conf.Database != nil && conf.Database.Managed {
		client, err := quickdb.NewClientFromConfig(combo.Global.Cloud)
		if err != nil {
			return nil, fmt.Errorf("client error: %w", err)
		}
		c.client = client
	}

	switch conf.Engine {
	case config.EngineSQLite:
		c.parser = sqlite.NewParser()
		c.catalog = sqlite.NewCatalog()
	case config.EngineMySQL:
		c.parser = dolphin.NewParser()
		c.catalog = dolphin.NewCatalog()
	case config.EnginePostgreSQL:
		c.parser = postgresql.NewParser()
		c.catalog = postgresql.NewCatalog()
		if conf.Database != nil {
			if conf.Analyzer.Database == nil || *conf.Analyzer.Database {
				c.analyzer = analyzer.Cached(
					pganalyze.New(c.client, *conf.Database),
					combo.Global,
					*conf.Database,
				)
			}
		}
	default:
		return nil, fmt.Errorf("unknown engine: %s", conf.Engine)
	}
	return c, nil
}

func (c *Compiler) Catalog() *catalog.Catalog {
	return c.catalog
}

func (c *Compiler) ParseCatalog(schema []string) error {
	return c.parseCatalog(schema)
}

func (c *Compiler) ParseQueries(queries []string, o opts.Parser) error {
	r, err := c.parseQueries(o)
	if err != nil {
		return err
	}
	c.result = r
	return nil
}

func (c *Compiler) Result() *Result {
	return c.result
}

func (c *Compiler) Close(ctx context.Context) {
	if c.analyzer != nil {
		c.analyzer.Close(ctx)
	}
}
