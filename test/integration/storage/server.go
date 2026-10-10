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

//go:build integration

package storage

import (
	"context"
	"crypto/aes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"k8s.io/apiserver/pkg/features"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/seanchann/apimaster/pkg/api/legacyscheme"
	_ "github.com/seanchann/apimaster/pkg/apis/coreres/install"
	_ "github.com/seanchann/apimaster/pkg/apis/rbac/install"
	rbacv1 "github.com/seanchann/apimaster/pkg/apis/rbac/v1"
	"github.com/seanchann/apimaster/pkg/apiserver"
	"github.com/seanchann/apimaster/pkg/apiserver/options"
	clientset "github.com/seanchann/apimaster/pkg/client/clientset"
	"github.com/seanchann/apimaster/pkg/client/informers"
	generatedopenapi "github.com/seanchann/apimaster/pkg/generated/openapi"
	rbacrest "github.com/seanchann/apimaster/pkg/registry/rbac/rest"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	server "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
	"k8s.io/apiserver/pkg/storage/value"
	encryptaes "k8s.io/apiserver/pkg/storage/value/encrypt/aes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kube-openapi/pkg/common"
)

// Options uses an externally owned isolated test database, so restart can retain it.
// DSN is confidential and is never formatted or logged.
type Options struct {
	Backend string
	DSN     string
	Address string
	Cache   bool
}

// Server owns its listener, runtime, registry stores, and SQL health probes.
type Server struct {
	URL           string
	AdminConfig   *rest.Config
	StorageConfig storagebackend.Config
	Transformer   value.Transformer
	Generic       *server.GenericAPIServer
	Captured      map[schema.GroupResource]generic.RESTOptions
	Stores        map[schema.GroupResource]storage.Interface
	Examples      map[schema.GroupResource]runtime.Object
	ReverseKeys   map[schema.GroupResource]storage.ReverseKeyFunc
	activeWatches *atomic.Int64
	cancel        context.CancelFunc
	done          chan error
	once          sync.Once
	t             testing.TB
}

// Close synchronously drains the running server and destroys its owned storage.
func (s *Server) Close() {
	s.once.Do(func() {
		s.cancel()
		select {
		case err := <-s.done:
			require.NoError(s.t, err)
		case <-time.After(20 * time.Second):
			s.t.Error("AM server failed to stop within ownership deadline")
		}
	})
}

