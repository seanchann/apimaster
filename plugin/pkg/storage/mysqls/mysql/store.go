/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package mysql

import (
	"fmt"
	"gofreezer/pkg/fields"
	"reflect"

	"github.com/seanchann/apimaster/plugin/mysqls"
	pluginstorage "github.com/seanchann/apimaster/plugin/storage"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage"

	"k8s.io/klog"
	dbmysql "github.com/jinzhu/gorm"
	"golang.org/x/net/context"
)

const (
	//give a resourceversion with 1 if resource exist
	resourceVersion = 1
)

type store struct {
	client         *dbmysql.DB
	codec          runtime.Codec
	versioner      APIObjectVersioner
	storageVersion string
}

type RowResult struct {
	data        []byte
	resourceKey string
}

//New create a mysql store
func New(client *dbmysql.DB, codec runtime.Codec, version string) *store {
	versioner := APIObjectVersioner{}
	if len(version) == 0 {
		klog.Fatalln("need give a storage version for mysql backend")
	}
	return &store{
		client:         client,
		codec:          codec,
		versioner:      versioner,
		storageVersion: version,
	}
}

const (
	tablecontextKey = iota
)

func (s *store) Type() string {
	return string("mysql")
}

// Versioner implements storage.Interface.Versioner.
func (s *store) Versioner() storage.Versioner {
	return s.versioner
}

