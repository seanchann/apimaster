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

package apiserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/version"
	apimachineryversion "k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/generic"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kube-openapi/pkg/common"
)

// ControllerProvider is a new custom controller that will be registered with the k8s/apiserver core.
// It will be called by the apiserver internally to invoke the client's implementation.
type ControllerProvider interface {
	Name() string
	PostFunc() genericapiserver.PostStartHookFunc
	PreShutdownFunc() genericapiserver.PreShutdownHookFunc
	RESTStorageProviderBuilderHandle() RESTStorageProviderBuilder
}

// ControllerProviderConfig controller provider config
// this call before install api
type ControllerProviderConfig struct {
	//NewParameters user input parameter and apimaster input parameter
	//these all use with NewFunc
	NewParameters []interface{}
	NewFunc       func(para []interface{}) (ControllerProvider, error)
}

// ExtraConfig user configure
type ExtraConfig struct {
	//APIServerName a name for this apiserver
	APIServerName string
	//StorageFactory serverstorage.StorageFactory

	ProxyTransport http.RoundTripper

	//ExtendRoutes add custom  route. will call this function to add
	ExtendRoutesFunc func(c *restful.Container)

	//RESTStorageProviderBuilder use this builder to crate RESTStorage
	RESTStorageProviderBuilder RESTStorageProviderBuilder

	//ControllerConfig config a controller
	ControllerConfig ControllerProviderConfig
}

// Config master config
type Config struct {
	GenericConfig *genericapiserver.Config
	ExtraConfig   ExtraConfig
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

// CompletedConfig complete config
type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// APIServer contains state for a  cluster master/apis server.
type APIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete(informers informers.SharedInformerFactory) CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(informers),
		&cfg.ExtraConfig,
	}

	c.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "0",
	}

	return CompletedConfig{&c}
}

// New returns a new instance of WardleServer from the given config.
func (c completedConfig) New(delegateAPIServer genericapiserver.DelegationTarget) (*APIServer, error) {
	genericServer, err := c.GenericConfig.New(c.ExtraConfig.APIServerName, genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	if c.ExtraConfig.ExtendRoutesFunc != nil {
		c.ExtraConfig.ExtendRoutesFunc(genericServer.Handler.GoRestfulContainer)
	}

	gm := &APIServer{
		GenericAPIServer: genericServer,
	}

	//add user config hook first
	if c.ExtraConfig.ControllerConfig.NewFunc != nil {
		if provider, err := c.ExtraConfig.ControllerConfig.NewFunc(c.ExtraConfig.ControllerConfig.NewParameters); err == nil {
			controllerName := provider.Name()
			gm.GenericAPIServer.AddPostStartHookOrDie(controllerName, provider.PostFunc())
			gm.GenericAPIServer.AddPreShutdownHookOrDie(controllerName, provider.PreShutdownFunc())

			c.ExtraConfig.RESTStorageProviderBuilder = provider.RESTStorageProviderBuilderHandle()
		}
	}

	if c.ExtraConfig.RESTStorageProviderBuilder == nil {
		return nil, fmt.Errorf("need rest storage provider builder")
	}

	restStorageProviders := c.ExtraConfig.RESTStorageProviderBuilder.NewProvider()
	apiResourceConfigSource := c.ExtraConfig.RESTStorageProviderBuilder.BuildAPIResourceConfigSource()
	gm.InstallAPIs(apiResourceConfigSource, c.GenericConfig.RESTOptionsGetter, restStorageProviders...)

	return gm, nil
}

// RESTStorageProviderBuilder a builder that construct []RESTStorageProvider for api install
type RESTStorageProviderBuilder interface {
	NewProvider() []RESTStorageProvider
	BuildAPIResourceConfigSource() serverstorage.APIResourceConfigSource
}

// RESTStorageProvider is a factory type for REST storage.
type RESTStorageProvider interface {
	GroupName() string
	NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, error)
}

// InstallAPIs will install the APIs for the restStorageProviders if they are enabled.
func (m *APIServer) InstallAPIs(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter, restStorageProviders ...RESTStorageProvider) {
	apiGroupsInfo := []genericapiserver.APIGroupInfo{}

	for _, restStorageBuilder := range restStorageProviders {
		groupName := restStorageBuilder.GroupName()
		if !apiResourceConfigSource.AnyResourceForGroupEnabled(groupName) {
			klog.V(1).Infof("Skipping disabled API group %q.", groupName)
			continue
		}
		apiGroupInfo, err := restStorageBuilder.NewRESTStorage(apiResourceConfigSource, restOptionsGetter)
		if err != nil {
			klog.Warningf("Problem initializing API group %q, skipping.", groupName)
			continue
		}
		klog.V(1).Infof("Enabling API group %q.", groupName)

		if postHookProvider, ok := restStorageBuilder.(genericapiserver.PostStartHookProvider); ok {
			name, hook, err := postHookProvider.PostStartHook()
			if err != nil {
				klog.Fatalf("Error building PostStartHook: %v", err)
			}
			m.GenericAPIServer.AddPostStartHookOrDie(name, hook)
		}

		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	for i := range apiGroupsInfo {
		if err := m.GenericAPIServer.InstallAPIGroup(&apiGroupsInfo[i]); err != nil {
			klog.Fatalf("Error in registering group versions: %v", err)
		}
	}
}

// APIServerProvider is an interface for APIServer to provide server information
// It is a callback function for constructing ControllerProvider
type APIServerProvider interface {
	APIName() string
	Version() *apimachineryversion.Info
	DefaultAPIResourceConfigSource() *serverstorage.ResourceConfig
	DefaultInstallExtendRoutes(c *restful.Container)
	NewControllerProvider(para []interface{}) (ControllerProvider, error)
	GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition

	//for client
	ClientNewForConfig(c *rest.Config) (interface{}, error)
	ClientNewSharedInformerFactory(interface{}, time.Duration) interface{}
}
