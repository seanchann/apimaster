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
	"encoding/json"
	"fmt"
	"io"
	"k8s.io/apiserver/pkg/features"
	"k8s.io/apiserver/pkg/storage/value"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	rbac "github.com/seanchann/apimaster/pkg/apis/rbac"
	rbacv1 "github.com/seanchann/apimaster/pkg/apis/rbac/v1"
	clientset "github.com/seanchann/apimaster/pkg/client/clientset"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/test/integration/storage/framework"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func TestBackendHTTPContracts(t *testing.T) {
	for _, backend := range []string{"sqlite", "mysql"} {
		for _, cached := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/cache=%t", backend, cached), func(t *testing.T) {
				db := framework.Open(t, backend)
				s := Start(t, Options{Backend: backend, DSN: db.DSN, Cache: cached})
				client, err := clientset.NewForConfig(s.AdminConfig)
				require.NoError(t, err)
				ctx := context.Background()
				roles := client.RbacV1().Roles("test-ns")
				input := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "sample", Labels: map[string]string{"group": "sample"}}}
				first, err := roles.Create(ctx, input, metav1.CreateOptions{})
				require.NoError(t, err)
				_, err = roles.Create(ctx, input, metav1.CreateOptions{})
				require.True(t, apierrors.IsAlreadyExists(err), "duplicate must map to AlreadyExists: %v", err)
				stale := first.DeepCopy()
				first.Labels["changed"] = "yes"
				updated, err := roles.Update(ctx, first, metav1.UpdateOptions{})
				require.NoError(t, err)
				require.NotEqual(t, stale.ResourceVersion, updated.ResourceVersion)
				_, err = roles.Update(ctx, stale, metav1.UpdateOptions{})
				require.True(t, apierrors.IsConflict(err), "stale update must map to Conflict: %v", err)
				require.NoError(t, roles.Delete(ctx, "sample", metav1.DeleteOptions{}))
				_, err = roles.Get(ctx, "sample", metav1.GetOptions{})
				require.True(t, apierrors.IsNotFound(err))
				for i := 0; i < 1025; i++ {
					_, err := roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("role-%04d", i), Labels: map[string]string{"last": fmt.Sprint(i == 1024)}}}, metav1.CreateOptions{})
					require.NoError(t, err)
				}
				_, err = client.RbacV1().Roles("other-ns").Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "role-0000"}}, metav1.CreateOptions{})
				require.NoError(t, err)
				all, err := roles.List(ctx, metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, all.Items, 1025)
				seen := map[string]bool{}
				continuation := ""
				revision := ""
				for {
					page, err := roles.List(ctx, metav1.ListOptions{Limit: 128, Continue: continuation})
					require.NoError(t, err)
					if revision == "" {
						revision = page.ResourceVersion
					} else {
						require.Equal(t, revision, page.ResourceVersion)
					}
					for _, obj := range page.Items {
						require.False(t, seen[obj.Name])
						seen[obj.Name] = true
					}
					continuation = page.Continue
					if continuation == "" {
						break
					}
				}
				require.Len(t, seen, 1025)
				filtered, err := roles.List(ctx, metav1.ListOptions{Limit: 1, LabelSelector: "last=true"})
				require.NoError(t, err)
				require.Len(t, filtered.Items, 1)
				require.Equal(t, "role-1024", filtered.Items[0].Name)
				exact, err := roles.List(ctx, metav1.ListOptions{FieldSelector: "metadata.name=role-1024"})
				require.NoError(t, err)
				require.Len(t, exact.Items, 1)
				other, err := client.RbacV1().Roles("other-ns").List(ctx, metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, other.Items, 1)
				page, err := roles.List(ctx, metav1.ListOptions{Limit: 1})
				require.NoError(t, err)
				_, err = client.RbacV1().Roles("other-ns").List(ctx, metav1.ListOptions{Limit: 1, Continue: page.Continue})
				require.True(t, apierrors.IsBadRequest(err), "HTTP cross-namespace continuation must fail: %v", err)
			})
		}
	}
}

// watchConnections closes real HTTP response streams; it never injects events.
type watchConnections struct {
	base     http.RoundTripper
	mu       sync.Mutex
	bodies   []io.ReadCloser
	requests atomic.Int64
}

