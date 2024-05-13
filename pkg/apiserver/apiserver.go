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
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/seanchann/apimaster/pkg/api/legacyscheme"
	apiserveradmission "github.com/seanchann/apimaster/pkg/apiserver/admission"
	"github.com/seanchann/apimaster/pkg/apiserver/options"
	insecureserver "github.com/seanchann/apimaster/pkg/apiserver/server"

	//k8s dependencies
	oteltrace "go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	openapinamer "k8s.io/apiserver/pkg/endpoints/openapi"
	genericfeatures "k8s.io/apiserver/pkg/features"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/apiserver/pkg/server/filters"
	apiserverstorage "k8s.io/apiserver/pkg/server/storage"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/apiserver/pkg/util/openapi"
	"k8s.io/apiserver/pkg/util/webhook"
	"k8s.io/client-go/dynamic"
	k8sclientgoinformers "k8s.io/client-go/informers"
	k8sclientgoclientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	openapicommon "k8s.io/kube-openapi/pkg/common"
)

func APIServerRun(opt *options.APIServerRunOptions, apiProvider APIServerProvider, stopCh <-chan struct{}) error {
	// set default options
	completedOptions, err := Complete(opt)
	if err != nil {
		return err
	}
	// validate options
	if errs := completedOptions.Validate(); len(errs) != 0 {
		return utilerrors.NewAggregate(errs)
	}

	if err := applyExternalConfig(completedOptions); err != nil {
		return err
	}

	//block call util occure a error
	return Run(completedOptions, stopCh)
}

func applyExternalConfig(completeOptions completedServerRunOptions) error {
	//insecure only support for localhost
	if addr := net.ParseIP("127.0.0.1"); addr != nil {
		completeOptions.InsecureServing.BindAddress = addr
	}

	switch completeOptions.Backend {
	case options.StorageBackendTypeMysql:
	case options.StorageBackendTypeEtcd:
		fallthrough
	default:
	}

	return nil
}

// Run runs the specified APIServer.  This should never exit.
func Run(completeOptions completedServerRunOptions, stopCh <-chan struct{}) error {
	server, err := CreateServerChain(completeOptions, stopCh)
	if err != nil {
		return err
	}
	return server.PrepareRun().Run(stopCh)
}

// CreateAPIServer creates and wires a workable agent-apiserver
func CreateAPIServer(apiServerConfig *Config,
	delegateAPIServer genericapiserver.DelegationTarget,
	versionedInformers k8sclientgoinformers.SharedInformerFactory) (*APIServer, error) {
	apiServer, err := apiServerConfig.Complete(versionedInformers).New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return apiServer, nil
}

// CreateProxyTransport creates the dialer infrastructure to connect to the nodes.
func CreateProxyTransport() *http.Transport {
	var proxyDialerFn utilnet.DialFunc
	// Proxying to pods and services is IP-based... don't expect to be able to verify the hostname
	proxyTLSClientConfig := &tls.Config{InsecureSkipVerify: true}
	proxyTransport := utilnet.SetTransportDefaults(&http.Transport{
		DialContext:     proxyDialerFn,
		TLSClientConfig: proxyTLSClientConfig,
	})
	return proxyTransport
}

