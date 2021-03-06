package levelds

import (
	"fmt"
	"path/filepath"

	"mbfs/go-mbfs/plugin"
	"mbfs/go-mbfs/repo"
	"mbfs/go-mbfs/repo/fsrepo"

	ldbopts "mbfs/go-mbfs/gx/QmbBhyDKsY4mbY6xsKt3qu9Y7FPvMJ6qbD8AMjYYvPRw1g/goleveldb/leveldb/opt"
	levelds "mbfs/go-mbfs/gx/QmccqjKZUTqp4ikWNyAbjBuP5HEdqSqRuAr9mcEhYab54a/go-ds-leveldb"
)

// Plugins is exported list of plugins that will be loaded
var Plugins = []plugin.Plugin{
	&leveldsPlugin{},
}

type leveldsPlugin struct{}

var _ plugin.PluginDatastore = (*leveldsPlugin)(nil)

func (*leveldsPlugin) Name() string {
	return "ds-level"
}

func (*leveldsPlugin) Version() string {
	return "0.1.0"
}

func (*leveldsPlugin) Init() error {
	return nil
}

func (*leveldsPlugin) DatastoreTypeName() string {
	return "levelds"
}

type datastoreConfig struct {
	path        string
	compression ldbopts.Compression
}

// BadgerdsDatastoreConfig returns a configuration stub for a badger datastore
// from the given parameters
func (*leveldsPlugin) DatastoreConfigParser() fsrepo.ConfigFromMap {
	return func(params map[string]interface{}) (fsrepo.DatastoreConfig, error) {
		var c datastoreConfig
		var ok bool

		c.path, ok = params["path"].(string)
		if !ok {
			return nil, fmt.Errorf("'path' field is missing or not string")
		}

		switch cm := params["compression"].(string); cm {
		case "none":
			c.compression = ldbopts.NoCompression
		case "snappy":
			c.compression = ldbopts.SnappyCompression
		case "":
			c.compression = ldbopts.DefaultCompression
		default:
			return nil, fmt.Errorf("unrecognized value for compression: %s", cm)
		}

		return &c, nil
	}
}

func (c *datastoreConfig) DiskSpec() fsrepo.DiskSpec {
	return map[string]interface{}{
		"type": "levelds",
		"path": c.path,
	}
}

func (c *datastoreConfig) Create(path string) (repo.Datastore, error) {
	p := c.path
	if !filepath.IsAbs(p) {
		p = filepath.Join(path, p)
	}

	return levelds.NewDatastore(p, &levelds.Options{
		Compression: c.compression,
	})
}