func (w *watchConnections) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := w.base.RoundTrip(req)
	if err == nil && req.URL.Query().Get("watch") == "true" {
		w.mu.Lock()
		w.bodies = append(w.bodies, res.Body)
		w.mu.Unlock()
		w.requests.Add(1)
	}
	return res, err
}
func (w *watchConnections) disconnect() {
	w.mu.Lock()
	bodies := w.bodies
	w.bodies = nil
	w.mu.Unlock()
	for _, body := range bodies {
		body.Close()
	}
}

func TestBackendInformerRecovery(t *testing.T) {
	for _, backend := range []string{"sqlite", "mysql"} {
		for _, cached := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/cache=%t", backend, cached), func(t *testing.T) {
				db := framework.Open(t, backend)
				opts := Options{Backend: backend, DSN: db.DSN, Cache: cached}
				s := Start(t, opts)
				client, err := clientset.NewForConfig(s.AdminConfig)
				require.NoError(t, err)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				roles := client.RbacV1().Roles("recovery")
				_, err = roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "initial"}}, metav1.CreateOptions{})
				require.NoError(t, err)
				connections := &watchConnections{}
				config := rest.CopyConfig(s.AdminConfig)
				config.WrapTransport = func(base http.RoundTripper) http.RoundTripper { connections.base = base; return connections }
				observedClient, err := clientset.NewForConfig(config)
				require.NoError(t, err)
				observedRoles := observedClient.RbacV1().Roles("recovery")
				informer := cache.NewSharedIndexInformer(&cache.ListWatch{ListFunc: func(o metav1.ListOptions) (runtime.Object, error) { return observedRoles.List(ctx, o) }, WatchFunc: func(o metav1.ListOptions) (watch.Interface, error) { return observedRoles.Watch(ctx, o) }}, &rbacv1.Role{}, 0, cache.Indexers{})
				stopped := make(chan struct{})
				go func() { defer close(stopped); informer.Run(ctx.Done()) }()
				t.Cleanup(func() {
					cancel()
					connections.disconnect()
					select {
					case <-stopped:
					case <-time.After(10 * time.Second):
						t.Error("informer did not release its worker")
					}
				})
				require.True(t, cache.WaitForCacheSync(ctx.Done(), informer.HasSynced))
				await := func(name, label string, exists bool) {
					t.Helper()
					require.Eventually(t, func() bool {
						o, ok, err := informer.GetStore().GetByKey("recovery/" + name)
						if err != nil || ok != exists {
							return false
						}
						return !ok || o.(*rbacv1.Role).Labels["state"] == label
					}, 10*time.Second, 10*time.Millisecond)
				}
				await("initial", "", true)
				added, err := roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "live", Labels: map[string]string{"state": "created"}}}, metav1.CreateOptions{})
				require.NoError(t, err)
				await("live", "created", true)
				added.Labels["state"] = "updated"
				_, err = roles.Update(ctx, added, metav1.UpdateOptions{})
				require.NoError(t, err)
				await("live", "updated", true)
				require.NoError(t, roles.Delete(ctx, "live", metav1.DeleteOptions{}))
				await("live", "", false)
				require.Eventually(t, func() bool { return connections.requests.Load() > 0 }, 5*time.Second, 10*time.Millisecond)
				before := connections.requests.Load()
				connections.disconnect()
				_, err = roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "disconnected"}}, metav1.CreateOptions{})
				require.NoError(t, err)
				await("disconnected", "", true)
				require.Eventually(t, func() bool { return connections.requests.Load() > before }, 10*time.Second, 10*time.Millisecond)
				address, err := url.Parse(s.URL)
				require.NoError(t, err)
				opts.Address = address.Host
				roleOptions := s.Captured[rbac.Resource("roles")]
				s.Close()
				// Change committed state while the HTTP server is absent. Restart must recover
				// from SQL history rather than a process-local notification or stale cache.
				raw, destroy, err := generic.NewRawStorage(roleOptions.StorageConfig, func() runtime.Object { return &rbac.Role{} }, func() runtime.Object { return &rbac.RoleList{} }, nil, "/"+roleOptions.ResourcePrefix)
				require.NoError(t, err)
				require.NoError(t, raw.Create(ctx, "/"+roleOptions.ResourcePrefix+"/recovery/offline", &rbac.Role{ObjectMeta: metav1.ObjectMeta{Name: "offline", Namespace: "recovery"}}, nil, 0))
				require.NoError(t, raw.Delete(ctx, "/"+roleOptions.ResourcePrefix+"/recovery/initial", &rbac.Role{}, nil, storage.ValidateAllObjectFunc, nil, storage.DeleteOptions{}))
				destroy()
				s = Start(t, opts)
				await("offline", "", true)
				await("initial", "", false)
				await("disconnected", "", true)
				require.Len(t, informer.GetStore().List(), 2)
				cancel()
				connections.disconnect()
				select {
				case <-stopped:
				case <-time.After(10 * time.Second):
					t.Fatal("informer shutdown timed out")
				}
			})
		}
	}
}

