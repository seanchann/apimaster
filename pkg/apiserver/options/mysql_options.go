package options

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/storage/storagebackend"

	"github.com/spf13/pflag"
)

//MysqlOptions mysql as a backend
type MysqlOptions struct {
	StorageConfig storagebackend.Config
}

//NewMysqlOptions create  mysql options
func NewMysqlOptions(backendConfig *storagebackend.Config) *MysqlOptions {
	mysql := &MysqlOptions{
		StorageConfig: *backendConfig,
	}
	mysql.StorageConfig.Type = storagebackend.StorageTypeMysql

	return mysql
}

//Validate validate mysql input options
func (s *MysqlOptions) Validate() []error {
	allErrors := []error{}
	if len(s.StorageConfig.ServerList) == 0 {
		allErrors = append(allErrors, fmt.Errorf("--mysql-servers must be specified"))
	}
	return allErrors
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet
// you must set storage-backend flag with mysql.
func (s *MysqlOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&s.StorageConfig.ServerList, "mysql-servers", s.StorageConfig.ServerList, ""+
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
