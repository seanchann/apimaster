/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package apiserver

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/generic"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
)

//ControllerProvider new custom controller and return this
type ControllerProvider struct {
	NameFunc        func() string
	PostFunc        genericapiserver.PostStartHookFunc
	PreShutdownFunc genericapiserver.PreShutdownHookFunc

	//impl RESTStorageProviderBuilder interface
	RESTStorageProviderBuilder
}

//ControllerProviderConfig controller provider config
//this call before install api
type ControllerProviderConfig struct {
	//NewParameters user input parameter and apimaster input parameter
	//these all use with NewFunc
	NewParameters []interface{}
	NewFunc       func(para []interface{}) (*ControllerProvider, error)
}

//ExtraConfig user configure
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

//Config master config
type Config struct {
	GenericConfig *genericapiserver.Config
	ExtraConfig   ExtraConfig
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

//CompletedConfig complete config
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
			controllerName := provider.NameFunc()
			gm.GenericAPIServer.AddPostStartHookOrDie(controllerName, provider.PostFunc)
			gm.GenericAPIServer.AddPreShutdownHookOrDie(controllerName, provider.PreShutdownFunc)
			c.ExtraConfig.RESTStorageProviderBuilder = provider.RESTStorageProviderBuilder
		}
	}

	if c.ExtraConfig.RESTStorageProviderBuilder == nil {
		return nil, fmt.Errorf("need rest storage provider builder")
	}

	restStorageProviders := c.ExtraConfig.RESTStorageProviderBuilder.NewProvider()
	apiResourceConfigSource := c.ExtraConfig.RESTStorageProviderBuilder.BuildAPIResouceConfigSource()
	gm.InstallAPIs(apiResourceConfigSource, c.GenericConfig.RESTOptionsGetter, restStorageProviders...)

	return gm, nil
}

//RESTStorageProviderBuilder a builder that construct []RESTStorageProvider for api install
type RESTStorageProviderBuilder interface {
	NewProvider() []RESTStorageProvider
	BuildAPIResouceConfigSource() serverstorage.APIResourceConfigSource
}

// RESTStorageProvider is a factory type for REST storage.
type RESTStorageProvider interface {
	GroupName() string
	NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool)
}

// InstallAPIs will install the APIs for the restStorageProviders if they are enabled.
func (m *APIServer) InstallAPIs(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter, restStorageProviders ...RESTStorageProvider) {
	apiGroupsInfo := []genericapiserver.APIGroupInfo{}

	for _, restStorageBuilder := range restStorageProviders {
		groupName := restStorageBuilder.GroupName()
		if !apiResourceConfigSource.AnyVersionForGroupEnabled(groupName) {
			klog.V(1).Infof("Skipping disabled API group %q.", groupName)
			continue
		}
		apiGroupInfo, enabled := restStorageBuilder.NewRESTStorage(apiResourceConfigSource, restOptionsGetter)
		if !enabled {
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