func TestBackendAssemblyOwnership(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.AllowUnsafeMalformedObjectDeletion, true)
	for _, backend := range []string{"sqlite", "mysql"} {
		t.Run(backend, func(t *testing.T) {
			db := framework.Open(t, backend)
			s := Start(t, Options{Backend: backend, DSN: db.DSN})
			ctx := t.Context()
			client, err := clientset.NewForConfig(s.AdminConfig)
			require.NoError(t, err)
			created, err := client.RbacV1().Roles("assembly").Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "encrypted", Labels: map[string]string{"confidential": "fixture-value"}}}, metav1.CreateOptions{})
			require.NoError(t, err)
			resource := rbac.Resource("roles")
			captured := s.Captured[resource]
			require.IsType(t, &rbac.Role{}, s.Examples[resource])
			require.NotNil(t, captured.StorageConfig.Codec)
			require.NotNil(t, captured.StorageConfig.EncodeVersioner)
			require.Same(t, s.Transformer, captured.StorageConfig.Transformer)
			require.Equal(t, 2*time.Minute, captured.StorageConfig.EventsHistoryWindow)
			logicalKey := "/" + captured.ResourcePrefix + "/assembly/encrypted"
			fullKey := "/task9/application" + logicalKey
			var storedKey, prefix []byte
			var ciphertext []byte
			require.NoError(t, db.DB.QueryRow("SELECT storage_key,resource_prefix,object FROM storage_objects WHERE storage_key=?", []byte(fullKey)).Scan(&storedKey, &prefix, &ciphertext))
			require.Equal(t, fullKey, string(storedKey))
			require.Equal(t, "/task9/application/"+captured.ResourcePrefix, string(prefix))
			require.NotContains(t, string(ciphertext), "fixture-value")
			plaintext, _, err := s.Transformer.TransformFromStorage(ctx, ciphertext, value.DefaultContext(storedKey))
			require.NoError(t, err)
			require.Contains(t, string(plaintext), `"apiVersion":"rbac/v1"`)
			require.Contains(t, string(plaintext), `"confidential":"fixture-value"`)
			_, _, err = s.Transformer.TransformFromStorage(ctx, ciphertext, value.DefaultContext([]byte(logicalKey)))
			require.Error(t, err)
			var historyWindow int64
			require.NoError(t, db.DB.QueryRow("SELECT history_window FROM storage_policy WHERE id=1").Scan(&historyWindow))
			require.Equal(t, int64(2*time.Minute), historyWindow)
			require.NotNil(t, s.ReverseKeys[resource])
			name, namespace, err := s.ReverseKeys[resource](logicalKey)
			require.NoError(t, err)
			require.Equal(t, "encrypted", name)
			require.Equal(t, "assembly", namespace)
			raw := s.Stores[resource]
			require.NoError(t, raw.ReadinessCheck())
			// Prove that the actual AM registry reverse-key callback reaches SQL's
			// malformed-delete path, rather than only invoking the callback in isolation.
			_, err = db.DB.Exec("UPDATE storage_objects SET object=? WHERE storage_key=?", []byte("malformed ciphertext"), []byte(fullKey))
			require.NoError(t, err)
			deleted := &rbac.Role{}
			require.NoError(t, raw.Delete(ctx, logicalKey, deleted, nil, storage.ValidateAllObjectFunc, nil, storage.DeleteOptions{ExpectTransformOrDecodeError: true}))
			require.Empty(t, deleted.Name, "unsafe deletion preserves the caller output as required upstream")
			var identityCiphertext []byte
			require.NoError(t, db.DB.QueryRow("SELECT previous_object FROM storage_history WHERE storage_key=? AND change_type='DELETED'", []byte(fullKey)).Scan(&identityCiphertext))
			identityPlain, _, err := s.Transformer.TransformFromStorage(ctx, identityCiphertext, value.DefaultContext([]byte(fullKey)))
			require.NoError(t, err)
			identity := &rbac.Role{}
			require.NoError(t, runtime.DecodeInto(captured.StorageConfig.Codec, identityPlain, identity))
			require.Equal(t, created.Name, identity.Name)
			require.Equal(t, created.Namespace, identity.Namespace)
			require.Empty(t, identity.UID)
			httpClient, err := rest.HTTPClientFor(s.AdminConfig)
			require.NoError(t, err)
			for _, path := range []string{"/healthz", "/readyz"} {
				req, _ := http.NewRequestWithContext(ctx, "GET", s.URL+path, nil)
				req.Header.Set("Authorization", "Bearer "+s.AdminConfig.BearerToken)
				response, err := httpClient.Do(req)
				require.NoError(t, err)
				response.Body.Close()
				require.Equal(t, 200, response.StatusCode)
			}
			if backend == "mysql" {
				var connections int
				require.NoError(t, db.DB.QueryRow("SELECT COUNT(*) FROM information_schema.PROCESSLIST WHERE DB=DATABASE() AND ID<>CONNECTION_ID()").Scan(&connections))
				require.Positive(t, connections, "real AM stores and probe pools must be acquired")
			}
			databasePath := strings.SplitN(db.DSN, "?", 2)[0]
			if backend == "sqlite" {
				require.Positive(t, sqliteDescriptors(t, databasePath))
			}
			s.Close()
			require.Error(t, raw.ReadinessCheck())
			if backend == "mysql" {
				require.Eventually(t, func() bool {
					var connections int
					err := db.DB.QueryRow("SELECT COUNT(*) FROM information_schema.PROCESSLIST WHERE DB=DATABASE() AND ID<>CONNECTION_ID()").Scan(&connections)
					return err == nil && connections == 0
				}, 5*time.Second, 10*time.Millisecond, "AM shutdown must close all store and SQL probe connections")
			}
			require.NoError(t, db.Close())
			if backend == "sqlite" {
				require.Zero(t, sqliteDescriptors(t, databasePath), "AM shutdown must release native SQLite store and probe descriptors")
			}
		})
	}
}
func sqliteDescriptors(t *testing.T, path string) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	require.NoError(t, err)
	count := 0
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err == nil && (target == path || target == path+"-wal" || target == path+"-shm") {
			count++
		}
	}
	return count
}

