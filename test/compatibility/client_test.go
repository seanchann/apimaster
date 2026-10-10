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
	"net/http"
	"testing"
	"time"

	"github.com/seanchann/apimaster/pkg/client/clientset"
	"github.com/seanchann/apimaster/pkg/client/clientset/fake"
	"github.com/seanchann/apimaster/pkg/client/informers"
	corereslisters "github.com/seanchann/apimaster/pkg/client/listers/coreres/v1"
	rbaclisters "github.com/seanchann/apimaster/pkg/client/listers/rbac/v1"
	"k8s.io/client-go/rest"
)

var (
	_ func(*rest.Config) (*clientset.Clientset, error)               = clientset.NewForConfig
	_ func(*rest.Config, *http.Client) (*clientset.Clientset, error) = clientset.NewForConfigAndClient
	_ func(*rest.Config) *clientset.Clientset                        = clientset.NewForConfigOrDie
	_ func(rest.Interface) *clientset.Clientset                      = clientset.New
	_ corereslisters.NamespaceLister
	_ rbaclisters.RoleLister
)

func TestDonorInformerPaths(t *testing.T) {
	factory := informers.NewSharedInformerFactory(fake.NewSimpleClientset(), time.Minute)
	if factory.Coreres().V1().Namespaces().Lister() == nil || factory.Rbac().V1().Roles().Lister() == nil {
		t.Fatal("missing donor informer/lister surface")
	}
}
