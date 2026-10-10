/********************************************************************
* Copyright (c) All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************/

package compatibility_test

import (
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/seanchann/apimaster/pkg/apiserver"
	"github.com/seanchann/apimaster/pkg/apiserver/options"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/registry/generic"
	server "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/client-go/rest"
	"k8s.io/kube-openapi/pkg/common"
)

type donorProvider struct{}

func (donorProvider) APIName() string        { return "compatibility" }
func (donorProvider) Version() *version.Info { return &version.Info{} }
func (donorProvider) DefaultAPIResourceConfigSource() *serverstorage.ResourceConfig {
	return serverstorage.NewResourceConfig()
}
func (donorProvider) DefaultInstallExtendRoutes(*restful.Container) {}
func (donorProvider) NewControllerProvider([]interface{}) (apiserver.ControllerProvider, error) {
	return donorController{}, nil
}
func (donorProvider) GetOpenAPIDefinitions(common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return nil
}
func (donorProvider) ClientNewForConfig(*rest.Config) (interface{}, error) { return nil, nil }
func (donorProvider) ClientNewSharedInformerFactory(interface{}, time.Duration) interface{} {
	return nil
}

type donorController struct{}

func (donorController) Name() string                                { return "compatibility" }
func (donorController) PostFunc() server.PostStartHookFunc          { return nil }
func (donorController) PreShutdownFunc() server.PreShutdownHookFunc { return nil }
func (donorController) RESTStorageProviderBuilderHandle() apiserver.RESTStorageProviderBuilder {
	return donorBuilder{}
}

type donorBuilder struct{}

func (donorBuilder) NewProvider() []apiserver.RESTStorageProvider {
	return []apiserver.RESTStorageProvider{donorStorage{}}
}
func (donorBuilder) BuildAPIResouceConfigSource() serverstorage.APIResourceConfigSource {
	return serverstorage.NewResourceConfig()
}

type donorStorage struct{}

func (donorStorage) GroupName() string { return "compatibility" }
func (donorStorage) NewRESTStorage(serverstorage.APIResourceConfigSource, generic.RESTOptionsGetter) (server.APIGroupInfo, error) {
	return server.APIGroupInfo{}, nil
}

type donorAdmission struct{}

func (donorAdmission) RegisterAllAdmissionPlugins(*admission.Plugins) {}
func (donorAdmission) AllPluginOrder() []string {
	return genericoptions.NewAdmissionOptions().RecommendedPluginOrder
}
func (donorAdmission) DefaultOffAdmissionPlugins() sets.String { return sets.NewString() }

var (
	_ apiserver.APIServerProvider                                                           = donorProvider{}
	_ apiserver.ControllerProvider                                                          = donorController{}
	_ apiserver.RESTStorageProviderBuilder                                                  = donorBuilder{}
	_ apiserver.RESTStorageProvider                                                         = donorStorage{}
	_ options.AdmissionProvider                                                             = donorAdmission{}
	_ func(options.AdmissionProvider, options.StorageBackendType) *options.APIMasterOptions = options.NewAPIMasterOptions
	_ func(*options.APIMasterOptions, apiserver.APIServerProvider, <-chan struct{}) error   = apiserver.APIServerRun
)
