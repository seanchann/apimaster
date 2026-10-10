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
	"context"
	"crypto/aes"
	"database/sql"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	coreres "github.com/seanchann/apimaster/pkg/apis/coreres"
	coreresv1 "github.com/seanchann/apimaster/pkg/apis/coreres/v1"
	"github.com/seanchann/apimaster/pkg/apiserver/options"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/value"
	encryptaes "k8s.io/apiserver/pkg/storage/value/encrypt/aes"
)

type exampleFactory struct {
	serverstorage.StorageFactory
	example runtime.Object
}

func (f *exampleFactory) NewConfig(resource schema.GroupResource, example runtime.Object) (*storagebackend.ConfigForResource, error) {
	f.example = example
	return f.StorageFactory.NewConfig(resource, example)
}

func TestSQLRESTOptionsGetter(t *testing.T) {
	temporaryParent := os.Getenv("APIMASTER_TEST_TMPDIR")
	if temporaryParent == "" {
		temporaryParent = "/tmp/easehold-upstream-storage"
	}
	require.NoError(t, os.MkdirAll(temporaryParent, 0755))
	directory, err := os.MkdirTemp(temporaryParent, "am-sql-options-")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(directory)) })
	scheme := runtime.NewScheme()
	require.NoError(t, coreres.AddToScheme(scheme))
	require.NoError(t, coreresv1.AddToScheme(scheme))
	codecs := serializer.NewCodecFactory(scheme)
	config := storagebackend.NewDefaultConfig("/task8/app", nil)
	config.Type = storagebackend.StorageTypeSqlite
	config.Sqlite.DSN = filepath.Join(directory, "storage.db")
	block, err := aes.NewCipher(make([]byte, 32))
	require.NoError(t, err)
	config.Transformer, err = encryptaes.NewGCMTransformer(block)
	require.NoError(t, err)
	config.EventsHistoryWindow = 2 * time.Minute
	factory := &exampleFactory{StorageFactory: serverstorage.NewDefaultStorageFactory(*config, "application/json", codecs, serverstorage.NewDefaultResourceEncodingConfig(scheme), serverstorage.NewResourceConfig(), nil)}
	genericConfig := server.NewConfig(codecs)
	opts := options.NewSqliteOptions(config)
	require.NoError(t, opts.ApplyWithStorageFactoryTo(factory, genericConfig))
	if closer, ok := genericConfig.RESTOptionsGetter.(interface{ Close() error }); ok {
		t.Cleanup(func() { require.NoError(t, closer.Close()) })
	}
	example := &coreres.Namespace{}
	resource := coreres.Resource("namespaces")
	restOptions, err := genericConfig.RESTOptionsGetter.GetRESTOptions(resource, example)
	require.NoError(t, err)
	require.Same(t, example, factory.example)
	require.Equal(t, resource, restOptions.StorageConfig.GroupResource)
	require.NotNil(t, restOptions.StorageConfig.Codec)
	require.NotNil(t, restOptions.StorageConfig.EncodeVersioner)
	require.Same(t, config.Transformer, restOptions.StorageConfig.Transformer)
	require.Equal(t, config.EventsHistoryWindow, restOptions.StorageConfig.EventsHistoryWindow)
	require.Equal(t, factory.ResourcePrefix(resource), restOptions.ResourcePrefix)
	require.Equal(t, "namespaces", restOptions.ResourcePrefix)
	require.Equal(t, "/task8/app", restOptions.StorageConfig.Prefix)
	backend, destroy, err := generic.NewRawStorage(restOptions.StorageConfig, func() runtime.Object { return &coreres.Namespace{} }, func() runtime.Object { return &coreres.NamespaceList{} }, func(string) (string, string, error) { return "", "sample", nil }, "/"+restOptions.ResourcePrefix)
	require.NoError(t, err)
	t.Cleanup(destroy)
	key := "/" + restOptions.ResourcePrefix + "/sample"
	input := &coreres.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "sample"}}
	output := &coreres.Namespace{}
	require.NoError(t, backend.Create(context.Background(), key, input, output, 0))
	require.NotEmpty(t, output.ResourceVersion)
	require.NoError(t, backend.Get(context.Background(), key, storage.GetOptions{}, &coreres.Namespace{}))
	observer, err := sql.Open("sqlite3", config.Sqlite.DSN)
	require.NoError(t, err)
	defer observer.Close()
	var storedKey string
	var payload []byte
	require.NoError(t, observer.QueryRow("SELECT storage_key,object FROM storage_objects").Scan(&storedKey, &payload))
	require.Equal(t, "/task8/app"+key, storedKey)
	plaintext, _, err := config.Transformer.TransformFromStorage(context.Background(), payload, value.DefaultContext([]byte(storedKey)))
	require.NoError(t, err)
	require.Contains(t, string(plaintext), `"apiVersion":"coreres/v1"`)
	require.Contains(t, string(plaintext), `"name":"sample"`)
	_, _, err = config.Transformer.TransformFromStorage(context.Background(), payload, value.DefaultContext([]byte(key)))
	require.Error(t, err)
	var historyWindow int64
	require.NoError(t, observer.QueryRow("SELECT history_window FROM storage_policy WHERE id=1").Scan(&historyWindow))
	require.Equal(t, int64(2*time.Minute), historyWindow)
	request := httptest.NewRequest("GET", "/healthz", nil)
	hasHealth, hasReady := false, false
	for _, check := range genericConfig.HealthzChecks {
		if check.Name() == "sql-storage" {
			hasHealth = true
			require.NoError(t, check.Check(request))
		}
	}
	for _, check := range genericConfig.ReadyzChecks {
		if check.Name() == "sql-storage-readiness" {
			hasReady = true
			require.NoError(t, check.Check(request))
		}
	}
	require.True(t, hasHealth, "SQL health check must be registered")
	require.True(t, hasReady, "SQL readiness check must be registered")
}