func TestBackendHTTPShutdown(t *testing.T) {
	for _, backend := range []string{"sqlite", "mysql"} {
		for _, cached := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/cache=%t", backend, cached), func(t *testing.T) {
				db := framework.Open(t, backend)
				s := Start(t, Options{Backend: backend, DSN: db.DSN, Cache: cached})
				client, err := clientset.NewForConfig(s.AdminConfig)
				require.NoError(t, err)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				roles := client.RbacV1().Roles("shutdown")
				current, err := roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "large", Annotations: map[string]string{"payload": strings.Repeat("x", 32*1024)}}}, metav1.CreateOptions{})
				require.NoError(t, err)
				httpClient, err := rest.HTTPClientFor(s.AdminConfig)
				require.NoError(t, err)
				openStream := func(parent context.Context) *http.Response {
					req, _ := http.NewRequestWithContext(parent, "GET", s.URL+"/apis/rbac/v1/namespaces/shutdown/roles?watch=true&resourceVersion=0", nil)
					req.Header.Set("Authorization", "Bearer "+s.AdminConfig.BearerToken)
					response, err := httpClient.Do(req)
					require.NoError(t, err)
					require.Equal(t, 200, response.StatusCode)
					require.Eventually(t, func() bool { return s.activeWatches.Load() == 1 }, 5*time.Second, 10*time.Millisecond)
					return response
				}
				streamContext, stopStream := context.WithCancel(ctx)
				response := openStream(streamContext)
				stopStream()
				response.Body.Close()
				require.Eventually(t, func() bool { return s.activeWatches.Load() == 0 }, 5*time.Second, 10*time.Millisecond, "canceled HTTP stream must release its handler")
				response = openStream(ctx)
				response.Body.Close()
				require.Eventually(t, func() bool { return s.activeWatches.Load() == 0 }, 5*time.Second, 10*time.Millisecond, "disconnected HTTP client must release its handler")
				response = openStream(ctx)
				// Leave the response body unread while committing over nine MiB of events.
				// SQL result-channel overflow is tested independently; here the observable
				// contract is bounded writes and released HTTP/storage ownership on stop.
				for i := 0; i < 300; i++ {
					current.Annotations["sequence"] = fmt.Sprint(i)
					current, err = roles.Update(ctx, current, metav1.UpdateOptions{})
					require.NoError(t, err)
				}
				require.NoError(t, ctx.Err(), "slow HTTP reader must not block database writers")
				s.Close()
				require.Zero(t, s.activeWatches.Load(), "server stop must join slow HTTP handlers")
				response.Body.Close()
				_, err = roles.Get(ctx, "large", metav1.GetOptions{})
				require.Error(t, err, "listener must reject requests after stop")
			})
		}
	}
}