// CreateServerChain creates the apiservers connected via delegation.
func CreateServerChain(completedOptions completedServerRunOptions, stopCh <-chan struct{}) (
	*genericapiserver.GenericAPIServer, error) {

	proxyTransport := CreateProxyTransport()

	apiServerCfg, insecureServingInfo, _,
		normalVersionedInformers, err := BuildGenericConfig(completedOptions,
		[]*runtime.Scheme{legacyscheme.Scheme},
		completedOptions.apiProvider.GetOpenAPIDefinitions)
	if err != nil {
		return nil, err
	}

	// setup admission
	admissionConfig := &apiserveradmission.Config{
		ExternalInformers:    normalVersionedInformers,
		LoopbackClientConfig: apiServerCfg.GenericConfig.LoopbackClientConfig,
	}
	serviceResolver := buildServiceResolver(false,
		apiServerCfg.GenericConfig.LoopbackClientConfig.Host, normalVersionedInformers)
	pluginInitializers, _, err := admissionConfig.New(proxyTransport,
		apiServerCfg.GenericConfig.EgressSelector, serviceResolver, apiServerCfg.GenericConfig.TracerProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create admission plugin initializer: %v", err)
	}
	clientgoExternalClientAdmission, err := completedOptions.apiProvider.ClientNewForConfig(apiServerCfg.GenericConfig.LoopbackClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create real client-go external client: %w", err)
	}
	dynamicExternalClient, err := dynamic.NewForConfig(apiServerCfg.GenericConfig.LoopbackClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create real dynamic external client: %w", err)
	}
	err = completedOptions.Admission.ApplyTo(
		apiServerCfg.GenericConfig,
		normalVersionedInformers,
		clientgoExternalClientAdmission,
		dynamicExternalClient,
		utilfeature.DefaultFeatureGate,
		pluginInitializers...)
	if err != nil {
		return nil, fmt.Errorf("failed to apply admission: %w", err)
	}
	// if err := apiServerCfg.GenericConfig.AddPostStartHook("start-apiserver-admission-initializer", admissionPostStartHook); err != nil {
	// 	return nil, err
	// }

	apiServer, err := CreateAPIServer(apiServerCfg, genericapiserver.NewEmptyDelegate(), nil)
	// apiServer, err := CreateAPIServer(apiServerCfg, genericapiserver.NewEmptyDelegate(), versionedInformers)
	if err != nil {
		return nil, err
	}

	if insecureServingInfo != nil {
		insecureHandlerChain := insecureserver.BuildInsecureHandlerChain(apiServer.GenericAPIServer.UnprotectedHandler(),
			apiServerCfg.GenericConfig, true)
		if err := insecureServingInfo.Serve(insecureHandlerChain, apiServerCfg.GenericConfig.RequestTimeout, stopCh); err != nil {
			return nil, err
		}
	}

	return apiServer.GenericAPIServer, nil
}

