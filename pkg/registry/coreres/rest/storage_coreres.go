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

package rest

import (
	apicommres "github.com/seanchann/apimaster/pkg/apis/coreres"
	apicommresv1 "github.com/seanchann/apimaster/pkg/apis/coreres/v1"
	namespacestore "github.com/seanchann/apimaster/pkg/registry/coreres/namespace/storage"

	"github.com/seanchann/apimaster/pkg/api/legacyscheme"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/klog/v2"
)

// RESTStorageProvider providers information needed to build RESTStorage for core.
type RESTStorageProvider struct {
}

// NewRESTStorage create a RESTStorage provider
func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apicommres.GroupName,
		legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)

	if storageMap, err := p.v1Storage(apicommresv1.SchemeGroupVersion, apiResourceConfigSource, restOptionsGetter); err != nil {
		return genericapiserver.APIGroupInfo{}, err
	} else if len(storageMap) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap[apicommresv1.SchemeGroupVersion.Version] = storageMap
	}

	return apiGroupInfo, nil
}

func (p RESTStorageProvider) v1Storage(version schema.GroupVersion, apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (map[string]rest.Storage, error) {
	klog.Infof("install core resource group rest")

	storage := map[string]rest.Storage{}

	if resource := "namespaces"; apiResourceConfigSource.ResourceEnabled(apicommresv1.SchemeGroupVersion.WithResource(resource)) {
		namespaceStorage, namespaceStatusStorage, namespaceFinalizeStorage, err := namespacestore.NewREST(restOptionsGetter)
		if err != nil {
			return storage, err
		}
		storage[resource] = namespaceStorage
		storage[resource+"/status"] = namespaceStatusStorage
		storage[resource+"/finalize"] = namespaceFinalizeStorage
	}

	return storage, nil
}

// GroupName return ami group name
func (p RESTStorageProvider) GroupName() string {
	return apicommres.GroupName
}
