/*
Copyright 2016 The Kubernetes Authors.

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

package bootstrappolicy

import (
	"strings"

	"k8s.io/klog/v2"

	rbacv1 "github.com/seanchann/apimaster/pkg/apis/rbac/v1"
	rbacv1helpers "github.com/seanchann/apimaster/pkg/apis/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const saRolePrefix = "system:controller:"

func addControllerRole(controllerRoles *[]rbacv1.ClusterRole, controllerRoleBindings *[]rbacv1.ClusterRoleBinding, role rbacv1.ClusterRole) {
	if !strings.HasPrefix(role.Name, saRolePrefix) {
		klog.Fatalf(`role %q must start with %q`, role.Name, saRolePrefix)
	}

	for _, existingRole := range *controllerRoles {
		if role.Name == existingRole.Name {
			klog.Fatalf("role %q was already registered", role.Name)
		}
	}

	*controllerRoles = append(*controllerRoles, role)
	addClusterRoleLabel(*controllerRoles)

}

func eventsRule() rbacv1.PolicyRule {
	return rbacv1helpers.NewRule("create", "update", "patch").Groups(legacyGroup, eventsGroup).Resources("events").RuleOrDie()
}

func buildControllerRoles() ([]rbacv1.ClusterRole, []rbacv1.ClusterRoleBinding) {
	// controllerRoles is a slice of roles used for controllers
	controllerRoles := []rbacv1.ClusterRole{}
	// controllerRoleBindings is a slice of roles used for controllers
	controllerRoleBindings := []rbacv1.ClusterRoleBinding{}

	addControllerRole(&controllerRoles, &controllerRoleBindings, rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: saRolePrefix + "clusterrole-aggregation-controller"},
		Rules: []rbacv1.PolicyRule{
			// this controller must have full permissions on clusterroles to allow it to mutate them in any way
			rbacv1helpers.NewRule("escalate", "get", "list", "watch", "update", "patch").Groups(rbacGroup).Resources("clusterroles").RuleOrDie(),
		},
	})

	addControllerRole(&controllerRoles, &controllerRoleBindings, rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: saRolePrefix + "generic-garbage-collector"},
		Rules: []rbacv1.PolicyRule{
			// the GC controller needs to run list/watches, selective gets, and updates against any resource
			rbacv1helpers.NewRule("get", "list", "watch", "patch", "update", "delete").Groups("*").Resources("*").RuleOrDie(),
			eventsRule(),
		},
	})

	return controllerRoles, controllerRoleBindings
}

// ControllerRoles returns the cluster roles used by controllers
func ControllerRoles() []rbacv1.ClusterRole {
	controllerRoles, _ := buildControllerRoles()
	return controllerRoles
}

// ControllerRoleBindings returns the role bindings used by controllers
func ControllerRoleBindings() []rbacv1.ClusterRoleBinding {
	_, controllerRoleBindings := buildControllerRoles()
	return controllerRoleBindings
}
