/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package mongodb

import (
	"strconv"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	pluginstorage "github.com/seanchann/apimaster/plugin/storage"
)

// APIObjectVersioner implements versioning and extracting database information
// for objects that have an embedded ObjectMeta or ListMeta field.
type APIObjectVersioner struct{}

// UpdateObject implements Versioner
func (a APIObjectVersioner) UpdateObject(obj runtime.Object, resourceVersion uint64) error {
	accessor, err := meta.TypeAccessor(obj)
	if err != nil {
		return err
	}
	accessor.SetAPIVersion("v1")
	accessor.SetKind(pluginstorage.GetObjKind(obj))
	return nil
}

// UpdateList implements Versioner
func (a APIObjectVersioner) UpdateList(obj runtime.Object, resourceVersion uint64) error {
	// listMeta, err := api.ListMetaFor(obj)
	// if err != nil || listMeta == nil {
	// 	return err
	// }
	// versionString := ""
	// if resourceVersion != 0 {
	// 	versionString = strconv.FormatUint(resourceVersion, 10)
	// }
	// listMeta.ResourceVersion = versionString
	return nil
}

// ObjectResourceVersion implements Versioner
func (a APIObjectVersioner) ObjectResourceVersion(obj runtime.Object) (uint64, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return 0, err
	}
	version := accessor.GetResourceVersion()
	if len(version) == 0 {
		return 0, nil
	}
	return strconv.ParseUint(version, 10, 64)
}
