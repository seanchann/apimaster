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
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/seanchann/apimaster/pkg/api/legacyscheme"
	"github.com/seanchann/apimaster/pkg/apiserver"
	"github.com/seanchann/apimaster/pkg/apiserver/options"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	server "k8s.io/apiserver/pkg/server"
	"k8s.io/kube-openapi/pkg/common"
)

func TestAuthenticationAudienceForwarding(t *testing.T) {
	o := options.NewBuiltInAuthenticationOptions().WithAnonymous()
	o.APIAudiences = []string{"explicit-consumer-audience"}
	converted, err := o.ToAuthenticationConfig()
	require.NoError(t, err)
	require.Nil(t, converted.APIAudiences, "donor conversion does not implement service-account audience population")
	info := &server.AuthenticationInfo{}
	require.NoError(t, o.ApplyTo(info, nil, nil, &common.Config{}, nil))
	require.Equal(t, authenticator.Audiences(o.APIAudiences), info.APIAudiences)
	response, ok, err := info.Authenticator.AuthenticateRequest(httptest.NewRequest("GET", "/healthz", nil))
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "system:anonymous", response.User.GetName())
}

func TestConfigurationFailureClosesListener(t *testing.T) {
	o := options.NewAPIMasterOptions(donorAdmission{}, options.StorageBackendTypeSqlite)
	o.InsecureServing.BindPort = 0
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	o.SecureServing.Listener = listener
	o.SecureServing.ServerCert.CertKey.CertFile = "/tmp/easehold-upstream-storage/nonexistent-task8-certificate.pem"
	o.SecureServing.ServerCert.CertKey.KeyFile = "/tmp/easehold-upstream-storage/nonexistent-task8-key.pem"
	completed, err := apiserver.Complete(o, donorProvider{})
	require.NoError(t, err)
	config, _, _, _, err := apiserver.BuildGenericConfig(completed, []*runtime.Scheme{legacyscheme.Scheme}, donorProvider{}.GetOpenAPIDefinitions)
	require.Error(t, err)
	require.Nil(t, config)
	listener.(*net.TCPListener).SetDeadline(time.Now())
	_, err = listener.Accept()
	require.True(t, errors.Is(err, net.ErrClosed), "configuration failure must close the bound listener")
}

func TestSQLFlagsCredentialPrivacy(t *testing.T) {
	sentinel := "task8-private-sentinel"
	o := options.NewAPIMasterOptions(donorAdmission{}, options.StorageBackendTypeMysql)
	o.Mysql.StorageConfig.Mysql.ServerList = []string{"operator:" + sentinel + "@tcp(localhost:3306)/db"}
	flags := pflag.NewFlagSet("mysql", pflag.ContinueOnError)
	o.Mysql.AddFlags(flags)
	var output bytes.Buffer
	flags.SetOutput(&output)
	flags.PrintDefaults()
	if strings.Contains(output.String(), sentinel) {
		t.Fatal("SQL flag help exposed a configured credential")
	}
	require.NoError(t, flags.Parse([]string{"--mysql-servers=operator:replacement@tcp(localhost:3306)/db2"}))
	require.Equal(t, []string{"operator:replacement@tcp(localhost:3306)/db2"}, o.Mysql.StorageConfig.Mysql.ServerList)
	actual, err := flags.GetStringSlice("mysql-servers")
	require.NoError(t, err)
	require.Equal(t, o.Mysql.StorageConfig.Mysql.ServerList, actual)
	o.Mysql.StorageConfig.Mysql.ServerList = []string{"operator:" + sentinel + "@tcp(localhost:3306)/"}
	for _, err := range o.Mysql.Validate() {
		if strings.Contains(err.Error(), sentinel) {
			t.Fatal("configuration validation exposed a credential")
		}
	}
}

func TestSQLiteFlagCredentialPrivacy(t *testing.T) {
	sentinel := "task8-private-sentinel"
	o := options.NewAPIMasterOptions(donorAdmission{}, options.StorageBackendTypeSqlite)
	o.Sqlite.StorageConfig.Sqlite.DSN = "file:/tmp/task8-flags.db?_auth_pass=" + sentinel
	flags := pflag.NewFlagSet("sqlite", pflag.ContinueOnError)
	o.Sqlite.AddFlags(flags)
	var output bytes.Buffer
	flags.SetOutput(&output)
	flags.PrintDefaults()
	if strings.Contains(output.String(), sentinel) {
		t.Fatal("SQLite flag help exposed a credential")
	}
	require.NoError(t, flags.Parse([]string{"--sqlite-dsn=file:/tmp/task8-overridden.db"}))
	actual, err := flags.GetString("sqlite-dsn")
	require.NoError(t, err)
	require.Equal(t, "file:/tmp/task8-overridden.db", actual)
	require.Equal(t, actual, o.Sqlite.StorageConfig.Sqlite.DSN)
}

