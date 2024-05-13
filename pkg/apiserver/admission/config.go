/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package admission

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"k8s.io/apiserver/pkg/admission"
	webhookinit "k8s.io/apiserver/pkg/admission/plugin/webhook/initializer"
	genericapiserver "k8s.io/apiserver/pkg/server"
	egressselector "k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/apiserver/pkg/util/webhook"
	"k8s.io/client-go/rest"
)

// Config holds the configuration needed to for initialize the admission plugins
type Config struct {
	LoopbackClientConfig *rest.Config
	ExternalInformers    interface{}
}

// New sets up the plugins and admission start hooks needed for admission
func (c *Config) New(proxyTransport *http.Transport, egressSelector *egressselector.EgressSelector, serviceResolver webhook.ServiceResolver, tp trace.TracerProvider) ([]admission.PluginInitializer, genericapiserver.PostStartHookFunc, error) {
	webhookAuthResolverWrapper := webhook.NewDefaultAuthenticationInfoResolverWrapper(proxyTransport, egressSelector, c.LoopbackClientConfig, tp)
	webhookPluginInitializer := webhookinit.NewPluginInitializer(webhookAuthResolverWrapper, serviceResolver)

	// clientset, err := kubernetes.NewForConfig(c.LoopbackClientConfig)
	// if err != nil {
	// 	return nil, nil, err
	// }
	// discoveryClient := cacheddiscovery.NewMemCacheClient(clientset.Discovery())
	// discoveryRESTMapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)

	// admissionPostStartHook := func(context genericapiserver.PostStartHookContext) error {
	// 	discoveryRESTMapper.Reset()
	// 	go utilwait.Until(discoveryRESTMapper.Reset, 30*time.Second, context.StopCh)
	// 	return nil
	// }

	// return []admission.PluginInitializer{webhookPluginInitializer}, admissionPostStartHook, nil
	return []admission.PluginInitializer{webhookPluginInitializer}, nil, nil

}