// Start runs the public AM option/configuration/registry assembly on a private TLS listener.
func Start(t testing.TB, opts Options) *Server {
	t.Helper()
	o := options.NewAPIMasterOptions(fixtureAdmission{}, options.StorageBackendType(opts.Backend))
	o.InsecureServing.BindPort = 0
	address := opts.Address
	if address == "" {
		address = "127.0.0.1:0"
	}
	listener, err := net.Listen("tcp", address)
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() })
	certificateFixture := httptest.NewTLSServer(http.NotFoundHandler())
	certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateFixture.Certificate().Raw})
	key, err := x509.MarshalPKCS8PrivateKey(certificateFixture.TLS.Certificates[0].PrivateKey)
	require.NoError(t, err)
	certificateFixture.Close()
	serving, err := dynamiccertificates.NewStaticCertKeyContent("task9-integration", certificate, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}))
	require.NoError(t, err)
	o.SecureServing.Listener = listener
	o.SecureServing.ServerCert.GeneratedCert = serving
	o.Authorization.Modes = []string{"AlwaysAllow"}
	cfg := storagebackend.NewDefaultConfig("/task9/application", nil)
	cfg.EventsHistoryWindow = 2 * time.Minute
	block, err := aes.NewCipher(make([]byte, 32))
	require.NoError(t, err)
	cfg.Transformer, err = encryptaes.NewGCMTransformer(block)
	require.NoError(t, err)
	switch opts.Backend {
	case "sqlite":
		cfg.Type = storagebackend.StorageTypeSqlite
		cfg.Sqlite.DSN = opts.DSN
		o.Sqlite = options.NewSqliteOptions(cfg)
		require.Empty(t, o.Sqlite.Validate())
	case "mysql":
		cfg.Type = storagebackend.StorageTypeMysql
		cfg.Mysql.ServerList = []string{opts.DSN}
		o.Mysql = options.NewMysqlOptions(cfg)
		require.Empty(t, o.Mysql.Validate())
	default:
		t.Fatalf("unsupported test backend %q", opts.Backend)
	}
	provider := fixtureProvider{}
	completed, err := apiserver.Complete(o, provider)
	require.NoError(t, err)
	configuration, _, kubeInformers, _, err := apiserver.BuildGenericConfig(completed, []*runtime.Scheme{legacyscheme.Scheme}, provider.GetOpenAPIDefinitions)
	require.NoError(t, err)
	t.Cleanup(configuration.Close)
	config := configuration.GenericConfig
	activeWatches := &atomic.Int64{}
	originalChain := config.BuildHandlerChainFunc
	config.BuildHandlerChainFunc = func(handler http.Handler, c *server.Config) http.Handler {
		actual := originalChain(handler, c)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("watch") == "true" {
				activeWatches.Add(1)
				defer activeWatches.Add(-1)
			}
			actual.ServeHTTP(w, r)
		})
	}
	t.Logf("backend=%s cacher=%t WatchList=%t ListFromCacheSnapshot=%t ConsistentListFromCacheSkipTimeoutFallback=%t", opts.Backend, opts.Cache, utilfeature.DefaultFeatureGate.Enabled(features.WatchList), utilfeature.DefaultFeatureGate.Enabled(features.ListFromCacheSnapshot), utilfeature.DefaultFeatureGate.Enabled(features.ConsistentListFromCacheSkipTimeoutFallback))

	// Explicit test-only administrator identity survives restarts; this fixture
	// exercises storage and HTTP errors, not production authentication policy.
	const adminToken = "task9-test-administrator"
	config.Authentication.Authenticator = authenticator.RequestFunc(func(r *http.Request) (*authenticator.Response, bool, error) {
		if r.Header.Get("Authorization") != "Bearer "+adminToken {
			return nil, false, nil
		}
		return &authenticator.Response{User: &user.DefaultInfo{Name: "task9-admin", Groups: []string{"system:masters", "system:authenticated"}}}, true, nil
	})
	config.RequestTimeout = 5 * time.Second
	config.ShutdownWatchTerminationGracePeriod = time.Second
	captured := &captureRESTOptions{RESTOptionsGetter: config.RESTOptionsGetter, cache: opts.Cache, options: map[schema.GroupResource]generic.RESTOptions{}, stores: map[schema.GroupResource]storage.Interface{}, examples: map[schema.GroupResource]runtime.Object{}, reverseKeys: map[schema.GroupResource]storage.ReverseKeyFunc{}}
	config.RESTOptionsGetter = captured
	built, err := apiserver.CreateAPIServer(configuration, server.NewEmptyDelegate(), kubeInformers)
	require.NoError(t, err)
	admin := rest.CopyConfig(config.LoopbackClientConfig)
	admin.Host = "https://" + listener.Addr().String()
	admin.BearerToken = adminToken
	admin.CAData = nil
	admin.CAFile = ""
	admin.Insecure = true
	admin.QPS = 10000
	admin.Burst = 2000
	admin.Timeout = 0
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{URL: admin.Host, AdminConfig: admin, StorageConfig: *cfg, Transformer: cfg.Transformer, Generic: built.GenericAPIServer, Captured: captured.options, Stores: captured.stores, Examples: captured.examples, ReverseKeys: captured.reverseKeys, activeWatches: activeWatches, cancel: cancel, done: make(chan error, 1), t: t}
	prepared := s.Generic.PrepareRun()
	go func() { s.done <- prepared.RunWithContext(ctx) }()
	t.Cleanup(s.Close)
	client, err := rest.HTTPClientFor(admin)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		req, _ := http.NewRequest("GET", s.URL+"/readyz", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}, 20*time.Second, 20*time.Millisecond)
	return s
}