// BuildGenericConfig creates all the resources for running the API server, but runs none of them
func BuildGenericConfig(s completedServerRunOptions,
	schemes []*runtime.Scheme,
	getOpenAPIDefinitions func(ref openapicommon.ReferenceCallback) map[string]openapicommon.OpenAPIDefinition) (
	config *Config,
	insecureServingInfo *genericapiserver.DeprecatedInsecureServingInfo,
	k8sversionedInformers k8sclientgoinformers.SharedInformerFactory,
	versionedInformers interface{},
	lastErr error,
) {

	genericConfig := genericapiserver.NewConfig(legacyscheme.Codecs)

	if lastErr = s.GenericServerRunOptions.ApplyTo(genericConfig); lastErr != nil {
		return
	}
	if lastErr = s.InsecureServing.ApplyTo(&insecureServingInfo); lastErr != nil {
		return
	}
	if lastErr = s.SecureServing.ApplyTo(&genericConfig.SecureServing, &genericConfig.LoopbackClientConfig); lastErr != nil {
		return
	}

	// Use protobufs for self-communication.
	// Since not every generic apiserver has to support protobufs, we
	// cannot default to it in generic apiserver and need to explicitly
	// set it in kube-apiserver.
	// genericConfig.LoopbackClientConfig.ContentConfig.ContentType = "application/vnd.kubernetes.protobuf"
	genericConfig.LoopbackClientConfig.ContentConfig.ContentType = "application/json"
	// Disable compression for self-communication, since we are going to be
	// on a fast local network
	genericConfig.LoopbackClientConfig.DisableCompression = true

	clientConfig := genericConfig.LoopbackClientConfig
	clientgoExternalClient, err := k8sclientgoclientset.NewForConfig(clientConfig)
	if err != nil {
		lastErr = fmt.Errorf("failed to create real external clientset: %v", err)
		return
	}
	k8sversionedInformers = k8sclientgoinformers.NewSharedInformerFactory(clientgoExternalClient, 10*time.Minute)

	versionClient, err := s.apiProvider.ClientNewForConfig(genericConfig.LoopbackClientConfig)
	if err != nil {
		lastErr = fmt.Errorf("failed to create clientset: %v", err)
		return
	}
	versionedInformers = s.apiProvider.ClientNewSharedInformerFactory(versionClient, 10*time.Minute)

	if lastErr = s.APIEnablement.ApplyTo(genericConfig, s.apiProvider.DefaultAPIResourceConfigSource(), legacyscheme.Scheme); lastErr != nil {
		return
	}

	if lastErr = s.Features.ApplyTo(genericConfig); lastErr != nil {
		return
	}
	if lastErr = s.EgressSelector.ApplyTo(genericConfig); lastErr != nil {
		return
	}
	if utilfeature.DefaultFeatureGate.Enabled(genericfeatures.APIServerTracing) {
		if lastErr = s.Traces.ApplyTo(genericConfig.EgressSelector, genericConfig); lastErr != nil {
			return
		}
	}

	// wrap the definitions to revert any changes from disabled features
	getOpenAPIDefinitions = openapi.GetOpenAPIDefinitionsWithoutDisabledFeatures(getOpenAPIDefinitions)
	namer := openapinamer.NewDefinitionNamer(schemes...)
	genericConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(getOpenAPIDefinitions, namer)
	genericConfig.OpenAPIConfig.Info.Title = s.apiProvider.APIName() //"SIPDispatch"
	genericConfig.OpenAPIV3Config = genericapiserver.DefaultOpenAPIV3Config(getOpenAPIDefinitions, namer)
	genericConfig.OpenAPIV3Config.Info.Title = s.apiProvider.APIName()

	genericConfig.LongRunningFunc = filters.BasicLongRunningRequestCheck(
		sets.NewString("watch", "proxy"),
		sets.NewString("attach", "exec", "proxy", "log", "portforward"),
	)

	genericConfig.Version = s.apiProvider.Version()

	// only for etcd
	if s.Backend == options.StorageBackendTypeEtcd {

		if genericConfig.EgressSelector != nil {
			s.Etcd.StorageConfig.Transport.EgressLookup = genericConfig.EgressSelector.Lookup
		}
		if utilfeature.DefaultFeatureGate.Enabled(genericfeatures.APIServerTracing) {
			s.Etcd.StorageConfig.Transport.TracerProvider = genericConfig.TracerProvider
		} else {
			s.Etcd.StorageConfig.Transport.TracerProvider = oteltrace.NewNoopTracerProvider()
		}
	}
	storageFactory, lastErr := BuildStorageFactory(s.APIServerRunOptions, genericConfig.MergedResourceConfig)
	if lastErr != nil {
		return
	}
	switch s.Backend {
	case options.StorageBackendTypeSqlite:
		if lastErr = s.Sqlite.ApplyWithStorageFactoryTo(storageFactory, genericConfig); lastErr != nil {
			return
		}
	case options.StorageBackendTypeEtcd:
		if lastErr = s.Etcd.ApplyWithStorageFactoryTo(storageFactory, genericConfig); lastErr != nil {
			return
		}
	case options.StorageBackendTypeMysql:
		if lastErr = s.Mysql.ApplyWithStorageFactoryTo(storageFactory, genericConfig); lastErr != nil {
			return
		}
	}

	klog.Infof("Successfully applied configuration authentication")
	if lastErr = s.Authentication.ApplyTo(&genericConfig.Authentication,
		genericConfig.SecureServing, genericConfig.EgressSelector,
		genericConfig.OpenAPIConfig, genericConfig.OpenAPIV3Config); lastErr != nil {
		return
	}

	klog.Infof("Successfully applied configuration authorization")
	var enablesRBAC bool
	genericConfig.Authorization.Authorizer, genericConfig.RuleResolver, enablesRBAC,
		err = BuildAuthorizer(s.APIServerRunOptions, genericConfig.EgressSelector)
	if err != nil {
		lastErr = fmt.Errorf("invalid authorization config: %v", err)
		return
	}
	if s.Authorization != nil && !enablesRBAC {
		// TODO: remove this once we have a way to disable RBAC
		// current we always disable internal RBAC
		genericConfig.DisabledPostStartHooks.Insert("rbac/bootstrap-roles")
	}

	lastErr = s.Audit.ApplyTo(genericConfig)
	if lastErr != nil {
		return
	}

	config = &Config{
		GenericConfig: genericConfig,
		ExtraConfig:   ExtraConfig{},
	}

	config.ExtraConfig.ExtendRoutesFunc = s.apiProvider.DefaultInstallExtendRoutes
	config.ExtraConfig.ControllerConfig.NewFunc = s.apiProvider.NewAPIServerProvider
	//append our private parameter for controller
	config.ExtraConfig.ControllerConfig.NewParameters = append(config.ExtraConfig.ControllerConfig.NewParameters, versionClient)
	config.ExtraConfig.ControllerConfig.NewParameters = append(config.ExtraConfig.ControllerConfig.NewParameters, clientgoExternalClient)

	return
}

// completedServerRunOptions is a private wrapper that enforces a call of Complete() before Run can be invoked.
type completedServerRunOptions struct {
	*options.APIServerRunOptions
	apiProvider APIServerProvider
}

// Complete set default ServerRunOptions.
// Should be called after server flags parsed.
func Complete(s *options.APIServerRunOptions) (completedServerRunOptions, error) {
	var options completedServerRunOptions
	options.APIServerRunOptions = s
	return options, nil
}

