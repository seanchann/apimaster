//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by conversion-gen. DO NOT EDIT.

package v1

import (
	unsafe "unsafe"

	rbac "github.com/seanchann/apimaster/pkg/apis/rbac"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*AggregationRule)(nil), (*rbac.AggregationRule)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_AggregationRule_To_rbac_AggregationRule(a.(*AggregationRule), b.(*rbac.AggregationRule), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.AggregationRule)(nil), (*AggregationRule)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_AggregationRule_To_v1_AggregationRule(a.(*rbac.AggregationRule), b.(*AggregationRule), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ClusterRole)(nil), (*rbac.ClusterRole)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ClusterRole_To_rbac_ClusterRole(a.(*ClusterRole), b.(*rbac.ClusterRole), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.ClusterRole)(nil), (*ClusterRole)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_ClusterRole_To_v1_ClusterRole(a.(*rbac.ClusterRole), b.(*ClusterRole), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ClusterRoleBinding)(nil), (*rbac.ClusterRoleBinding)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ClusterRoleBinding_To_rbac_ClusterRoleBinding(a.(*ClusterRoleBinding), b.(*rbac.ClusterRoleBinding), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.ClusterRoleBinding)(nil), (*ClusterRoleBinding)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_ClusterRoleBinding_To_v1_ClusterRoleBinding(a.(*rbac.ClusterRoleBinding), b.(*ClusterRoleBinding), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ClusterRoleBindingBuilder)(nil), (*rbac.ClusterRoleBindingBuilder)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ClusterRoleBindingBuilder_To_rbac_ClusterRoleBindingBuilder(a.(*ClusterRoleBindingBuilder), b.(*rbac.ClusterRoleBindingBuilder), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.ClusterRoleBindingBuilder)(nil), (*ClusterRoleBindingBuilder)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_ClusterRoleBindingBuilder_To_v1_ClusterRoleBindingBuilder(a.(*rbac.ClusterRoleBindingBuilder), b.(*ClusterRoleBindingBuilder), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ClusterRoleBindingList)(nil), (*rbac.ClusterRoleBindingList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ClusterRoleBindingList_To_rbac_ClusterRoleBindingList(a.(*ClusterRoleBindingList), b.(*rbac.ClusterRoleBindingList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.ClusterRoleBindingList)(nil), (*ClusterRoleBindingList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_ClusterRoleBindingList_To_v1_ClusterRoleBindingList(a.(*rbac.ClusterRoleBindingList), b.(*ClusterRoleBindingList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ClusterRoleList)(nil), (*rbac.ClusterRoleList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ClusterRoleList_To_rbac_ClusterRoleList(a.(*ClusterRoleList), b.(*rbac.ClusterRoleList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.ClusterRoleList)(nil), (*ClusterRoleList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_ClusterRoleList_To_v1_ClusterRoleList(a.(*rbac.ClusterRoleList), b.(*ClusterRoleList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*PolicyRule)(nil), (*rbac.PolicyRule)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_PolicyRule_To_rbac_PolicyRule(a.(*PolicyRule), b.(*rbac.PolicyRule), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.PolicyRule)(nil), (*PolicyRule)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_PolicyRule_To_v1_PolicyRule(a.(*rbac.PolicyRule), b.(*PolicyRule), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*PolicyRuleBuilder)(nil), (*rbac.PolicyRuleBuilder)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_PolicyRuleBuilder_To_rbac_PolicyRuleBuilder(a.(*PolicyRuleBuilder), b.(*rbac.PolicyRuleBuilder), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.PolicyRuleBuilder)(nil), (*PolicyRuleBuilder)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_PolicyRuleBuilder_To_v1_PolicyRuleBuilder(a.(*rbac.PolicyRuleBuilder), b.(*PolicyRuleBuilder), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Role)(nil), (*rbac.Role)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_Role_To_rbac_Role(a.(*Role), b.(*rbac.Role), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.Role)(nil), (*Role)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_Role_To_v1_Role(a.(*rbac.Role), b.(*Role), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*RoleBinding)(nil), (*rbac.RoleBinding)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_RoleBinding_To_rbac_RoleBinding(a.(*RoleBinding), b.(*rbac.RoleBinding), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.RoleBinding)(nil), (*RoleBinding)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_RoleBinding_To_v1_RoleBinding(a.(*rbac.RoleBinding), b.(*RoleBinding), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*RoleBindingBuilder)(nil), (*rbac.RoleBindingBuilder)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_RoleBindingBuilder_To_rbac_RoleBindingBuilder(a.(*RoleBindingBuilder), b.(*rbac.RoleBindingBuilder), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.RoleBindingBuilder)(nil), (*RoleBindingBuilder)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_RoleBindingBuilder_To_v1_RoleBindingBuilder(a.(*rbac.RoleBindingBuilder), b.(*RoleBindingBuilder), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*RoleBindingList)(nil), (*rbac.RoleBindingList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_RoleBindingList_To_rbac_RoleBindingList(a.(*RoleBindingList), b.(*rbac.RoleBindingList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.RoleBindingList)(nil), (*RoleBindingList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_RoleBindingList_To_v1_RoleBindingList(a.(*rbac.RoleBindingList), b.(*RoleBindingList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*RoleList)(nil), (*rbac.RoleList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_RoleList_To_rbac_RoleList(a.(*RoleList), b.(*rbac.RoleList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.RoleList)(nil), (*RoleList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_RoleList_To_v1_RoleList(a.(*rbac.RoleList), b.(*RoleList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*RoleRef)(nil), (*rbac.RoleRef)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_RoleRef_To_rbac_RoleRef(a.(*RoleRef), b.(*rbac.RoleRef), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.RoleRef)(nil), (*RoleRef)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_RoleRef_To_v1_RoleRef(a.(*rbac.RoleRef), b.(*RoleRef), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Subject)(nil), (*rbac.Subject)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_Subject_To_rbac_Subject(a.(*Subject), b.(*rbac.Subject), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*rbac.Subject)(nil), (*Subject)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_rbac_Subject_To_v1_Subject(a.(*rbac.Subject), b.(*Subject), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1_AggregationRule_To_rbac_AggregationRule(in *AggregationRule, out *rbac.AggregationRule, s conversion.Scope) error {
	out.ClusterRoleSelectors = *(*[]metav1.LabelSelector)(unsafe.Pointer(&in.ClusterRoleSelectors))
	return nil
}

// Convert_v1_AggregationRule_To_rbac_AggregationRule is an autogenerated conversion function.
func Convert_v1_AggregationRule_To_rbac_AggregationRule(in *AggregationRule, out *rbac.AggregationRule, s conversion.Scope) error {
	return autoConvert_v1_AggregationRule_To_rbac_AggregationRule(in, out, s)
}

func autoConvert_rbac_AggregationRule_To_v1_AggregationRule(in *rbac.AggregationRule, out *AggregationRule, s conversion.Scope) error {
	out.ClusterRoleSelectors = *(*[]metav1.LabelSelector)(unsafe.Pointer(&in.ClusterRoleSelectors))
	return nil
}

// Convert_rbac_AggregationRule_To_v1_AggregationRule is an autogenerated conversion function.
func Convert_rbac_AggregationRule_To_v1_AggregationRule(in *rbac.AggregationRule, out *AggregationRule, s conversion.Scope) error {
	return autoConvert_rbac_AggregationRule_To_v1_AggregationRule(in, out, s)
}

func autoConvert_v1_ClusterRole_To_rbac_ClusterRole(in *ClusterRole, out *rbac.ClusterRole, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Rules = *(*[]rbac.PolicyRule)(unsafe.Pointer(&in.Rules))
	out.AggregationRule = (*rbac.AggregationRule)(unsafe.Pointer(in.AggregationRule))
	return nil
}

// Convert_v1_ClusterRole_To_rbac_ClusterRole is an autogenerated conversion function.
func Convert_v1_ClusterRole_To_rbac_ClusterRole(in *ClusterRole, out *rbac.ClusterRole, s conversion.Scope) error {
	return autoConvert_v1_ClusterRole_To_rbac_ClusterRole(in, out, s)
}

func autoConvert_rbac_ClusterRole_To_v1_ClusterRole(in *rbac.ClusterRole, out *ClusterRole, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Rules = *(*[]PolicyRule)(unsafe.Pointer(&in.Rules))
	out.AggregationRule = (*AggregationRule)(unsafe.Pointer(in.AggregationRule))
	return nil
}

// Convert_rbac_ClusterRole_To_v1_ClusterRole is an autogenerated conversion function.
func Convert_rbac_ClusterRole_To_v1_ClusterRole(in *rbac.ClusterRole, out *ClusterRole, s conversion.Scope) error {
	return autoConvert_rbac_ClusterRole_To_v1_ClusterRole(in, out, s)
}

func autoConvert_v1_ClusterRoleBinding_To_rbac_ClusterRoleBinding(in *ClusterRoleBinding, out *rbac.ClusterRoleBinding, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Subjects = *(*[]rbac.Subject)(unsafe.Pointer(&in.Subjects))
	if err := Convert_v1_RoleRef_To_rbac_RoleRef(&in.RoleRef, &out.RoleRef, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1_ClusterRoleBinding_To_rbac_ClusterRoleBinding is an autogenerated conversion function.
func Convert_v1_ClusterRoleBinding_To_rbac_ClusterRoleBinding(in *ClusterRoleBinding, out *rbac.ClusterRoleBinding, s conversion.Scope) error {
	return autoConvert_v1_ClusterRoleBinding_To_rbac_ClusterRoleBinding(in, out, s)
}

func autoConvert_rbac_ClusterRoleBinding_To_v1_ClusterRoleBinding(in *rbac.ClusterRoleBinding, out *ClusterRoleBinding, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Subjects = *(*[]Subject)(unsafe.Pointer(&in.Subjects))
	if err := Convert_rbac_RoleRef_To_v1_RoleRef(&in.RoleRef, &out.RoleRef, s); err != nil {
		return err
	}
	return nil
}

// Convert_rbac_ClusterRoleBinding_To_v1_ClusterRoleBinding is an autogenerated conversion function.
func Convert_rbac_ClusterRoleBinding_To_v1_ClusterRoleBinding(in *rbac.ClusterRoleBinding, out *ClusterRoleBinding, s conversion.Scope) error {
	return autoConvert_rbac_ClusterRoleBinding_To_v1_ClusterRoleBinding(in, out, s)
}

func autoConvert_v1_ClusterRoleBindingBuilder_To_rbac_ClusterRoleBindingBuilder(in *ClusterRoleBindingBuilder, out *rbac.ClusterRoleBindingBuilder, s conversion.Scope) error {
	if err := Convert_v1_ClusterRoleBinding_To_rbac_ClusterRoleBinding(&in.ClusterRoleBinding, &out.ClusterRoleBinding, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1_ClusterRoleBindingBuilder_To_rbac_ClusterRoleBindingBuilder is an autogenerated conversion function.
func Convert_v1_ClusterRoleBindingBuilder_To_rbac_ClusterRoleBindingBuilder(in *ClusterRoleBindingBuilder, out *rbac.ClusterRoleBindingBuilder, s conversion.Scope) error {
	return autoConvert_v1_ClusterRoleBindingBuilder_To_rbac_ClusterRoleBindingBuilder(in, out, s)
}

func autoConvert_rbac_ClusterRoleBindingBuilder_To_v1_ClusterRoleBindingBuilder(in *rbac.ClusterRoleBindingBuilder, out *ClusterRoleBindingBuilder, s conversion.Scope) error {
	if err := Convert_rbac_ClusterRoleBinding_To_v1_ClusterRoleBinding(&in.ClusterRoleBinding, &out.ClusterRoleBinding, s); err != nil {
		return err
	}
	return nil
}

// Convert_rbac_ClusterRoleBindingBuilder_To_v1_ClusterRoleBindingBuilder is an autogenerated conversion function.
func Convert_rbac_ClusterRoleBindingBuilder_To_v1_ClusterRoleBindingBuilder(in *rbac.ClusterRoleBindingBuilder, out *ClusterRoleBindingBuilder, s conversion.Scope) error {
	return autoConvert_rbac_ClusterRoleBindingBuilder_To_v1_ClusterRoleBindingBuilder(in, out, s)
}

func autoConvert_v1_ClusterRoleBindingList_To_rbac_ClusterRoleBindingList(in *ClusterRoleBindingList, out *rbac.ClusterRoleBindingList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]rbac.ClusterRoleBinding)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1_ClusterRoleBindingList_To_rbac_ClusterRoleBindingList is an autogenerated conversion function.
func Convert_v1_ClusterRoleBindingList_To_rbac_ClusterRoleBindingList(in *ClusterRoleBindingList, out *rbac.ClusterRoleBindingList, s conversion.Scope) error {
	return autoConvert_v1_ClusterRoleBindingList_To_rbac_ClusterRoleBindingList(in, out, s)
}

func autoConvert_rbac_ClusterRoleBindingList_To_v1_ClusterRoleBindingList(in *rbac.ClusterRoleBindingList, out *ClusterRoleBindingList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]ClusterRoleBinding)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_rbac_ClusterRoleBindingList_To_v1_ClusterRoleBindingList is an autogenerated conversion function.
func Convert_rbac_ClusterRoleBindingList_To_v1_ClusterRoleBindingList(in *rbac.ClusterRoleBindingList, out *ClusterRoleBindingList, s conversion.Scope) error {
	return autoConvert_rbac_ClusterRoleBindingList_To_v1_ClusterRoleBindingList(in, out, s)
}

func autoConvert_v1_ClusterRoleList_To_rbac_ClusterRoleList(in *ClusterRoleList, out *rbac.ClusterRoleList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]rbac.ClusterRole)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1_ClusterRoleList_To_rbac_ClusterRoleList is an autogenerated conversion function.
func Convert_v1_ClusterRoleList_To_rbac_ClusterRoleList(in *ClusterRoleList, out *rbac.ClusterRoleList, s conversion.Scope) error {
	return autoConvert_v1_ClusterRoleList_To_rbac_ClusterRoleList(in, out, s)
}

func autoConvert_rbac_ClusterRoleList_To_v1_ClusterRoleList(in *rbac.ClusterRoleList, out *ClusterRoleList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]ClusterRole)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_rbac_ClusterRoleList_To_v1_ClusterRoleList is an autogenerated conversion function.
func Convert_rbac_ClusterRoleList_To_v1_ClusterRoleList(in *rbac.ClusterRoleList, out *ClusterRoleList, s conversion.Scope) error {
	return autoConvert_rbac_ClusterRoleList_To_v1_ClusterRoleList(in, out, s)
}

func autoConvert_v1_PolicyRule_To_rbac_PolicyRule(in *PolicyRule, out *rbac.PolicyRule, s conversion.Scope) error {
	out.Verbs = *(*[]string)(unsafe.Pointer(&in.Verbs))
	out.APIGroups = *(*[]string)(unsafe.Pointer(&in.APIGroups))
	out.Resources = *(*[]string)(unsafe.Pointer(&in.Resources))
	out.ResourceNames = *(*[]string)(unsafe.Pointer(&in.ResourceNames))
	out.NonResourceURLs = *(*[]string)(unsafe.Pointer(&in.NonResourceURLs))
	return nil
}

// Convert_v1_PolicyRule_To_rbac_PolicyRule is an autogenerated conversion function.
func Convert_v1_PolicyRule_To_rbac_PolicyRule(in *PolicyRule, out *rbac.PolicyRule, s conversion.Scope) error {
	return autoConvert_v1_PolicyRule_To_rbac_PolicyRule(in, out, s)
}

func autoConvert_rbac_PolicyRule_To_v1_PolicyRule(in *rbac.PolicyRule, out *PolicyRule, s conversion.Scope) error {
	out.Verbs = *(*[]string)(unsafe.Pointer(&in.Verbs))
	out.APIGroups = *(*[]string)(unsafe.Pointer(&in.APIGroups))
	out.Resources = *(*[]string)(unsafe.Pointer(&in.Resources))
	out.ResourceNames = *(*[]string)(unsafe.Pointer(&in.ResourceNames))
	out.NonResourceURLs = *(*[]string)(unsafe.Pointer(&in.NonResourceURLs))
	return nil
}

// Convert_rbac_PolicyRule_To_v1_PolicyRule is an autogenerated conversion function.
func Convert_rbac_PolicyRule_To_v1_PolicyRule(in *rbac.PolicyRule, out *PolicyRule, s conversion.Scope) error {
	return autoConvert_rbac_PolicyRule_To_v1_PolicyRule(in, out, s)
}

func autoConvert_v1_PolicyRuleBuilder_To_rbac_PolicyRuleBuilder(in *PolicyRuleBuilder, out *rbac.PolicyRuleBuilder, s conversion.Scope) error {
	if err := Convert_v1_PolicyRule_To_rbac_PolicyRule(&in.PolicyRule, &out.PolicyRule, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1_PolicyRuleBuilder_To_rbac_PolicyRuleBuilder is an autogenerated conversion function.
func Convert_v1_PolicyRuleBuilder_To_rbac_PolicyRuleBuilder(in *PolicyRuleBuilder, out *rbac.PolicyRuleBuilder, s conversion.Scope) error {
	return autoConvert_v1_PolicyRuleBuilder_To_rbac_PolicyRuleBuilder(in, out, s)
}

func autoConvert_rbac_PolicyRuleBuilder_To_v1_PolicyRuleBuilder(in *rbac.PolicyRuleBuilder, out *PolicyRuleBuilder, s conversion.Scope) error {
	if err := Convert_rbac_PolicyRule_To_v1_PolicyRule(&in.PolicyRule, &out.PolicyRule, s); err != nil {
		return err
	}
	return nil
}

// Convert_rbac_PolicyRuleBuilder_To_v1_PolicyRuleBuilder is an autogenerated conversion function.
func Convert_rbac_PolicyRuleBuilder_To_v1_PolicyRuleBuilder(in *rbac.PolicyRuleBuilder, out *PolicyRuleBuilder, s conversion.Scope) error {
	return autoConvert_rbac_PolicyRuleBuilder_To_v1_PolicyRuleBuilder(in, out, s)
}

func autoConvert_v1_Role_To_rbac_Role(in *Role, out *rbac.Role, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Rules = *(*[]rbac.PolicyRule)(unsafe.Pointer(&in.Rules))
	return nil
}

// Convert_v1_Role_To_rbac_Role is an autogenerated conversion function.
func Convert_v1_Role_To_rbac_Role(in *Role, out *rbac.Role, s conversion.Scope) error {
	return autoConvert_v1_Role_To_rbac_Role(in, out, s)
}

func autoConvert_rbac_Role_To_v1_Role(in *rbac.Role, out *Role, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Rules = *(*[]PolicyRule)(unsafe.Pointer(&in.Rules))
	return nil
}

// Convert_rbac_Role_To_v1_Role is an autogenerated conversion function.
func Convert_rbac_Role_To_v1_Role(in *rbac.Role, out *Role, s conversion.Scope) error {
	return autoConvert_rbac_Role_To_v1_Role(in, out, s)
}

func autoConvert_v1_RoleBinding_To_rbac_RoleBinding(in *RoleBinding, out *rbac.RoleBinding, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Subjects = *(*[]rbac.Subject)(unsafe.Pointer(&in.Subjects))
	if err := Convert_v1_RoleRef_To_rbac_RoleRef(&in.RoleRef, &out.RoleRef, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1_RoleBinding_To_rbac_RoleBinding is an autogenerated conversion function.
func Convert_v1_RoleBinding_To_rbac_RoleBinding(in *RoleBinding, out *rbac.RoleBinding, s conversion.Scope) error {
	return autoConvert_v1_RoleBinding_To_rbac_RoleBinding(in, out, s)
}

func autoConvert_rbac_RoleBinding_To_v1_RoleBinding(in *rbac.RoleBinding, out *RoleBinding, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Subjects = *(*[]Subject)(unsafe.Pointer(&in.Subjects))
	if err := Convert_rbac_RoleRef_To_v1_RoleRef(&in.RoleRef, &out.RoleRef, s); err != nil {
		return err
	}
	return nil
}

// Convert_rbac_RoleBinding_To_v1_RoleBinding is an autogenerated conversion function.
func Convert_rbac_RoleBinding_To_v1_RoleBinding(in *rbac.RoleBinding, out *RoleBinding, s conversion.Scope) error {
	return autoConvert_rbac_RoleBinding_To_v1_RoleBinding(in, out, s)
}

func autoConvert_v1_RoleBindingBuilder_To_rbac_RoleBindingBuilder(in *RoleBindingBuilder, out *rbac.RoleBindingBuilder, s conversion.Scope) error {
	if err := Convert_v1_RoleBinding_To_rbac_RoleBinding(&in.RoleBinding, &out.RoleBinding, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1_RoleBindingBuilder_To_rbac_RoleBindingBuilder is an autogenerated conversion function.
func Convert_v1_RoleBindingBuilder_To_rbac_RoleBindingBuilder(in *RoleBindingBuilder, out *rbac.RoleBindingBuilder, s conversion.Scope) error {
	return autoConvert_v1_RoleBindingBuilder_To_rbac_RoleBindingBuilder(in, out, s)
}

func autoConvert_rbac_RoleBindingBuilder_To_v1_RoleBindingBuilder(in *rbac.RoleBindingBuilder, out *RoleBindingBuilder, s conversion.Scope) error {
	if err := Convert_rbac_RoleBinding_To_v1_RoleBinding(&in.RoleBinding, &out.RoleBinding, s); err != nil {
		return err
	}
	return nil
}

// Convert_rbac_RoleBindingBuilder_To_v1_RoleBindingBuilder is an autogenerated conversion function.
func Convert_rbac_RoleBindingBuilder_To_v1_RoleBindingBuilder(in *rbac.RoleBindingBuilder, out *RoleBindingBuilder, s conversion.Scope) error {
	return autoConvert_rbac_RoleBindingBuilder_To_v1_RoleBindingBuilder(in, out, s)
}

func autoConvert_v1_RoleBindingList_To_rbac_RoleBindingList(in *RoleBindingList, out *rbac.RoleBindingList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]rbac.RoleBinding)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1_RoleBindingList_To_rbac_RoleBindingList is an autogenerated conversion function.
func Convert_v1_RoleBindingList_To_rbac_RoleBindingList(in *RoleBindingList, out *rbac.RoleBindingList, s conversion.Scope) error {
	return autoConvert_v1_RoleBindingList_To_rbac_RoleBindingList(in, out, s)
}

func autoConvert_rbac_RoleBindingList_To_v1_RoleBindingList(in *rbac.RoleBindingList, out *RoleBindingList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]RoleBinding)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_rbac_RoleBindingList_To_v1_RoleBindingList is an autogenerated conversion function.
func Convert_rbac_RoleBindingList_To_v1_RoleBindingList(in *rbac.RoleBindingList, out *RoleBindingList, s conversion.Scope) error {
	return autoConvert_rbac_RoleBindingList_To_v1_RoleBindingList(in, out, s)
}

func autoConvert_v1_RoleList_To_rbac_RoleList(in *RoleList, out *rbac.RoleList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]rbac.Role)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1_RoleList_To_rbac_RoleList is an autogenerated conversion function.
func Convert_v1_RoleList_To_rbac_RoleList(in *RoleList, out *rbac.RoleList, s conversion.Scope) error {
	return autoConvert_v1_RoleList_To_rbac_RoleList(in, out, s)
}

func autoConvert_rbac_RoleList_To_v1_RoleList(in *rbac.RoleList, out *RoleList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]Role)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_rbac_RoleList_To_v1_RoleList is an autogenerated conversion function.
func Convert_rbac_RoleList_To_v1_RoleList(in *rbac.RoleList, out *RoleList, s conversion.Scope) error {
	return autoConvert_rbac_RoleList_To_v1_RoleList(in, out, s)
}

func autoConvert_v1_RoleRef_To_rbac_RoleRef(in *RoleRef, out *rbac.RoleRef, s conversion.Scope) error {
	out.APIGroup = in.APIGroup
	out.Kind = in.Kind
	out.Name = in.Name
	return nil
}

// Convert_v1_RoleRef_To_rbac_RoleRef is an autogenerated conversion function.
func Convert_v1_RoleRef_To_rbac_RoleRef(in *RoleRef, out *rbac.RoleRef, s conversion.Scope) error {
	return autoConvert_v1_RoleRef_To_rbac_RoleRef(in, out, s)
}

func autoConvert_rbac_RoleRef_To_v1_RoleRef(in *rbac.RoleRef, out *RoleRef, s conversion.Scope) error {
	out.APIGroup = in.APIGroup
	out.Kind = in.Kind
	out.Name = in.Name
	return nil
}

// Convert_rbac_RoleRef_To_v1_RoleRef is an autogenerated conversion function.
func Convert_rbac_RoleRef_To_v1_RoleRef(in *rbac.RoleRef, out *RoleRef, s conversion.Scope) error {
	return autoConvert_rbac_RoleRef_To_v1_RoleRef(in, out, s)
}

func autoConvert_v1_Subject_To_rbac_Subject(in *Subject, out *rbac.Subject, s conversion.Scope) error {
	out.Kind = in.Kind
	out.APIGroup = in.APIGroup
	out.Name = in.Name
	out.Namespace = in.Namespace
	return nil
}

// Convert_v1_Subject_To_rbac_Subject is an autogenerated conversion function.
func Convert_v1_Subject_To_rbac_Subject(in *Subject, out *rbac.Subject, s conversion.Scope) error {
	return autoConvert_v1_Subject_To_rbac_Subject(in, out, s)
}

func autoConvert_rbac_Subject_To_v1_Subject(in *rbac.Subject, out *Subject, s conversion.Scope) error {
	out.Kind = in.Kind
	out.APIGroup = in.APIGroup
	out.Name = in.Name
	out.Namespace = in.Namespace
	return nil
}

// Convert_rbac_Subject_To_v1_Subject is an autogenerated conversion function.
func Convert_rbac_Subject_To_v1_Subject(in *rbac.Subject, out *Subject, s conversion.Scope) error {
	return autoConvert_rbac_Subject_To_v1_Subject(in, out, s)
}