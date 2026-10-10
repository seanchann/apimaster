/********************************************************************
* Copyright (c) 2008 - 2024. Authors: seanchann <seandev@foxmail.com>
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*         http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************/

package options

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
)

// SqliteOptions sqlite as a backend
type SqliteOptions struct {
	StorageConfig           storagebackend.Config
	DefaultStorageMediaType string
}

// NewSqliteOptions create  mysql options
func NewSqliteOptions(backendConfig *storagebackend.Config) *SqliteOptions {
	sqlite := &SqliteOptions{
		StorageConfig:           *backendConfig,
		DefaultStorageMediaType: "application/json",
	}
	sqlite.StorageConfig.Type = storagebackend.StorageTypeSqlite

	return sqlite
}

// Validate validate mysql input options
func (s *SqliteOptions) Validate() []error {
	if s == nil || strings.TrimSpace(s.StorageConfig.Sqlite.DSN) == "" {
		return []error{fmt.Errorf("--sqlite-dsn must be specified")}
	}
	parts := strings.SplitN(s.StorageConfig.Sqlite.DSN, "?", 2)
	if strings.TrimSpace(parts[0]) == "" {
		return []error{fmt.Errorf("SQLite DSN requires a database path")}
	}
	if len(parts) == 2 {
		if _, err := url.ParseQuery(parts[1]); err != nil {
			return []error{fmt.Errorf("invalid SQLite DSN query")}
		}
	}
	if s.StorageConfig.EventsHistoryWindow < 0 {
		return []error{fmt.Errorf("SQL history window cannot be negative")}
	}
	return nil
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet
// you must set storage-backend flag with mysql.
func (s *SqliteOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.StorageConfig.Sqlite.DSN, "sqlite-dsn", s.StorageConfig.Sqlite.DSN, ""+
		"specify server to connented backend.eg:./test.db?cache=share&mode=memory, comma separated.")
	fs.Lookup("sqlite-dsn").DefValue = "<redacted>"

	fs.BoolVar(&s.StorageConfig.Sqlite.Debug, "sqlite-debug", s.StorageConfig.Sqlite.Debug, ""+
		"enable sqlite debug mode.")
	fs.IntVar(&s.StorageConfig.Sqlite.ListDefaultLimit, "sqlite-default-limit", s.StorageConfig.Sqlite.ListDefaultLimit, ""+
		"the default limit for sqlite query.")
	fs.StringVar(&s.StorageConfig.Prefix, "sql-prefix", s.StorageConfig.Prefix, "The full backend prefix, combined once with the resource prefix.")
	fs.DurationVar(&s.StorageConfig.EventsHistoryWindow, "sql-events-history-window", s.StorageConfig.EventsHistoryWindow, "Persistent event history retention; all users of a database must agree.")

}

// ApplyTo apply to server
func (s *SqliteOptions) ApplyTo(c *server.Config) error {
	if s == nil {
		return nil
	}
	return applySQLRESTOptions(c, s.StorageConfig, &SqliteSimpleRestOptionsFactory{Options: *s})
}

// ApplyWithStorageFactoryTo apply to storage factory
func (s *SqliteOptions) ApplyWithStorageFactoryTo(factory serverstorage.StorageFactory, c *server.Config) error {
	if s == nil || factory == nil {
		return fmt.Errorf("SQL options and storage factory are required")
	}
	return applySQLRESTOptions(c, s.StorageConfig, &sqliteStorageFactoryRestOptionsFactory{Options: *s, StorageFactory: factory})
}

// SqliteSimpleRestOptionsFactory simple rest options factory
type SqliteSimpleRestOptionsFactory struct {
	Options SqliteOptions
}

// GetRESTOptions impl generic.RESTOptions
func (f *SqliteSimpleRestOptionsFactory) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (generic.RESTOptions, error) {
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

type sqliteStorageFactoryRestOptionsFactory struct {
	Options        SqliteOptions
	StorageFactory serverstorage.StorageFactory
}

func (f *sqliteStorageFactoryRestOptionsFactory) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (generic.RESTOptions, error) {
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
