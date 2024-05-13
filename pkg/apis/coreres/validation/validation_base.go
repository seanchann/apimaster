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

package validation

import (
	"fmt"
	"net"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/seanchann/apimaster/pkg/apis/coreres"
	"github.com/seanchann/apimaster/pkg/apis/coreres/helper"
	"k8s.io/apimachinery/pkg/api/resource"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	netutils "k8s.io/utils/net"
)

const isNegativeErrorMsg string = apimachineryvalidation.IsNegativeErrorMsg
const isInvalidQuotaResource string = `must be a standard resource for quota`
const fieldImmutableErrorMsg string = apimachineryvalidation.FieldImmutableErrorMsg
const isNotIntegerErrorMsg string = `must be an integer`
const isNotPositiveErrorMsg string = `must be greater than zero`

var pdPartitionErrorMsg string = validation.InclusiveRangeError(1, 255)
var fileModeErrorMsg = "must be a number between 0 and 0777 (octal), both inclusive"

// BannedOwners is a black list of object that are not allowed to be owners.
var BannedOwners = apimachineryvalidation.BannedOwners

var iscsiInitiatorIqnRegex = regexp.MustCompile(`iqn\.\d{4}-\d{2}\.([[:alnum:]-.]+)(:[^,;*&$|\s]+)$`)
var iscsiInitiatorEuiRegex = regexp.MustCompile(`^eui.[[:alnum:]]{16}$`)
var iscsiInitiatorNaaRegex = regexp.MustCompile(`^naa.[[:alnum:]]{32}$`)

// ValidateHasLabel requires that metav1.ObjectMeta has a Label with key and expectedValue
func ValidateHasLabel(meta metav1.ObjectMeta, fldPath *field.Path, key, expectedValue string) field.ErrorList {
	allErrs := field.ErrorList{}
	actualValue, found := meta.Labels[key]
	if !found {
		allErrs = append(allErrs, field.Required(fldPath.Child("labels").Key(key),
			fmt.Sprintf("must be '%s'", expectedValue)))
		return allErrs
	}
	if actualValue != expectedValue {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("labels").Key(key), meta.Labels,
			fmt.Sprintf("must be '%s'", expectedValue)))
	}
	return allErrs
}

// ValidateAnnotations validates that a set of annotations are correctly defined.
func ValidateAnnotations(annotations map[string]string, fldPath *field.Path) field.ErrorList {
	return apimachineryvalidation.ValidateAnnotations(annotations, fldPath)
}

func ValidateDNS1123Label(value string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for _, msg := range validation.IsDNS1123Label(value) {
		allErrs = append(allErrs, field.Invalid(fldPath, value, msg))
	}
	return allErrs
}

// ValidateQualifiedName validates if name is what Kubernetes calls a "qualified name".
func ValidateQualifiedName(value string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for _, msg := range validation.IsQualifiedName(value) {
		allErrs = append(allErrs, field.Invalid(fldPath, value, msg))
	}
	return allErrs
}

// ValidateDNS1123Subdomain validates that a name is a proper DNS subdomain.
func ValidateDNS1123Subdomain(value string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for _, msg := range validation.IsDNS1123Subdomain(value) {
		allErrs = append(allErrs, field.Invalid(fldPath, value, msg))
	}
	return allErrs
}

// ValidateRuntimeClassName can be used to check whether the given RuntimeClass name is valid.
// Prefix indicates this name will be used as part of generation, in which case
// trailing dashes are allowed.
func ValidateRuntimeClassName(name string, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	for _, msg := range apimachineryvalidation.NameIsDNSSubdomain(name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath, name, msg))
	}
	return allErrs
}

// Validates that given value is not negative.
func ValidateNonnegativeField(value int64, fldPath *field.Path) field.ErrorList {
	return apimachineryvalidation.ValidateNonnegativeField(value, fldPath)
}

// IsNegativeErrorMsg is a error message for value must be greater than or equal to 0.
const IsOutIntervalErrorMsg string = `must be within the interval [%d,%d]`

