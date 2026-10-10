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

package options

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/server"
)

// ApplyFeaturesTo requires explicit Kubernetes capabilities for priority and fairness.
func (o *APIMasterOptions) ApplyFeaturesTo(config *server.Config) error {
	if o.Features == nil {
		return nil
	}
	if o.Features.EnablePriorityAndFairness {
		if o.KubernetesClient == nil || o.KubernetesInformers == nil {
			return fmt.Errorf("priority and fairness requires an explicit Kubernetes client and matching informer factory")
		}
		resources, err := o.KubernetesClient.Discovery().ServerResourcesForGroupVersion("flowcontrol.apiserver.k8s.io/v1")
		if err != nil {
			return fmt.Errorf("priority and fairness requires flow-control API discovery: %w", err)
		}
		required := sets.New("flowschemas", "prioritylevelconfigurations")
		for _, resource := range resources.APIResources {
			verbs := sets.New(resource.Verbs...)
			if verbs.HasAll("get", "list", "watch") {
				required.Delete(resource.Name)
			}
		}
		if required.Len() != 0 {
			return fmt.Errorf("priority and fairness requires Kubernetes flow-control resources with get/list/watch capability")
		}
	}
	return o.Features.ApplyTo(config, o.KubernetesClient, o.KubernetesInformers)
}
