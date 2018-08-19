/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package storage

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
)

func CloneRuntimeObj(objPtr runtime.Object) (runtime.Object, error) {
	v, err := conversion.EnforcePtr(objPtr)
	if err != nil {
		return nil, err
	}

	newObj := reflect.New(v.Type())
	storeObj := newObj.Interface().(runtime.Object)
	return storeObj, nil
}

func GetObjKind(objPtr runtime.Object) string {
	v, err := conversion.EnforcePtr(objPtr)
	if err != nil {
		return string("")
	}

	kind := v.Type().String()
	if i := strings.IndexAny(kind, "."); i >= 0 {
		kind = kind[i+1:]
	}
	return kind
}

func GetListItemObj(listObj runtime.Object) (listPtr interface{}, itemObj runtime.Object, err error) {
	listPtr, err = meta.GetItemsPtr(listObj)
	if err != nil {
		return
	}

	items, err := conversion.EnforcePtr(listPtr)
	if err != nil {
		return
	}
	if items.Kind() != reflect.Slice {
		err = fmt.Errorf("object(%v) not a slice", items.Kind())
		return
	}

	itemObj = reflect.New(items.Type().Elem()).Interface().(runtime.Object)

	return
}