// ValidateIntervalField validates that given value is in some interval
func ValidateIntervalField(value int64, start, end int64, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if !(value >= start && value <= end) {
		allErrs = append(allErrs, field.Invalid(fldPath, value, fmt.Sprintf(IsOutIntervalErrorMsg, start, end)))
	}
	return allErrs
}

// Validates that a Quantity is not negative
func ValidateNonnegativeQuantity(value resource.Quantity, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if value.Cmp(resource.Quantity{}) < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, value.String(), isNegativeErrorMsg))
	}
	return allErrs
}

// Validates that a Quantity is positive
func ValidatePositiveQuantityValue(value resource.Quantity, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if value.Cmp(resource.Quantity{}) <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, value.String(), isNotPositiveErrorMsg))
	}
	return allErrs
}

func ValidateImmutableField(newVal, oldVal interface{}, fldPath *field.Path) field.ErrorList {
	return apimachineryvalidation.ValidateImmutableField(newVal, oldVal, fldPath)
}

func ValidateImmutableAnnotation(newVal string, oldVal string, annotation string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if oldVal != newVal {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("annotations", annotation), newVal, fieldImmutableErrorMsg))
	}
	return allErrs
}

// ValidateNameFunc validates that the provided name is valid for a given resource type.
// Not all resources have the same validation rules for names. Prefix is true
// if the name will have a value appended to it.  If the name is not valid,
// this returns a list of descriptions of individual characteristics of the
// value that were not valid.  Otherwise this returns an empty list or nil.
type ValidateNameFunc apimachineryvalidation.ValidateNameFunc

// ValidateNamespaceName can be used to check whether the given namespace name is valid.
// Prefix indicates this name will be used as part of generation, in which case
// trailing dashes are allowed.
var ValidateNamespaceName = apimachineryvalidation.ValidateNamespaceName

// ValidateObjectMeta validates an object's metadata on creation. It expects that name generation has already
// been performed.
// It doesn't return an error for rootscoped resources with namespace, because namespace should already be cleared before.
// TODO: Remove calls to this method scattered in validations of specific resources, e.g., ValidatePodUpdate.
func ValidateObjectMeta(meta *metav1.ObjectMeta, requiresNamespace bool, nameFn ValidateNameFunc, fldPath *field.Path) field.ErrorList {
	allErrs := apimachineryvalidation.ValidateObjectMeta(meta, requiresNamespace, apimachineryvalidation.ValidateNameFunc(nameFn), fldPath)
	// run additional checks for the finalizer name
	for i := range meta.Finalizers {
		allErrs = append(allErrs, validateKubeFinalizerName(string(meta.Finalizers[i]), fldPath.Child("finalizers").Index(i))...)
	}
	return allErrs
}

// ValidateObjectMetaUpdate validates an object's metadata when updated
func ValidateObjectMetaUpdate(newMeta, oldMeta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	allErrs := apimachineryvalidation.ValidateObjectMetaUpdate(newMeta, oldMeta, fldPath)
	// run additional checks for the finalizer name
	for i := range newMeta.Finalizers {
		allErrs = append(allErrs, validateKubeFinalizerName(string(newMeta.Finalizers[i]), fldPath.Child("finalizers").Index(i))...)
	}

	return allErrs
}

// This validate will make sure targetPath:
// 1. is not abs path
// 2. does not have any element which is ".."
func validateLocalDescendingPath(targetPath string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if path.IsAbs(targetPath) {
		allErrs = append(allErrs, field.Invalid(fldPath, targetPath, "must be a relative path"))
	}

	allErrs = append(allErrs, validatePathNoBacksteps(targetPath, fldPath)...)

	return allErrs
}

// validatePathNoBacksteps makes sure the targetPath does not have any `..` path elements when split
//
// This assumes the OS of the apiserver and the nodes are the same. The same check should be done
// on the node to ensure there are no backsteps.
func validatePathNoBacksteps(targetPath string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	parts := strings.Split(filepath.ToSlash(targetPath), "/")
	for _, item := range parts {
		if item == ".." {
			allErrs = append(allErrs, field.Invalid(fldPath, targetPath, "must not contain '..'"))
			break // even for `../../..`, one error is sufficient to make the point
		}
	}
	return allErrs
}

