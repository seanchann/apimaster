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

package clientset

import (
	"fmt"
	"net/http"

	coreresv1 "github.com/seanchann/apimaster/pkg/client/generated/clientset/typed/coreres/v1"
	rbacv1 "github.com/seanchann/apimaster/pkg/client/generated/clientset/typed/rbac/v1"

	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

// ClientSetProvider is the interface for the clientset provider.
type ClientSetProvider[T any] interface {
	New(c rest.Interface) *T
	NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (*T, error)
}

type Interface interface {
	CoreresV1() coreresv1.CoreresV1Interface
	RbacV1() rbacv1.RbacV1Interface
}

// GenericClientset contains the clients for groups.
type GenericClientset[T any] struct {
	UserClient *T
	coreresV1  *coreresv1.CoreresV1Client
	rbacV1     *rbacv1.RbacV1Client
}

// CoreresV1 retrieves the CoreresV1Client
func (c *GenericClientset[T]) CoreresV1() coreresv1.CoreresV1Interface {
	return c.coreresV1
}

// RbacV1 retrieves the RbacV1Client
func (c *GenericClientset[T]) RbacV1() rbacv1.RbacV1Interface {
	return c.rbacV1
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig[T any](c *rest.Config, provider ClientSetProvider[T]) (*GenericClientset[T], error) {
	configShallowCopy := *c

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	// share the transport between all clients
	httpClient, err := rest.HTTPClientFor(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return NewForConfigAndClient[T](&configShallowCopy, httpClient, provider)
}

// NewForConfigAndClient creates a new Clientset for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfigAndClient will generate a rate-limiter in configShallowCopy.
func NewForConfigAndClient[T any](c *rest.Config, httpClient *http.Client, provider ClientSetProvider[T]) (*GenericClientset[T], error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	var cs GenericClientset[T]
	var err error
	if provider != nil {
		gencs, err := provider.NewForConfigAndClient(&configShallowCopy, httpClient)
		if err != nil {
			return nil, err
		}
		cs.UserClient = gencs
	}

	cs.coreresV1, err = coreresv1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.rbacV1, err = rbacv1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie[T any](c *rest.Config, provider ClientSetProvider[T]) *GenericClientset[T] {
	cs, err := NewForConfig[T](c, provider)
	if err != nil {
		panic(err)
	}
	return cs
}

// New creates a new Clientset for the given RESTClient.
func New[T any](c rest.Interface, provider ClientSetProvider[T]) *GenericClientset[T] {
	var cs GenericClientset[T]
	cs.UserClient = provider.New(c)

	cs.coreresV1 = coreresv1.New(c)
	cs.rbacV1 = rbacv1.New(c)

	return &cs
}
