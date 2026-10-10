package options

import (
	"fmt"
	nativemysql "github.com/go-sql-driver/mysql"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
)

// MysqlOptions mysql as a backend
type MysqlOptions struct {
	StorageConfig           storagebackend.Config
	DefaultStorageMediaType string
}

// NewMysqlOptions create  mysql options
func NewMysqlOptions(backendConfig *storagebackend.Config) *MysqlOptions {
	mysql := &MysqlOptions{
		StorageConfig:           *backendConfig,
		DefaultStorageMediaType: "application/json",
	}
	mysql.StorageConfig.Type = storagebackend.StorageTypeMysql

	return mysql
}

// Validate validate mysql input options
func (s *MysqlOptions) Validate() []error {
	if s == nil {
		return []error{fmt.Errorf("MySQL options are required")}
	}
	if len(s.StorageConfig.Mysql.ServerList) != 1 {
		return []error{fmt.Errorf("--mysql-servers requires exactly one database DSN")}
	}
	config, err := nativemysql.ParseDSN(s.StorageConfig.Mysql.ServerList[0])
	if err != nil || strings.TrimSpace(config.DBName) == "" {
		return []error{fmt.Errorf("--mysql-servers requires a valid database DSN")}
	}
	if config.Timeout < 0 || config.ReadTimeout < 0 || config.WriteTimeout < 0 {
		return []error{fmt.Errorf("MySQL deadlines cannot be negative")}
	}
	if s.StorageConfig.EventsHistoryWindow < 0 {
		return []error{fmt.Errorf("SQL history window cannot be negative")}
	}
	return nil
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet
// you must set storage-backend flag with mysql.
func (s *MysqlOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&s.StorageConfig.Mysql.ServerList, "mysql-servers", s.StorageConfig.Mysql.ServerList, ""+
		"specify server to connented backend.eg:user:password@tcp(host:port)/dbname, comma separated.")
	fs.Lookup("mysql-servers").DefValue = "<redacted>"

	fs.BoolVar(&s.StorageConfig.Mysql.Debug, "mysql-debug", s.StorageConfig.Mysql.Debug, ""+
		"enable mysql debug mode.")
	fs.IntVar(&s.StorageConfig.Mysql.ListDefaultLimit, "mysql-default-limit", s.StorageConfig.Mysql.ListDefaultLimit, ""+
		"the default limit for mysql query.")
	fs.StringVar(&s.StorageConfig.Prefix, "sql-prefix", s.StorageConfig.Prefix, "The full backend prefix, combined once with the resource prefix.")
	fs.DurationVar(&s.StorageConfig.EventsHistoryWindow, "sql-events-history-window", s.StorageConfig.EventsHistoryWindow, "Persistent event history retention; all users of a database must agree.")

}

// ApplyTo apply to server
func (s *MysqlOptions) ApplyTo(c *server.Config) error {
	if s == nil {
		return nil
	}
	return applySQLRESTOptions(c, s.StorageConfig, &SimpleRestOptionsFactory{Options: *s})
}

// ApplyWithStorageFactoryTo apply to storage factory
func (s *MysqlOptions) ApplyWithStorageFactoryTo(factory serverstorage.StorageFactory, c *server.Config) error {
	if s == nil || factory == nil {
		return fmt.Errorf("SQL options and storage factory are required")
	}
	return applySQLRESTOptions(c, s.StorageConfig, &storageFactoryRestOptionsFactory{Options: *s, StorageFactory: factory})
}

// SimpleRestOptionsFactory simple rest options factory
type SimpleRestOptionsFactory struct {
	Options MysqlOptions
}

// GetRESTOptions impl generic.RESTOptions
func (f *SimpleRestOptionsFactory) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (generic.RESTOptions, error) {
	ret := generic.RESTOptions{
		StorageConfig: &storagebackend.ConfigForResource{
			Config:        f.Options.StorageConfig,
			GroupResource: resource,
		},
		Decorator:                 generic.UndecoratedStorage,
		EnableGarbageCollection:   false,
		DeleteCollectionWorkers:   0,
		ResourcePrefix:            resource.Group + "/" + resource.Resource,
		CountMetricPollPeriod:     f.Options.StorageConfig.CountMetricPollPeriod,
		StorageObjectCountTracker: f.Options.StorageConfig.StorageObjectCountTracker,
	}
	return ret, nil
}

type storageFactoryRestOptionsFactory struct {
	Options        MysqlOptions
	StorageFactory serverstorage.StorageFactory
}

func (f *storageFactoryRestOptionsFactory) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (generic.RESTOptions, error) {
	storageConfig, err := f.StorageFactory.NewConfig(resource, example)
	if err != nil {
		return generic.RESTOptions{}, fmt.Errorf("unable to find storage destination for %v, due to %v", resource, err.Error())
	}

	ret := generic.RESTOptions{
		StorageConfig:             storageConfig,
		Decorator:                 generic.UndecoratedStorage,
		DeleteCollectionWorkers:   0,
		EnableGarbageCollection:   false,
		ResourcePrefix:            f.StorageFactory.ResourcePrefix(resource),
		CountMetricPollPeriod:     f.Options.StorageConfig.CountMetricPollPeriod,
		StorageObjectCountTracker: f.Options.StorageConfig.StorageObjectCountTracker,
	}

	return ret, nil
}