// This validate will make sure targetPath:
// 1. is not abs path
// 2. does not contain any '..' elements
// 3. does not start with '..'
func validateLocalNonReservedPath(targetPath string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateLocalDescendingPath(targetPath, fldPath)...)
	// Don't report this error if the check for .. elements already caught it.
	if strings.HasPrefix(targetPath, "..") && !strings.HasPrefix(targetPath, "../") {
		allErrs = append(allErrs, field.Invalid(fldPath, targetPath, "must not start with '..'"))
	}
	return allErrs
}

func ValidatePortNumOrName(port intstr.IntOrString, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if port.Type == intstr.Int {
		for _, msg := range validation.IsValidPortNum(port.IntValue()) {
			allErrs = append(allErrs, field.Invalid(fldPath, port.IntValue(), msg))
		}
	} else if port.Type == intstr.String {
		for _, msg := range validation.IsValidPortName(port.StrVal) {
			allErrs = append(allErrs, field.Invalid(fldPath, port.StrVal, msg))
		}
	} else {
		allErrs = append(allErrs, field.InternalError(fldPath, fmt.Errorf("unknown type: %v", port.Type)))
	}
	return allErrs
}

// ValidateFieldAcceptList checks that only allowed fields are set.
// The value must be a struct (not a pointer to a struct!).
func validateFieldAllowList(value interface{}, allowedFields map[string]bool, errorText string, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	reflectType, reflectValue := reflect.TypeOf(value), reflect.ValueOf(value)
	for i := 0; i < reflectType.NumField(); i++ {
		f := reflectType.Field(i)
		if allowedFields[f.Name] {
			continue
		}

		// Compare the value of this field to its zero value to determine if it has been set
		if !reflect.DeepEqual(reflectValue.Field(i).Interface(), reflect.Zero(f.Type).Interface()) {
			r, n := utf8.DecodeRuneInString(f.Name)
			lcName := string(unicode.ToLower(r)) + f.Name[n:]
			allErrs = append(allErrs, field.Forbidden(fldPath.Child(lcName), errorText))
		}
	}

	return allErrs
}

const (
	// Limits on various DNS parameters. These are derived from
	// restrictions in Linux libc name resolution handling.
	// Max number of DNS name servers.
	MaxDNSNameservers = 3
	// Max number of domains in the search path list.
	MaxDNSSearchPaths = 32
	// Max number of characters in the search path.
	MaxDNSSearchListChars = 2048
)

const (
	// a sysctl segment regex, concatenated with dots to form a sysctl name
	SysctlSegmentFmt string = "[a-z0-9]([-_a-z0-9]*[a-z0-9])?"

	// a sysctl name regex with slash allowed
	SysctlContainSlashFmt string = "(" + SysctlSegmentFmt + "[\\./])*" + SysctlSegmentFmt

	// the maximal length of a sysctl name
	SysctlMaxLength int = 253
)

var sysctlContainSlashRegexp = regexp.MustCompile("^" + SysctlContainSlashFmt + "$")

// IsValidSysctlName checks that the given string is a valid sysctl name,
// i.e. matches SysctlContainSlashFmt.
// More info:
//
//	https://man7.org/linux/man-pages/man8/sysctl.8.html
//	https://man7.org/linux/man-pages/man5/sysctl.d.5.html
func IsValidSysctlName(name string) bool {
	if len(name) > SysctlMaxLength {
		return false
	}
	return sysctlContainSlashRegexp.MatchString(name)
}

// maxGMSACredentialSpecLength is the max length, in bytes, for the actual contents
// of a GMSA cred spec. In general, those shouldn't be more than a few hundred bytes,
// so we want to give plenty of room here while still providing an upper bound.
// The runAsUserName field will be used to execute the given container's entrypoint, and
// it can be formatted as "DOMAIN/USER", where the DOMAIN is optional, maxRunAsUserNameDomainLength
// is the max character length for the user's DOMAIN, and maxRunAsUserNameUserLength
// is the max character length for the USER itself. Both the DOMAIN and USER have their
// own restrictions, and more information about them can be found here:
// https://support.microsoft.com/en-us/help/909264/naming-conventions-in-active-directory-for-computers-domains-sites-and
// https://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-2000-server/bb726984(v=technet.10)
const (
	maxGMSACredentialSpecLengthInKiB = 64
	maxGMSACredentialSpecLength      = maxGMSACredentialSpecLengthInKiB * 1024
	maxRunAsUserNameDomainLength     = 256
	maxRunAsUserNameUserLength       = 104
)

