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

//MysqlOptions mysql as a backend
type MysqlOptions struct {
	StorageConfig           storagebackend.Config
	DefaultStorageMediaType string
}

//NewMysqlOptions create  mysql options
func NewMysqlOptions(backendConfig *storagebackend.Config) *MysqlOptions {
	mysql := &MysqlOptions{
		StorageConfig:           *backendConfig,
		DefaultStorageMediaType: "application/json",
	}
	mysql.StorageConfig.Type = storagebackend.StorageTypeMysql

	return mysql
}

//Validate validate mysql input options
func (s *MysqlOptions) Validate() []error {
	allErrors := []error{}
	if len(s.StorageConfig.Mysql.ServerList) == 0 {
		allErrors = append(allErrors, fmt.Errorf("--mysql-servers must be specified"))
	}
	return allErrors
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet
// you must set storage-backend flag with mysql.
func (s *MysqlOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&s.StorageConfig.Mysql.ServerList, "mysql-servers", s.StorageConfig.Mysql.ServerList, ""+
		"specify server to connented backend.eg:user:password@tcp(host:port)/dbname, comma separated.")
}

//ApplyTo apply to server
func (s *MysqlOptions) ApplyTo(c *server.Config) error {
	if s == nil {
		return nil
	}
	c.RESTOptionsGetter = &SimpleRestOptionsFactory{Options: *s}
	return nil
}

//ApplyWithStorageFactoryTo apply to storage factory
func (s *MysqlOptions) ApplyWithStorageFactoryTo(factory serverstorage.StorageFactory, c *server.Config) error {
	c.RESTOptionsGetter = &storageFactoryRestOptionsFactory{Options: *s, StorageFactory: factory}
	return nil
}

//SimpleRestOptionsFactory simple rest options factory
type SimpleRestOptionsFactory struct {
	Options MysqlOptions
}

//GetRESTOptions impl generic.RESTOptions
func (f *SimpleRestOptionsFactory) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
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

type storageFactoryRestOptionsFactory struct {
	Options        MysqlOptions
	StorageFactory serverstorage.StorageFactory
}

func (f *storageFactoryRestOptionsFactory) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
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
