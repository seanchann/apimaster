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
	cleanup       func()
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
	cleanup       func()
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

// Close releases resources held by a built configuration before ownership transfers to a server.
func (cfg *Config) Close() {
	if cfg != nil && cfg.cleanup != nil {
		cfg.cleanup()
	}
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete(informers informers.SharedInformerFactory) CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(informers),
		&cfg.ExtraConfig,
		cfg.cleanup,
	}

	return CompletedConfig{&c}
}

// New returns a new instance of WardleServer from the given config.
func (c completedConfig) New(delegateAPIServer genericapiserver.DelegationTarget) (result *APIServer, resultError error) {
	defer func() {
		if resultError != nil && c.cleanup != nil {
			c.cleanup()
		}
	}()
	genericServer, err := c.GenericConfig.New(c.ExtraConfig.APIServerName, delegateAPIServer)
	if err != nil {
		return nil, err
	}

	if c.cleanup != nil {
		genericServer.RegisterDestroyFunc(c.cleanup)
	}
	if closer, ok := c.GenericConfig.RESTOptionsGetter.(interface{ Close() error }); ok {
		genericServer.RegisterDestroyFunc(func() { closer.Close() })
	}
	defer func() {
		if resultError != nil {
			genericServer.Destroy()
		}
	}()
	if c.ExtraConfig.ExtendRoutesFunc != nil {
		c.ExtraConfig.ExtendRoutesFunc(genericServer.Handler.GoRestfulContainer)
	}

	gm := &APIServer{
		GenericAPIServer: genericServer,
	}

	// Construct the controller and return errors before exposing an incomplete server.
	if c.ExtraConfig.ControllerConfig.NewFunc != nil {
		provider, err := c.ExtraConfig.ControllerConfig.NewFunc(c.ExtraConfig.ControllerConfig.NewParameters)
		if err != nil {
			return nil, err
		}
		if provider == nil {
			return nil, fmt.Errorf("controller provider is required")
		}
		if err = gm.GenericAPIServer.AddPostStartHook(provider.Name(), provider.PostFunc()); err != nil {
			return nil, err
		}
		if err = gm.GenericAPIServer.AddPreShutdownHook(provider.Name(), provider.PreShutdownFunc()); err != nil {
			return nil, err
		}
		c.ExtraConfig.RESTStorageProviderBuilder = provider.RESTStorageProviderBuilderHandle()
	}

	if c.ExtraConfig.RESTStorageProviderBuilder == nil {
		return nil, fmt.Errorf("need rest storage provider builder")
	}

	restStorageProviders := c.ExtraConfig.RESTStorageProviderBuilder.NewProvider()
	apiResourceConfigSource := c.ExtraConfig.RESTStorageProviderBuilder.BuildAPIResouceConfigSource()
	if err := gm.InstallAPIs(apiResourceConfigSource, c.GenericConfig.RESTOptionsGetter, restStorageProviders...); err != nil {
		return nil, err
	}

	return gm, nil
}

// RESTStorageProviderBuilder a builder that construct []RESTStorageProvider for api install
type RESTStorageProviderBuilder interface {
	NewProvider() []RESTStorageProvider
	BuildAPIResouceConfigSource() serverstorage.APIResourceConfigSource
}

// RESTStorageProvider is a factory type for REST storage.
type RESTStorageProvider interface {
	GroupName() string
	NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, error)
}

// InstallAPIs will install the APIs for the restStorageProviders if they are enabled.
func (m *APIServer) InstallAPIs(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter, restStorageProviders ...RESTStorageProvider) error {
	for _, restStorageBuilder := range restStorageProviders {
		groupName := restStorageBuilder.GroupName()
		if !apiResourceConfigSource.AnyResourceForGroupEnabled(groupName) {
			klog.V(1).Infof("Skipping disabled API group %q.", groupName)
			continue
		}
		apiGroupInfo, err := restStorageBuilder.NewRESTStorage(apiResourceConfigSource, restOptionsGetter)
		installed := false
		// Until installation succeeds, this function owns every returned storage.
		defer func() {
			if !installed {
				for _, resources := range apiGroupInfo.VersionedResourcesStorageMap {
					for _, storage := range resources {
						storage.Destroy()
					}
				}
			}
		}()
		if err != nil {
			return fmt.Errorf("initialize API group %q: %w", groupName, err)
		}
		klog.V(1).Infof("Enabling API group %q.", groupName)

		if postHookProvider, ok := restStorageBuilder.(genericapiserver.PostStartHookProvider); ok {
			name, hook, err := postHookProvider.PostStartHook()
			if err != nil {
				return fmt.Errorf("build API group hook: %w", err)
			}
			if err = m.GenericAPIServer.AddPostStartHook(name, hook); err != nil {
				return err
			}
		}

		if err := m.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
			return fmt.Errorf("register API group: %w", err)
		}
		installed = true
	}
	return nil
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