var (
	// control characters are not permitted in the runAsUserName field.
	ctrlRegex = regexp.MustCompile(`[[:cntrl:]]+`)

	// a valid NetBios Domain name cannot start with a dot, has at least 1 character,
	// at most 15 characters, and it cannot the characters: \ / : * ? " < > |
	validNetBiosRegex = regexp.MustCompile(`^[^\\/:\*\?"<>|\.][^\\/:\*\?"<>|]{0,14}$`)

	// a valid DNS name contains only alphanumeric characters, dots, and dashes.
	dnsLabelFormat                 = `[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?`
	dnsSubdomainFormat             = fmt.Sprintf(`^%s(?:\.%s)*$`, dnsLabelFormat, dnsLabelFormat)
	validWindowsUserDomainDNSRegex = regexp.MustCompile(dnsSubdomainFormat)

	// a username is invalid if it contains the characters: " / \ [ ] : ; | = , + * ? < > @
	// or it contains only dots or spaces.
	invalidUserNameCharsRegex      = regexp.MustCompile(`["/\\:;|=,\+\*\?<>@\[\]]`)
	invalidUserNameDotsSpacesRegex = regexp.MustCompile(`^[\. ]+$`)
)

// ValidateCIDR validates whether a CIDR matches the conventions expected by net.ParseCIDR
func ValidateCIDR(cidr string) (*net.IPNet, error) {
	_, net, err := netutils.ParseCIDRSloppy(cidr)
	if err != nil {
		return nil, err
	}
	return net, nil
}

func IsDecremented(update, old *int32) bool {
	if update == nil && old != nil {
		return true
	}
	if update == nil || old == nil {
		return false
	}
	return *update < *old
}

// ValidateNamespace tests if required fields are set.
func ValidateNamespace(namespace *coreres.Namespace) field.ErrorList {
	allErrs := ValidateObjectMeta(&namespace.ObjectMeta, false, ValidateNamespaceName, field.NewPath("metadata"))
	for i := range namespace.Spec.Finalizers {
		allErrs = append(allErrs, validateFinalizerName(string(namespace.Spec.Finalizers[i]), field.NewPath("spec", "finalizers"))...)
	}
	return allErrs
}

// Validate finalizer names
func validateFinalizerName(stringValue string, fldPath *field.Path) field.ErrorList {
	allErrs := apimachineryvalidation.ValidateFinalizerName(stringValue, fldPath)
	allErrs = append(allErrs, validateKubeFinalizerName(stringValue, fldPath)...)
	return allErrs
}

// validateKubeFinalizerName checks for "standard" names of legacy finalizer
func validateKubeFinalizerName(stringValue string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(strings.Split(stringValue, "/")) == 1 {
		if !helper.IsStandardFinalizerName(stringValue) {
			return append(allErrs, field.Invalid(fldPath, stringValue, "name is neither a standard finalizer name nor is it fully qualified"))
		}
	}

	return allErrs
}

// ValidateNamespaceUpdate tests to make sure a namespace update can be applied.
func ValidateNamespaceUpdate(newNamespace *coreres.Namespace, oldNamespace *coreres.Namespace) field.ErrorList {
	allErrs := ValidateObjectMetaUpdate(&newNamespace.ObjectMeta, &oldNamespace.ObjectMeta, field.NewPath("metadata"))
	return allErrs
}