type captureRESTOptions struct {
	generic.RESTOptionsGetter
	cache       bool
	options     map[schema.GroupResource]generic.RESTOptions
	stores      map[schema.GroupResource]storage.Interface
	examples    map[schema.GroupResource]runtime.Object
	reverseKeys map[schema.GroupResource]storage.ReverseKeyFunc
}

func (c *captureRESTOptions) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (generic.RESTOptions, error) {
	o, err := c.RESTOptionsGetter.GetRESTOptions(resource, example)
	if err != nil {
		return o, err
	}
	if example == nil {
		return o, fmt.Errorf("AM registry must supply a storage example")
	}
	if c.cache {
		o.Decorator = genericregistry.StorageWithCacher()
	}

	actual := o.Decorator
	o.Decorator = func(config *storagebackend.ConfigForResource, prefix string, key generic.CacheKeyFunc, reverse storage.ReverseKeyFunc, newFunc, newListFunc func() runtime.Object, attrs storage.AttrFunc, triggers storage.IndexerFuncs, indexers *cache.Indexers) (storage.Interface, factory.DestroyFunc, error) {
		result, destroy, err := actual(config, prefix, key, reverse, newFunc, newListFunc, attrs, triggers, indexers)
		if err == nil {
			c.stores[resource] = result
			c.reverseKeys[resource] = reverse
		}
		return result, destroy, err
	}
	c.examples[resource] = example
	c.options[resource] = o
	return o, nil
}
func (c *captureRESTOptions) Close() error {
	if closer, ok := c.RESTOptionsGetter.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

type fixtureAdmission struct{}

func (fixtureAdmission) RegisterAllAdmissionPlugins(*admission.Plugins) {}
func (fixtureAdmission) AllPluginOrder() []string {
	return genericoptions.NewAdmissionOptions().RecommendedPluginOrder
}
func (fixtureAdmission) DefaultOffAdmissionPlugins() sets.String { return sets.NewString() }

type fixtureProvider struct{}

func (fixtureProvider) APIName() string { return "task9-integration" }
func (fixtureProvider) Version() *version.Info {
	return &version.Info{Major: "1", Minor: "0", GitVersion: "v1.0.0"}
}
func (fixtureProvider) DefaultAPIResourceConfigSource() *serverstorage.ResourceConfig {
	c := serverstorage.NewResourceConfig()
	c.EnableVersions(rbacv1.SchemeGroupVersion)
	return c
}
func (fixtureProvider) DefaultInstallExtendRoutes(*restful.Container) {}
func (fixtureProvider) NewControllerProvider([]interface{}) (apiserver.ControllerProvider, error) {
	return fixtureController{}, nil
}
func (fixtureProvider) GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return generatedopenapi.GetOpenAPIDefinitions(ref)
}
func (fixtureProvider) ClientNewForConfig(c *rest.Config) (interface{}, error) {
	return clientset.NewForConfig(c)
}
func (fixtureProvider) ClientNewSharedInformerFactory(c interface{}, d time.Duration) interface{} {
	return informers.NewSharedInformerFactory(c.(clientset.Interface), d)
}

type fixtureController struct{}

func (fixtureController) Name() string { return "task9-fixture" }
func (fixtureController) PostFunc() server.PostStartHookFunc {
	return func(server.PostStartHookContext) error { return nil }
}
func (fixtureController) PreShutdownFunc() server.PreShutdownHookFunc {
	return func() error { return nil }
}
func (fixtureController) RESTStorageProviderBuilderHandle() apiserver.RESTStorageProviderBuilder {
	return fixtureBuilder{}
}

type fixtureBuilder struct{}

func (fixtureBuilder) NewProvider() []apiserver.RESTStorageProvider {
	return []apiserver.RESTStorageProvider{rbacrest.RESTStorageProvider{Authorizer: authorizerfactory.NewAlwaysAllowAuthorizer()}}
}
func (fixtureBuilder) BuildAPIResouceConfigSource() serverstorage.APIResourceConfigSource {
	return fixtureProvider{}.DefaultAPIResourceConfigSource()
}
