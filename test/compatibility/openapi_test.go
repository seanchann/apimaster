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
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	abacv1beta1 "github.com/seanchann/apimaster/pkg/apis/abac/v1beta1"
	coreresv1 "github.com/seanchann/apimaster/pkg/apis/coreres/v1"
	rbacv1 "github.com/seanchann/apimaster/pkg/apis/rbac/v1"
	"github.com/seanchann/apimaster/pkg/apiserver"
	"github.com/seanchann/apimaster/pkg/apiserver/options"
	generatedopenapi "github.com/seanchann/apimaster/pkg/generated/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	"k8s.io/kube-openapi/pkg/builder3"
	"k8s.io/kube-openapi/pkg/common"
	openapiutil "k8s.io/kube-openapi/pkg/util"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// customOpenAPIObject exercises pointer model-name precedence over canonical aliases.
type customOpenAPIObject struct {
	metav1.TypeMeta `json:",inline"`
	Value           string `json:"value,omitempty"`
}

func (o *customOpenAPIObject) DeepCopyObject() runtime.Object { copy := *o; return &copy }
func (*customOpenAPIObject) OpenAPIModelName() string         { return "custom.example/v1.OwnedObject" }
func (*customOpenAPIObject) OpenAPICanonicalTypeName() string { return "unused.example/v1.LegacyName" }

func TestOpenAPIModelNamesAndTypeConversion(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, coreresv1.AddToScheme(scheme))
	require.NoError(t, rbacv1.AddToScheme(scheme))
	customGVK := schema.GroupVersionKind{Group: "custom.example", Version: "v1", Kind: "OwnedObject"}
	scheme.AddKnownTypeWithName(customGVK, &customOpenAPIObject{})
	second := runtime.NewScheme()
	aliasGVK := schema.GroupVersionKind{Group: "alias.example", Version: "v2", Kind: "OwnedObject"}
	second.AddKnownTypeWithName(aliasGVK, &customOpenAPIObject{})
	definitions := func(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
		result := generatedopenapi.GetOpenAPIDefinitions(ref)
		result[(&customOpenAPIObject{}).OpenAPIModelName()] = common.OpenAPIDefinition{Schema: spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"object"}, Properties: map[string]spec.Schema{
			"apiVersion": {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
			"kind":       {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
			"value":      {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
		}}}}
		return result
	}
	configuration := buildOpenAPIConfiguration(t, []*runtime.Scheme{scheme, second}, definitions)
	config := configuration.GenericConfig
	objects := map[string]runtime.Object{}
	for _, registered := range []*runtime.Scheme{scheme, second} {
		for gvk, typ := range registered.AllKnownTypes() {
			if gvk.Version == runtime.APIVersionInternal || (!strings.HasPrefix(typ.PkgPath(), "github.com/seanchann/apimaster/pkg/apis/") && gvk != customGVK && gvk != aliasGVK) {
				continue
			}
			obj, err := registered.New(gvk)
			require.NoError(t, err)
			obj.GetObjectKind().SetGroupVersionKind(gvk)
			objects[gvk.String()] = obj
		}
	}
	require.Len(t, objects, 12, "all ten served AM resource/list kinds and two custom aliases")
	names := []string{}
	for _, obj := range objects {
		names = append(names, openapiutil.GetCanonicalTypeName(obj))
	}
	models, err := builder3.BuildOpenAPIDefinitionsForResources(config.OpenAPIV3Config, names...)
	require.NoError(t, err)
	converter, err := managedfields.NewTypeConverter(models, false)
	require.NoError(t, err)
	keys := []string{}
	for name := range objects {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			obj := objects[key]
			gvk := obj.GetObjectKind().GroupVersionKind()
			name := openapiutil.GetCanonicalTypeName(obj)
			for _, lookup := range []func(string) (string, spec.Extensions){config.OpenAPIConfig.GetDefinitionName, config.OpenAPIV3Config.GetDefinitionName} {
				actual, extensions := lookup(name)
				require.Equal(t, name, actual, "generated definition and reference names must not change")
				if _, ok := extensions["x-kubernetes-group-version-kind"]; !ok {
					t.Errorf("missing runtime GVK extension for %s", name)
				}
			}
			model := models[common.EscapeJsonPointer(name)]
			require.NotNil(t, model)
			assert.Contains(t, model.Extensions["x-kubernetes-group-version-kind"], map[string]interface{}{"group": gvk.Group, "version": gvk.Version, "kind": gvk.Kind})
			typed, err := converter.ObjectToTyped(obj)
			require.NoError(t, err)
			decoded, err := converter.TypedToObject(typed)
			require.NoError(t, err)
			require.Equal(t, gvk, decoded.GetObjectKind().GroupVersionKind())
		})
	}
	for _, lookup := range []func(string) (string, spec.Extensions){config.OpenAPIConfig.GetDefinitionName, config.OpenAPIV3Config.GetDefinitionName} {
		name, ext := lookup("unknown.example/v1.Unknown")
		require.Equal(t, "unknown.example/v1.Unknown", name)
		require.Empty(t, ext, "unknown definitions must not acquire an invented GVK")
	}
	// Generated dependencies and all emitted refs must still identify the same models.
	for _, name := range names {
		for _, dependency := range config.OpenAPIV3Config.Definitions[name].Dependencies {
			require.Contains(t, models, common.EscapeJsonPointer(dependency))
		}
	}
	encoded, err := json.Marshal(models)
	require.NoError(t, err)
	var document interface{}
	require.NoError(t, json.Unmarshal(encoded, &document))
	var checkRefs func(interface{})
	checkRefs = func(node interface{}) {
		switch typed := node.(type) {
		case map[string]interface{}:
			if ref, ok := typed["$ref"].(string); ok {
				require.True(t, strings.HasPrefix(ref, "#/components/schemas/"))
				require.Contains(t, models, strings.TrimPrefix(ref, "#/components/schemas/"))
			}
			for _, value := range typed {
				checkRefs(value)
			}
		case []interface{}:
			for _, value := range typed {
				checkRefs(value)
			}
		}
	}
	checkRefs(document)
}