func TestBackendHTTPDatabaseFault(t *testing.T) {
	for _, backend := range []string{"sqlite", "mysql"} {
		for _, cached := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/cache=%t", backend, cached), func(t *testing.T) {
				db := framework.Open(t, backend)
				s := Start(t, Options{Backend: backend, DSN: db.DSN, Cache: cached})
				client, err := clientset.NewForConfig(s.AdminConfig)
				require.NoError(t, err)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				roles := client.RbacV1().Roles("fault")
				created, err := roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "before"}}, metav1.CreateOptions{})
				require.NoError(t, err)
				w, err := roles.Watch(ctx, metav1.ListOptions{ResourceVersion: created.ResourceVersion})
				require.NoError(t, err)
				defer w.Stop()
				require.Eventually(t, func() bool { return s.activeWatches.Load() == 1 }, 5*time.Second, 10*time.Millisecond)
				_, err = db.DB.Exec("ALTER TABLE storage_history RENAME TO task9_history_offline")
				require.NoError(t, err)
				restored := false
				defer func() {
					if !restored {
						_, err := db.DB.Exec("ALTER TABLE task9_history_offline RENAME TO storage_history")
						require.NoError(t, err)
					}
				}()
				_, err = roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "must-rollback"}}, metav1.CreateOptions{})
				require.Error(t, err)
				require.True(t, apierrors.IsInternalError(err), "database fault must map to HTTP InternalError: %v", err)
				if !cached {
					select {
					case event, ok := <-w.ResultChan():
						require.True(t, ok, "raw SQL fault must carry an error event")
						require.Equal(t, watch.Error, event.Type)
					case <-time.After(10 * time.Second):
						t.Fatal("database fault did not terminate raw HTTP watch")
					}
					w.Stop()
					require.Eventually(t, func() bool { return s.activeWatches.Load() == 0 }, 5*time.Second, 10*time.Millisecond)
				}
				// The upstream Cacher retains client streams while its reflector
				// reconnects from the durable cursor. Verify recovery below, rather
				// than imposing the raw store's terminal-error behavior on it.
				_, err = db.DB.Exec("ALTER TABLE task9_history_offline RENAME TO storage_history")
				require.NoError(t, err)
				restored = true
				require.Eventually(t, func() bool { return s.Stores[rbac.Resource("roles")].ReadinessCheck() == nil }, 10*time.Second, 20*time.Millisecond)
				_, err = roles.Get(ctx, "must-rollback", metav1.GetOptions{})
				require.True(t, apierrors.IsNotFound(err), "failed transaction must leave no object")
				if !cached {
					w, err = roles.Watch(ctx, metav1.ListOptions{ResourceVersion: created.ResourceVersion})
					require.NoError(t, err)
					defer w.Stop()
				}
				_, err = roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "after-recovery"}}, metav1.CreateOptions{})
				require.NoError(t, err)
				select {
				case event, ok := <-w.ResultChan():
					require.True(t, ok)
					require.Equal(t, watch.Added, event.Type)
					require.Equal(t, "after-recovery", event.Object.(*rbacv1.Role).Name)
				case <-time.After(10 * time.Second):
					t.Fatal("HTTP watch failed to resume after database recovery")
				}
				w.Stop()
				require.Eventually(t, func() bool { return s.activeWatches.Load() == 0 }, 5*time.Second, 10*time.Millisecond)
				s.Close()
				require.Zero(t, s.activeWatches.Load())
			})
		}
	}
}

