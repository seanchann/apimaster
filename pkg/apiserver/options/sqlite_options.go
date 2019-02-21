package options

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
)

//SqliteOptions sqlite as a backend
type SqliteOptions struct {
	StorageConfig           storagebackend.Config
	DefaultStorageMediaType string
}

//NewSqliteOptions create  mysql options
func NewSqliteOptions(backendConfig *storagebackend.Config) *SqliteOptions {
	sqlite := &SqliteOptions{
		StorageConfig:           *backendConfig,
		DefaultStorageMediaType: "application/json",
	}
	sqlite.StorageConfig.Type = storagebackend.StorageTypeSqlite

	return sqlite
}

//Validate validate mysql input options
func (s *SqliteOptions) Validate() []error {
	allErrors := []error{}
	if len(s.StorageConfig.Sqlite.DSN) == 0 {
		allErrors = append(allErrors, fmt.Errorf("--sqlite-dsn must be specified"))
	}
	return allErrors
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet
// you must set storage-backend flag with mysql.
func (s *SqliteOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.StorageConfig.Sqlite.DSN, "sqlite-dsn", s.StorageConfig.Sqlite.DSN, ""+
		"specify server to connented backend.eg:./test.db?cache=share&mode=memory, comma separated.")

	fs.BoolVar(&s.StorageConfig.Sqlite.Debug, "sqlite-debug", s.StorageConfig.Sqlite.Debug, ""+
		"enable sqlite debug mode.")
	fs.IntVar(&s.StorageConfig.Sqlite.ListDefaultLimit, "sqlite-default-limit", s.StorageConfig.Sqlite.ListDefaultLimit, ""+
		"the default limit for sqlite query.")
}

//ApplyTo apply to server
func (s *SqliteOptions) ApplyTo(c *server.Config) error {
	if s == nil {
		return nil
	}
	c.RESTOptionsGetter = &SqliteSimpleRestOptionsFactory{Options: *s}
	return nil
}

//ApplyWithStorageFactoryTo apply to storage factory
func (s *SqliteOptions) ApplyWithStorageFactoryTo(factory serverstorage.StorageFactory, c *server.Config) error {
	c.RESTOptionsGetter = &sqliteStorageFactoryRestOptionsFactory{Options: *s, StorageFactory: factory}
	return nil
}

//SqliteSimpleRestOptionsFactory simple rest options factory
type SqliteSimpleRestOptionsFactory struct {
	Options SqliteOptions
}

//GetRESTOptions impl generic.RESTOptions
func (f *SqliteSimpleRestOptionsFactory) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	ret := generic.RESTOptions{
		StorageConfig:           &f.Options.StorageConfig,
		Decorator:               generic.UndecoratedStorage,
		EnableGarbageCollection: false,
		DeleteCollectionWorkers: 0,
		ResourcePrefix:          resource.Group + "/" + resource.Resource,
		CountMetricPollPeriod:   0,
	}
	return ret, nil
}

type sqliteStorageFactoryRestOptionsFactory struct {
	Options        SqliteOptions
	StorageFactory serverstorage.StorageFactory
}

func (f *sqliteStorageFactoryRestOptionsFactory) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	storageConfig, err := f.StorageFactory.NewConfig(resource)
	if err != nil {
		return generic.RESTOptions{}, fmt.Errorf("unable to find storage destination for %v, due to %v", resource, err.Error())
	}

	ret := generic.RESTOptions{
		StorageConfig:           storageConfig,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: 0,
		EnableGarbageCollection: false,
		ResourcePrefix:          f.StorageFactory.ResourcePrefix(resource),
		CountMetricPollPeriod:   0,
	}

	return ret, nil
}