// Offline ABAC policy is not a served API. Its embedded TypeMeta promotes the
// model-name method, which must be honored without claiming Policy conversion.
func TestOpenAPIInheritedModelName(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, abacv1beta1.AddToScheme(scheme))
	config := buildOpenAPIConfiguration(t, []*runtime.Scheme{scheme}, generatedopenapi.GetOpenAPIDefinitions).GenericConfig
	canonical := openapiutil.GetCanonicalTypeName(&abacv1beta1.Policy{})
	require.Equal(t, metav1.TypeMeta{}.OpenAPIModelName(), canonical)
	for _, lookup := range []func(string) (string, spec.Extensions){config.OpenAPIConfig.GetDefinitionName, config.OpenAPIV3Config.GetDefinitionName} {
		name, extensions := lookup(canonical)
		require.Equal(t, canonical, name)
		require.Contains(t, extensions["x-kubernetes-group-version-kind"], map[string]interface{}{"group": abacv1beta1.SchemeGroupVersion.Group, "version": "v1beta1", "kind": "Policy"})
	}
}

func buildOpenAPIConfiguration(t *testing.T, schemes []*runtime.Scheme, definitions common.GetOpenAPIDefinitions) *apiserver.Config {
	t.Helper()
	directory, err := os.MkdirTemp("/tmp", "easehold-task9-schema-")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(directory)) })
	certificateFixture := httptest.NewTLSServer(http.NotFoundHandler())
	certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateFixture.Certificate().Raw})
	key, err := x509.MarshalPKCS8PrivateKey(certificateFixture.TLS.Certificates[0].PrivateKey)
	require.NoError(t, err)
	certificateFixture.Close()
	serving, err := dynamiccertificates.NewStaticCertKeyContent("task9-schema", certificate, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}))
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() })
	o := options.NewAPIMasterOptions(donorAdmission{}, options.StorageBackendTypeSqlite)
	o.InsecureServing.BindPort = 0
	o.SecureServing.Listener = listener
	o.SecureServing.ServerCert.GeneratedCert = serving
	o.Sqlite.StorageConfig.Sqlite.DSN = filepath.Join(directory, "probes.db")
	o.Authorization.Modes = []string{"AlwaysAllow"}
	completed, err := apiserver.Complete(o, donorProvider{})
	require.NoError(t, err)
	configuration, _, _, _, err := apiserver.BuildGenericConfig(completed, schemes, definitions)
	require.NoError(t, err)
	t.Cleanup(configuration.Close)
	return configuration
}