func TestAuthorizerConstructionFailureCleansConfiguration(t *testing.T) {
	for _, useOIDC := range []bool{false, true} {
		name := "missing-policy"
		if useOIDC {
			name = "invalid-policy-after-oidc-start"
		}
		t.Run(name, func(t *testing.T) {
			directory, err := os.MkdirTemp("/tmp", "easehold-task8-authorizer-")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, os.RemoveAll(directory)) })
			requestStarted := make(chan struct{})
			requestCancelled := make(chan struct{})
			abortRequests := make(chan struct{})
			var requestOnce sync.Once
			issuer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestOnce.Do(func() { close(requestStarted) })
				select {
				case <-r.Context().Done():
					close(requestCancelled)
				case <-abortRequests:
				}
			}))
			t.Cleanup(func() { close(abortRequests); issuer.Close() })
			certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: issuer.Certificate().Raw})
			key, err := x509.MarshalPKCS8PrivateKey(issuer.TLS.Certificates[0].PrivateKey)
			require.NoError(t, err)
			servingCertificate, err := dynamiccertificates.NewStaticCertKeyContent("task8-construction", certificate, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}))
			require.NoError(t, err)
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			t.Cleanup(func() { listener.Close() })
			o := options.NewAPIMasterOptions(donorAdmission{}, options.StorageBackendTypeSqlite)
			o.InsecureServing.BindPort = 0
			o.SecureServing.Listener = listener
			o.SecureServing.ServerCert.GeneratedCert = servingCertificate
			database := filepath.Join(directory, "probes.db")
			o.Sqlite.StorageConfig.Sqlite.DSN = database
			o.Authorization.Modes = []string{"ABAC"}
			o.Authorization.PolicyFile = filepath.Join(directory, "invalid-policy")
			if useOIDC {
				require.NoError(t, unix.Mkfifo(o.Authorization.PolicyFile, 0600))
				caPath := filepath.Join(directory, "issuer-ca.pem")
				require.NoError(t, os.WriteFile(caPath, certificate, 0600))
				o.Authentication = options.NewBuiltInAuthenticationOptions().WithOIDC()
				o.Authentication.OIDC.IssuerURL = issuer.URL
				o.Authentication.OIDC.ClientID = "task8-client"
				o.Authentication.OIDC.UsernameClaim = "sub"
				o.Authentication.OIDC.CAFile = caPath
			}
			completed, err := apiserver.Complete(o, donorProvider{})
			require.NoError(t, err)
			require.Empty(t, o.Validate(), "the nonempty policy path passes options validation")
			type result struct {
				configuration *apiserver.Config
				err           error
			}
			outcome := make(chan result, 1)
			go func() {
				configuration, _, _, _, err := apiserver.BuildGenericConfig(completed, []*runtime.Scheme{legacyscheme.Scheme}, donorProvider{}.GetOpenAPIDefinitions)
				outcome <- result{configuration, err}
			}()
			if useOIDC {
				select {
				case <-requestStarted:
				case <-time.After(5 * time.Second):
					t.Fatal("OIDC discovery did not start before authorizer construction")
				}
				require.Positive(t, countSQLiteDescriptors(t, database), "SQL probes must have acquired native database handles")
				// Authorizer construction is blocked on the policy FIFO until actual OIDC work is observable.
				policy, err := os.OpenFile(o.Authorization.PolicyFile, os.O_WRONLY, 0600)
				require.NoError(t, err)
				_, err = fmt.Fprintln(policy, "invalid-policy-json")
				require.NoError(t, err)
				require.NoError(t, policy.Close())
			}
			var built result
			select {
			case built = <-outcome:
			case <-time.After(5 * time.Second):
				t.Fatal("configuration construction did not return")
			}
			if built.configuration != nil {
				t.Cleanup(built.configuration.Close)
			}
			require.Error(t, built.err, "authorizer constructor failure must reach the public configuration caller")
			if !useOIDC {
				require.ErrorIs(t, built.err, os.ErrNotExist)
			}
			require.Nil(t, built.configuration)
			listener.(*net.TCPListener).SetDeadline(time.Now())
			_, err = listener.Accept()
			require.ErrorIs(t, err, net.ErrClosed)
			require.Zero(t, countSQLiteDescriptors(t, database), "failed construction must close all acquired SQL probe handles")
			if useOIDC {
				select {
				case <-requestCancelled:
				case <-time.After(5 * time.Second):
					t.Fatal("failed construction did not cancel acquired OIDC discovery work")
				}
			}
		})
	}
}

func countSQLiteDescriptors(t *testing.T, database string) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	require.NoError(t, err)
	count := 0
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err == nil && (target == database || target == database+"-wal" || target == database+"-shm") {
			count++
		}
	}
	return count
}