// ValidateNamespaceStatusUpdate tests to see if the update is legal for an end user to make.
func ValidateNamespaceStatusUpdate(newNamespace, oldNamespace *coreres.Namespace) field.ErrorList {
	allErrs := ValidateObjectMetaUpdate(&newNamespace.ObjectMeta, &oldNamespace.ObjectMeta, field.NewPath("metadata"))
	if newNamespace.DeletionTimestamp.IsZero() {
		if newNamespace.Status.Phase != coreres.NamespaceActive {
			allErrs = append(allErrs, field.Invalid(field.NewPath("status", "Phase"), newNamespace.Status.Phase, "may only be 'Active' if `deletionTimestamp` is empty"))
		}
	} else {
		if newNamespace.Status.Phase != coreres.NamespaceTerminating {
			allErrs = append(allErrs, field.Invalid(field.NewPath("status", "Phase"), newNamespace.Status.Phase, "may only be 'Terminating' if `deletionTimestamp` is not empty"))
		}
	}
	return allErrs
}

// ValidateNamespaceFinalizeUpdate tests to see if the update is legal for an end user to make.
func ValidateNamespaceFinalizeUpdate(newNamespace, oldNamespace *coreres.Namespace) field.ErrorList {
	allErrs := ValidateObjectMetaUpdate(&newNamespace.ObjectMeta, &oldNamespace.ObjectMeta, field.NewPath("metadata"))

	fldPath := field.NewPath("spec", "finalizers")
	for i := range newNamespace.Spec.Finalizers {
		idxPath := fldPath.Index(i)
		allErrs = append(allErrs, validateFinalizerName(string(newNamespace.Spec.Finalizers[i]), idxPath)...)
	}
	return allErrs
}

// ValidateTypedLocalObjectReference validates a TypedLocalObjectReference
func ValidateTypedLocalObjectReference(dataSource *coreres.TypedLocalObjectReference, referKind string, fldPath *field.Path, allowInvalidAPIGroupInDataSourceOrRef bool) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(dataSource.Name) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
	}
	if len(dataSource.Kind) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("kind"), ""))
	}
	apiVersion := ""
	if dataSource.APIVersion != nil {
		apiVersion = *dataSource.APIVersion
	}
	if len(apiVersion) == 0 && dataSource.Kind != referKind {
		allErrs = append(allErrs, field.Invalid(fldPath, dataSource.Kind, fmt.Sprintf("must be '%s' when referencing the default apiVersion", referKind)))
	}

	return allErrs
}

// ValidateTypedObjectReference validates a TypedObjectReference
func ValidateTypedObjectReference(dataSource *coreres.TypedObjectReference, referKind string, fldPath *field.Path, allowInvalidAPIGroupInDataSourceOrRef bool) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(dataSource.Name) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
	}
	if len(dataSource.Kind) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("kind"), ""))
	}
	apiVersion := ""
	if dataSource.APIVersion != nil {
		apiVersion = *dataSource.APIVersion
	}
	if len(apiVersion) == 0 && dataSource.Kind != referKind {
		allErrs = append(allErrs, field.Invalid(fldPath, dataSource.Kind, fmt.Sprintf("must be '%s' when referencing the default apiVersion", referKind)))
	}

	namespace := ""
	if dataSource.Namespace != nil {
		namespace = *dataSource.Namespace
	}
	if len(namespace) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("namespace"), ""))
	} else {
		for _, msg := range ValidateNamespaceName(namespace, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), namespace, msg))
		}
	}

	return allErrs
}

// ValidateLocalObjectReference validates a LocalObjectReference
func ValidateLocalObjectReference(dataSource *coreres.LocalObjectReference, fldPath *field.Path, allowInvalidAPIGroupInDataSourceOrRef bool) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(dataSource.Name) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
	}

	return allErrs
}

// ValidateObjectReference validates a ObjectReference
func ValidateObjectReference(dataSource *coreres.ObjectReference, referKind string, fldPath *field.Path, allowInvalidAPIGroupInDataSourceOrRef bool) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(dataSource.Name) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
	}
	if len(dataSource.Kind) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("kind"), ""))
	}

	if len(dataSource.UID) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("uid"), ""))
	}

	namespace := dataSource.Namespace
	if len(namespace) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("namespace"), ""))
	} else {
		for _, msg := range ValidateNamespaceName(namespace, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), namespace, msg))
		}
	}

	return allErrs
}