// Managed-field assertions exercise real native typed conversion and persisted ownership.
func TestBackendHTTPManagedFields(t *testing.T) {
	for _, backend := range []string{"sqlite", "mysql"} {
		for _, cached := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/cache=%t", backend, cached), func(t *testing.T) {
				db := framework.Open(t, backend)
				s := Start(t, Options{Backend: backend, DSN: db.DSN, Cache: cached})
				client, err := clientset.NewForConfig(s.AdminConfig)
				require.NoError(t, err)
				roles := client.RbacV1().Roles("ownership")
				ctx := t.Context()
				t.Run("create-update", func(t *testing.T) {
					created, err := roles.Create(ctx, &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "tracked", Labels: map[string]string{"created": "first"}}}, metav1.CreateOptions{FieldManager: "create-owner"})
					require.NoError(t, err)
					requireOwnedLabel(t, created, "create-owner", metav1.ManagedFieldsOperationUpdate, "created")
					created.Labels["created"] = "changed"
					created.Labels["updated"] = "second"
					updated, err := roles.Update(ctx, created, metav1.UpdateOptions{FieldManager: "update-owner"})
					require.NoError(t, err)
					requireOwnedLabel(t, updated, "update-owner", metav1.ManagedFieldsOperationUpdate, "created")
					requireOwnedLabel(t, updated, "update-owner", metav1.ManagedFieldsOperationUpdate, "updated")
					for _, entry := range updated.ManagedFields {
						if entry.Manager == "create-owner" && entry.FieldsV1 != nil {
							require.NotContains(t, string(entry.FieldsV1.Raw), `"f:created"`, "update must remove previous field ownership")
						}
					}
					stored, err := roles.Get(ctx, "tracked", metav1.GetOptions{})
					require.NoError(t, err)
					require.Equal(t, updated.ManagedFields, stored.ManagedFields)
					require.Equal(t, "changed", stored.Labels["created"])
				})
				t.Run("apply-update-conflict-transfer", func(t *testing.T) {
					apply := func(value, manager string, force *bool) (*rbacv1.Role, error) {
						payload, err := json.Marshal(map[string]interface{}{"apiVersion": rbacv1.SchemeGroupVersion.String(), "kind": "Role", "metadata": map[string]interface{}{"name": "applied", "namespace": "ownership", "labels": map[string]string{"applied": value}}})
						require.NoError(t, err)
						return roles.Patch(ctx, "applied", types.ApplyPatchType, payload, metav1.PatchOptions{FieldManager: manager, Force: force})
					}
					first, err := apply("first", "apply-owner", nil)
					require.NoError(t, err, "actual server-side apply must use the registered Role schema")
					requireOwnedLabel(t, first, "apply-owner", metav1.ManagedFieldsOperationApply, "applied")
					first.Labels["manual"] = "kept"
					updated, err := roles.Update(ctx, first, metav1.UpdateOptions{FieldManager: "manual-owner"})
					require.NoError(t, err)
					requireOwnedLabel(t, updated, "apply-owner", metav1.ManagedFieldsOperationApply, "applied")
					requireOwnedLabel(t, updated, "manual-owner", metav1.ManagedFieldsOperationUpdate, "manual")
					_, err = apply("second", "other-owner", nil)
					require.True(t, apierrors.IsConflict(err), "another manager must not overwrite owned fields: %v", err)
					force := true
					transferred, err := apply("second", "other-owner", &force)
					require.NoError(t, err)
					requireOwnedLabel(t, transferred, "other-owner", metav1.ManagedFieldsOperationApply, "applied")
					requireOwnedLabel(t, transferred, "manual-owner", metav1.ManagedFieldsOperationUpdate, "manual")
					for _, entry := range transferred.ManagedFields {
						if entry.Manager == "apply-owner" && entry.FieldsV1 != nil {
							require.NotContains(t, string(entry.FieldsV1.Raw), `"f:applied"`, "forced apply must transfer previous ownership")
						}
					}
					stored, err := roles.Get(ctx, "applied", metav1.GetOptions{})
					require.NoError(t, err)
					require.Equal(t, "second", stored.Labels["applied"])
					require.Equal(t, "kept", stored.Labels["manual"])
					require.Equal(t, transferred.ManagedFields, stored.ManagedFields)
				})
			})
		}
	}
}

func requireOwnedLabel(t *testing.T, role *rbacv1.Role, manager string, operation metav1.ManagedFieldsOperationType, label string) {
	t.Helper()
	for _, entry := range role.ManagedFields {
		if entry.Manager != manager || entry.Operation != operation {
			continue
		}
		require.Equal(t, rbacv1.SchemeGroupVersion.String(), entry.APIVersion)
		require.NotNil(t, entry.FieldsV1)
		var fields map[string]interface{}
		require.NoError(t, json.Unmarshal(entry.FieldsV1.Raw, &fields))
		for _, key := range []string{"f:metadata", "f:labels", "f:" + label} {
			require.Contains(t, fields, key)
			child, ok := fields[key].(map[string]interface{})
			require.True(t, ok)
			fields = child
		}
		return
	}
	t.Fatalf("no %s managed-field entry for %s; actual=%+v", operation, manager, role.ManagedFields)
}