func (s *store) Create(ctx context.Context, key string, obj, out runtime.Object) error {
	err := s.GetResourceWithKey(ctx, key, out, true)
	if err != nil {
		return err
	}
	table := s.table(ctx, out)

	err = table.ExtractTableObj(obj, func(tObj reflect.Value) error {

		data, err := runtime.Encode(s.codec, obj)
		if err != nil {
			return storage.NewInternalErrorf("key %v, object encode error %v", key, err.Error())
		}
		rawObjField := table.freezerTag[jsonTagRawObj].structField
		tObj.FieldByName(rawObjField).SetBytes(data)

		klog.V(9).Infof("insert into %s with obj (%+v)  ", table.name, tObj)
		err = s.client.Table(table.name).Create(tObj.Addr().Interface()).Error
		if err != nil {
			return storage.NewInternalErrorf(key, err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}

	return s.GetResourceWithKey(ctx, key, out, false)
}

func (s *store) Delete(ctx context.Context, key string, out runtime.Object, preconditions *storage.Preconditions) error {

	err := s.GetResourceWithKey(ctx, key, out, false)
	if err != nil {
		return err
	}
	table := s.table(ctx, out)

	delObj := reflect.New(table.obj.Type())

	query := fmt.Sprintf("%s = ?", table.resoucekey)
	args := GetActualResourceKey(key)
	err = s.client.Table(table.name).Where(query, args).Delete(delObj).Error
	if err != nil {
		return storage.NewInternalErrorf(key, err.Error())
	}

	return err
}

func (s *store) Get(ctx context.Context, key string, objPtr runtime.Object, ignoreNotFound bool) error {

	return s.GetResourceWithKey(ctx, key, objPtr, ignoreNotFound)
}

func (s *store) GetToList(ctx context.Context, key string, p storage.SelectionPredicate, listObj runtime.Object) error {

	listPtr, itemPtrObj, err := pluginstorage.GetListItemObj(listObj)
	if err != nil {
		return storage.NewInvalidObjError(key, err.Error())
	}

	rowList, _, err := s.doQuery(ctx, key, itemPtrObj, p)
	if err != nil {
		return err
	}

	if len(rowList) == 0 {
		return nil
	}

	if err := decodeList(rowList, listPtr, s.codec, s.versioner); err != nil {
		return err
	}
	return nil
}

func (s *store) GuaranteedUpdate(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool,
	precondtions *storage.Preconditions, tryUpdate mysqls.UpdateFunc, suggestion ...runtime.Object) error {

	exist := true
	err := s.GetResourceWithKey(ctx, key, out, false)
	if err != nil {
		if storage.IsNotFound(err) {
			klog.V(9).Infof("item not found check if allow create on update(%v)\r\n", err)
			exist = false
		} else {
			return err
		}
	}
	table := s.table(ctx, out)

	ret, fields, err := userUpdate(out, tryUpdate)
	if err != nil {
		klog.V(9).Infof("user update error :%v\r\n", err)
		return storage.NewInternalErrorf("key %s error:%v", key, err.Error())
	}

	if exist {
		//build update fields
		update := make(map[string]interface{})

		query := fmt.Sprintf("%s = ?", table.resoucekey)
		args := GetActualResourceKey(key)
		dbhandler := s.client.Table(table.name).Where(query, args)
		err = table.ExtractTableObj(ret, func(obj reflect.Value) error {
			//update all fields
			if len(fields) == 0 {
				//encode object update rawobj filed
				data, err := runtime.Encode(s.codec, ret)
				if err != nil {
					return storage.NewInternalErrorf("key %v, object encode error %v", key, err.Error())
				}
				rawObjField := table.freezerTag[jsonTagRawObj].structField
				obj.FieldByName(rawObjField).SetBytes(data)

				upMap := table.ObjMapField(obj, nil, true)
				dbhandler = dbhandler.Updates(upMap)
			} else {
				for _, v := range fields {
					update[v] = obj.FieldByName(v).Interface()
				}
				dbhandler = dbhandler.Updates(update)
			}
			return nil
		})
		if err != nil {
			return storage.NewInternalErrorf("key %s error:%v", key, err.Error())
		}

		err = dbhandler.Error
		if err != nil {
			return storage.NewInternalErrorf("key %s error:%v", key, err.Error())
		}
		return s.GetResourceWithKey(ctx, key, out, false)
	}

	klog.V(9).Infof("Create obj for update method, obj: %+v\r\n", ret)

	return s.Create(ctx, key, ret, out)

}

func (s *store) doQuery(ctx context.Context, key string, objPtr runtime.Object, p storage.SelectionPredicate) ([]*RowResult, *Table, error) {

	table := s.table(ctx, objPtr)

	//get count
	var count uint64
	err := s.GetCount(ctx, key, objPtr, p, &count)
	if err != nil {
		return nil, nil, err
	}
	if count == 0 {
		return nil, table, nil
	}

	dbHandle := s.client.Table(table.name).Model(table.obj.Type())
	selectionField := []string{queryAllField}
	dbHandle = table.BaseCondition(dbHandle, p, selectionField)
	dbHandle = table.PageCondition(dbHandle, p, count)

	rows, err := dbHandle.Rows()
	if err != nil {
		return nil, nil, storage.NewInternalErrorf("key %s error:%v", key, err.Error())
	}
	defer rows.Close()

	cloneObj, err := pluginstorage.CloneRuntimeObj(objPtr)
	if err != nil {
		return nil, nil, storage.NewInternalErrorf("key %s error: %v", key, err.Error())
	}

	s.versioner.UpdateTypeMeta(cloneObj, pluginstorage.GetObjKind(cloneObj), s.storageVersion)

	rowList, err := ScanRows(rows, table, cloneObj)
	if err != nil {
		return nil, nil, storage.NewInternalErrorf("key %s error:%v", key, err.Error())
	}
	return rowList, table, err
}

func userUpdate(input runtime.Object, userUpdate mysqls.UpdateFunc) (runtime.Object, []string, error) {
	ret, fields, err := userUpdate(input)
	if err != nil {
		return nil, nil, err
	}

	return ret, fields, nil
}

// decode decodes value of bytes into object. It will also set the object resource version to rev.
// On success, objPtr would be set to the object.
func decode(codec runtime.Codec, versioner storage.Versioner, elem *RowResult, objPtr runtime.Object) error {
	if _, err := conversion.EnforcePtr(objPtr); err != nil {
		panic("unable to convert output object to pointer")
	}
	_, _, err := codec.Decode(elem.data, nil, objPtr)
	if err != nil {
		return err
	}
	// being unable to set the version does not prevent the object from being extracted
	versioner.UpdateObject(objPtr, uint64(resourceVersion))
	UpdateNameWithResouceKey(objPtr, elem.resourceKey)
	return nil
}

// decodeList decodes a list of values into a list of objects
// On success, ListPtr would be set to the list of objects.
func decodeList(elems []*RowResult, ListPtr interface{}, codec runtime.Codec, versioner storage.Versioner) error {
	v, err := conversion.EnforcePtr(ListPtr)
	if err != nil || v.Kind() != reflect.Slice {
		panic("need ptr to slice")
	}
	for _, elem := range elems {
		obj, _, err := codec.Decode(elem.data, nil, reflect.New(v.Type().Elem()).Interface().(runtime.Object))
		if err != nil {
			return err
		}
		// being unable to set the version does not prevent the object from being extracted
		versioner.UpdateObject(obj, resourceVersion)
		UpdateNameWithResouceKey(obj, elem.resourceKey)
		v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
	}
	return nil
}

//filter support query arg
func (s *store) GetCount(ctx context.Context, key string, objPtr runtime.Object, p storage.SelectionPredicate, result *uint64) error {

	table := s.table(ctx, objPtr)

	dbHandle := s.client.Table(table.name).Model(table.obj.Type())

	selectionField := []string{queryCount}
	dbHandle = table.BaseCondition(dbHandle, p, selectionField)
	err := dbHandle.Count(result).Error
	if err != nil {
		return storage.NewInternalErrorf("key %v, query count error %v", key, err)
	}

	return nil
}

//GetResourceWithKey build a sql request with key
func (s *store) GetResourceWithKey(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool) error {

	table := s.table(ctx, out)

	klog.V(9).Infof("Get resource with key %s", key)

	p := storage.SelectionPredicate{}
	resourceField := table.columnToFreezerTagKey[table.resoucekey]
	p.Field = fields.SelectorFromSet(map[string]string{resourceField: GetActualResourceKey(key)})

	rowList, _, err := s.doQuery(ctx, key, out, p)
	if err != nil {
		return storage.NewUnreachableError(key, resourceVersion)
	}

	if len(rowList) == 0 {
		if ignoreNotFound {
			return runtime.SetZeroValue(out)
		}
		return storage.NewKeyNotFoundError(key, resourceVersion)
	} else if len(rowList) > 1 {
		panic(fmt.Sprintf("resource key(%s) must to be unique", key))
	}

	if err := decode(s.codec, s.versioner, rowList[0], out); err != nil {
		return err
	}

	return nil
}

func (s *store) table(ctx context.Context, obj runtime.Object) *Table {
	ctxTable, ok := ctx.Value(tablecontextKey).(*Table)
	if !ok {
		table, err := GetTable(ctx, obj)
		if err != nil {
			panic(fmt.Sprintf("struct must to be as a table. error(%v).", err))
		}
		return table
	}

	return ctxTable
}
