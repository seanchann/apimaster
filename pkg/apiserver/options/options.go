/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"net"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	cliflag "k8s.io/component-base/cli/flag"
)

type StorageBackendType string

const (
	StorageBackendTypeSqlite StorageBackendType = "sqlite"
	StorageBackendTypeEtcd   StorageBackendType = "etcd"
	StorageBackendTypeMysql  StorageBackendType = "mysql"
)

// DefaultServiceNodePortRange is the default port range for NodePort services.
var DefaultServiceNodePortRange = utilnet.PortRange{Base: 30000, Size: 2768}

// DefaultServiceIPCIDR is a CIDR notation of IP range from which to allocate service cluster IPs
var DefaultServiceIPCIDR net.IPNet = net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}

// DefaultEtcdPathPrefix default key path prefix in etcd
const DefaultEtcdPathPrefix = "/registry"

// APIMasterOptions
type APIMasterOptions struct {
	Backend                 StorageBackendType
	GenericServerRunOptions *genericoptions.ServerRunOptions
	Sqlite                  *SqliteOptions
	Etcd                    *genericoptions.EtcdOptions
	Mysql                   *MysqlOptions
	SecureServing           *genericoptions.SecureServingOptionsWithLoopback
	InsecureServing         *genericoptions.DeprecatedInsecureServingOptions
	Audit                   *genericoptions.AuditOptions
	Features                *genericoptions.FeatureOptions
	Authentication          *BuiltInAuthenticationOptions
	Authorization           *BuiltInAuthorizationOptions
	StorageSerialization    *StorageSerializationOptions
	APIEnablement           *genericoptions.APIEnablementOptions
	EgressSelector          *genericoptions.EgressSelectorOptions
	Traces                  *genericoptions.TracingOptions
	Admission               *AdmissionOptions
}

// NewAPIMasterOptions new a APIMasterOptions
func NewAPIMasterOptions(admission AdmissionProvider, backend StorageBackendType) *APIMasterOptions {
	o := &APIMasterOptions{
		Backend:                 backend,
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		SecureServing:           NewSecureServingOptions(),
		InsecureServing:         NewInsecureServingOptions(),
		Audit:                   genericoptions.NewAuditOptions(),
		Features:                genericoptions.NewFeatureOptions(),
		Authentication:          NewBuiltInAuthenticationOptions().WithWebHook(),
		Authorization:           NewBuiltInAuthorizationOptions(),
		StorageSerialization:    NewStorageSerializationOptions(),
		APIEnablement:           genericoptions.NewAPIEnablementOptions(),
		EgressSelector:          genericoptions.NewEgressSelectorOptions(),
		Traces:                  genericoptions.NewTracingOptions(),
		Admission:               NewAdmissionOptions(admission),
	}

	switch backend {
	case StorageBackendTypeSqlite:
		o.Sqlite = NewSqliteOptions(storagebackend.NewDefaultConfig("sqlite", nil))
	case StorageBackendTypeEtcd:
		o.Etcd = genericoptions.NewEtcdOptions(storagebackend.NewDefaultConfig(DefaultEtcdPathPrefix, nil))
	case StorageBackendTypeMysql:
		o.Mysql = NewMysqlOptions(storagebackend.NewDefaultConfig("mysql", nil))
	}

	return o
}

// AddFlags adds flags for a specific APIServer to the specified FlagSet
func (o *APIMasterOptions) AddFlags(fss *cliflag.NamedFlagSets) {
	o.GenericServerRunOptions.AddUniversalFlags(fss.FlagSet("generic"))
	o.SecureServing.AddFlags(fss.FlagSet("secure serving"))
	o.InsecureServing.AddFlags(fss.FlagSet("insecure serving"))
	o.Audit.AddFlags(fss.FlagSet("auditing"))
	o.Features.AddFlags(fss.FlagSet("features"))
	o.Authentication.AddFlags(fss.FlagSet("authentication"))
	o.Authorization.AddFlags(fss.FlagSet("authorization"))
	o.StorageSerialization.AddFlags(fss.FlagSet("storage serialization"))
	o.APIEnablement.AddFlags(fss.FlagSet("api enablement"))
	o.Admission.AddFlags(fss.FlagSet("admission"))

	switch o.Backend {
	case StorageBackendTypeSqlite:
		o.Sqlite.AddFlags(fss.FlagSet("sqlite"))
	case StorageBackendTypeEtcd:
		o.Etcd.AddFlags(fss.FlagSet("etcd"))
	case StorageBackendTypeMysql:
		o.Mysql.AddFlags(fss.FlagSet("mysql"))
	}

}
