/*
Copyright 2017 The Kubernetes Authors.

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

package internalversion

import (
	"github.com/seanchann/apimaster/pkg/printers"
)

const (
	loadBalancerWidth = 16

	// labelNodeRolePrefix is a label prefix for node roles
	// It's copied over to here until it's merged in core: https://github.com/kubernetes/kubernetes/pull/39112
	labelNodeRolePrefix = "node-role.kubernetes.io/"

	// nodeLabelRole specifies the role of a node
	nodeLabelRole = "kubernetes.io/role"
)

// AddHandlers adds print handlers for default Kubernetes types dealing with internal versions.
func AddHandlers(h printers.PrintHandler) {
	// authUserColumnDefinitions := []metav1.TableColumnDefinition{
	// 	{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
	// 	{Name: "Completions", Type: "string", Description: ""},
	// 	{Name: "Duration", Type: "string", Description: "Time required to complete the job."},
	// 	{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	// 	{Name: "Containers", Type: "string", Priority: 1, Description: "Names of each container in the template."},
	// 	{Name: "Images", Type: "string", Priority: 1, Description: "Images referenced by each container in the template."},
	// 	{Name: "Selector", Type: "string", Priority: 1, Description: ""},
	// }
	// _ = h.TableHandler(authUserColumnDefinitions, printJob)
	// _ = h.TableHandler(authUserColumnDefinitions, printJobList)

}