// BuildAuthorizer constructs the authorizer
func BuildAuthorizer(s *options.APIServerRunOptions, egressSelector *egressselector.EgressSelector) (authorizer.Authorizer, authorizer.RuleResolver, bool, error) {
	authorizationConfig, err := s.Authorization.ToAuthorizationConfig()
	if err != nil {
		return nil, nil, false, err
	}
	if authorizationConfig == nil {
		return nil, nil, false, nil
	}

	if egressSelector != nil {
		egressDialer, err := egressSelector.Lookup(egressselector.ControlPlane.AsNetworkContext())
		if err != nil {
			return nil, nil, false, err
		}
		authorizationConfig.CustomDial = egressDialer
	}

	authorizer, ruleResolver, err := authorizationConfig.New()
	if err != nil {
		klog.Infof("error building authorizer: %v", err)
	}

	//always return false， because we don't want to use the internal rbac
	return authorizer, ruleResolver, false, nil
}

// BuildStorageFactory constructs the storage factory. If encryption at rest is used, it expects
// all supported KMS plugins to be registered in the KMS plugin registry before being called.
func BuildStorageFactory(s *options.APIServerRunOptions, apiResourceConfig *apiserverstorage.ResourceConfig) (*apiserverstorage.DefaultStorageFactory, error) {
	storageGroupsToEncodingVersion, err := s.StorageSerialization.StorageGroupsToEncodingVersion()
	if err != nil {
		return nil, fmt.Errorf("error generating storage version map: %s", err)
	}

	switch s.Backend {
	case options.StorageBackendTypeSqlite:
		storageFactory, err := NewStorageFactory(
			s.Sqlite.StorageConfig, s.Sqlite.DefaultStorageMediaType, legacyscheme.Codecs,
			apiserverstorage.NewDefaultResourceEncodingConfig(legacyscheme.Scheme), storageGroupsToEncodingVersion,
			// The list includes resources that need to be stored in a different
			// group version than other resources in the groups.
			// FIXME (soltysh): this GroupVersionResource override should be configurable
			[]schema.GroupVersionResource{},
			apiResourceConfig)
		if err != nil {
			return nil, fmt.Errorf("error in initializing storage factory: %s", err)
		}
		return storageFactory, nil
	case options.StorageBackendTypeEtcd:
		storageFactory, err := NewStorageFactory(
			s.Etcd.StorageConfig, s.Etcd.DefaultStorageMediaType, legacyscheme.Codecs,
			apiserverstorage.NewDefaultResourceEncodingConfig(legacyscheme.Scheme), storageGroupsToEncodingVersion,
			// The list includes resources that need to be stored in a different
			// group version than other resources in the groups.
			// FIXME (soltysh): this GroupVersionResource override should be configurable
			[]schema.GroupVersionResource{},
			apiResourceConfig)
		if err != nil {
			return nil, fmt.Errorf("error in initializing storage factory: %s", err)
		}

		return storageFactory, nil
	case options.StorageBackendTypeMysql:
		storageFactory, err := NewStorageFactory(
			s.Mysql.StorageConfig, s.Mysql.DefaultStorageMediaType, legacyscheme.Codecs,
			apiserverstorage.NewDefaultResourceEncodingConfig(legacyscheme.Scheme), storageGroupsToEncodingVersion,
			// The list includes resources that need to be stored in a different
			// group version than other resources in the groups.
			// FIXME (soltysh): this GroupVersionResource override should be configurable
			[]schema.GroupVersionResource{},
			apiResourceConfig)
		if err != nil {
			return nil, fmt.Errorf("error in initializing storage factory: %s", err)
		}

		return storageFactory, nil
	}

	return nil, fmt.Errorf("not configure any storage backend")
}

func buildServiceResolver(enabledAggregatorRouting bool, hostname string, informer interface{}) webhook.ServiceResolver {
	// if testServiceResolver != nil {
	// 	return testServiceResolver
	// }

	// var serviceResolver webhook.ServiceResolver
	// if enabledAggregatorRouting {
	// 	serviceResolver = aggregatorapiserver.NewEndpointServiceResolver(
	// 		informer.Core().V1().Services().Lister(),
	// 		informer.Core().V1().Endpoints().Lister(),
	// 	)
	// } else {
	// 	serviceResolver = aggregatorapiserver.NewClusterIPServiceResolver(
	// 		informer.Core().V1().Services().Lister(),
	// 	)
	// }

	// // resolve kubernetes.default.svc locally
	// if localHost, err := url.Parse(hostname); err == nil {
	// 	serviceResolver = aggregatorapiserver.NewLoopbackServiceResolver(serviceResolver, localHost)
	// }
	// return serviceResolver
	return nil
}
