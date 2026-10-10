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
	"encoding/json"
	"github.com/seanchann/apimaster/pkg/apiserver/options"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStorageBackendConfiguration(t *testing.T) {
	cases := []struct {
		name      string
		backend   options.StorageBackendType
		configure func(*options.APIMasterOptions)
		valid     bool
	}{
		{"sqlite valid", options.StorageBackendTypeSqlite, func(o *options.APIMasterOptions) { o.Sqlite.StorageConfig.Sqlite.DSN = "file:/tmp/task8-options.db" }, true},
		{"sqlite missing", options.StorageBackendTypeSqlite, func(o *options.APIMasterOptions) {}, false},
		{"sqlite missing options", options.StorageBackendTypeSqlite, func(o *options.APIMasterOptions) { o.Sqlite = nil }, false},
		{"mysql valid", options.StorageBackendTypeMysql, func(o *options.APIMasterOptions) {
			o.Mysql.StorageConfig.Mysql.ServerList = []string{"user@tcp(localhost:3306)/db"}
		}, true},
		{"mysql missing", options.StorageBackendTypeMysql, func(o *options.APIMasterOptions) {}, false},
		{"mysql multiple", options.StorageBackendTypeMysql, func(o *options.APIMasterOptions) {
			o.Mysql.StorageConfig.Mysql.ServerList = []string{"user@tcp(localhost:3306)/db", "user@tcp(localhost:3306)/db2"}
		}, false},
		{"mysql no database", options.StorageBackendTypeMysql, func(o *options.APIMasterOptions) {
			o.Mysql.StorageConfig.Mysql.ServerList = []string{"user@tcp(localhost:3306)/"}
		}, false},
		{"etcd valid", options.StorageBackendTypeEtcd, func(o *options.APIMasterOptions) {
			o.Etcd.StorageConfig.Transport.ServerList = []string{"http://127.0.0.1:2379"}
		}, true},
		{"etcd missing", options.StorageBackendTypeEtcd, func(o *options.APIMasterOptions) {}, false},
		{"unknown", options.StorageBackendType("unknown"), func(o *options.APIMasterOptions) {}, false},
		{"selected backend only", options.StorageBackendTypeSqlite, func(o *options.APIMasterOptions) {
			o.Sqlite.StorageConfig.Sqlite.DSN = "file:/tmp/task8-options.db"
			o.Mysql = options.NewMysqlOptions(&o.Sqlite.StorageConfig)
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := options.NewAPIMasterOptions(donorAdmission{}, tc.backend)
			tc.configure(o)
			o.Authorization.Complete()
			if tc.valid {
				require.Empty(t, o.Validate())
			} else {
				require.NotEmpty(t, o.Validate())
			}
		})
	}
}

func TestApplicationFeatureOptions(t *testing.T) {
	o := options.NewAPIMasterOptions(donorAdmission{}, options.StorageBackendTypeSqlite)
	require.False(t, o.Features.EnablePriorityAndFairness, "standalone API has no flowcontrol API")
	config := server.NewConfig(serializer.CodecFactory{})
	require.NoError(t, o.ApplyFeaturesTo(config))
	require.Nil(t, config.FlowControl)
	o.Features.EnablePriorityAndFairness = true
	require.Error(t, o.ApplyFeaturesTo(config))
}

func TestApplicationFeatureOptionsEnabled(t *testing.T) {
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/apis/flowcontrol.apiserver.k8s.io/v1":
			json.NewEncoder(w).Encode(&metav1.APIResourceList{GroupVersion: "flowcontrol.apiserver.k8s.io/v1", APIResources: []metav1.APIResource{{Name: "flowschemas", Kind: "FlowSchema", Verbs: metav1.Verbs{"get", "list", "watch"}}, {Name: "prioritylevelconfigurations", Kind: "PriorityLevelConfiguration", Verbs: metav1.Verbs{"get", "list", "watch"}}}})
		case "/apis/flowcontrol.apiserver.k8s.io/v1/flowschemas":
			w.Write([]byte(`{"apiVersion":"flowcontrol.apiserver.k8s.io/v1","kind":"FlowSchemaList","metadata":{"resourceVersion":"1"},"items":[]}`))
		case "/apis/flowcontrol.apiserver.k8s.io/v1/prioritylevelconfigurations":
			w.Write([]byte(`{"apiVersion":"flowcontrol.apiserver.k8s.io/v1","kind":"PriorityLevelConfigurationList","metadata":{"resourceVersion":"1"},"items":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer endpoint.Close()
	client, err := kubernetes.NewForConfig(&rest.Config{Host: endpoint.URL, Timeout: time.Second})
	require.NoError(t, err)
	o := options.NewAPIMasterOptions(donorAdmission{}, options.StorageBackendTypeSqlite)
	o.Features.EnablePriorityAndFairness = true
	o.KubernetesClient = client
	config := server.NewConfig(serializer.CodecFactory{})
	require.Error(t, o.ApplyFeaturesTo(config), "an informer factory is required")
	o.KubernetesInformers = informers.NewSharedInformerFactory(client, time.Minute)
	require.NoError(t, o.ApplyFeaturesTo(config))
	require.NotNil(t, config.FlowControl)
	require.NoError(t, o.GenericServerRunOptions.ApplyTo(config))
	config.ExternalAddress = "127.0.0.1:1"
	completed := config.Complete(o.KubernetesInformers)
	require.Same(t, o.KubernetesInformers, completed.SharedInformerFactory)
	_, err = client.FlowcontrolV1().FlowSchemas().List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	_, err = client.FlowcontrolV1().PriorityLevelConfigurations().List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
}
