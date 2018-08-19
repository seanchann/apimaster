/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package mysql

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	pluginstorage "github.com/seanchann/apimaster/plugin/storage"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apiserver/pkg/storage"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
)

const (
	//tableTagKeyWordResourceKey indicate field is resrouce key.
	//append attr for strcut field.
	tableTagKeyWordResourceKey = "resoucekey"
	//tableTagKeyWordConstant indicate field is constant.it mean not update this filed
	tableTagKeyWordConstant = "const"

	//tableTagKeyWordDefaultValue if keyword not given a value. set default value
	tableTagKeyWordDefaultValue = "yes"
)

const (
	queryAllField = "*"
	queryCount    = "count(*)"
)

const (
	freezerTag = "freezer"
	jsonTag    = "json"

	//StructTagKey struct tag key for mysql
	tagColumn = "column"
	//tagTableKey table name in tag
	tagTable = "table"
	//primary_key tag
	//tagResourceKey = "resoucekey"

	//columnRawObj contains a column that name is rawobj.  store runtime.Object json into it
	columnRawObj = "rawobj"
	//jsonTagRawObj always set rawobj filed jsontag with this value
	jsonTagRawObj = "rawobj"
)

//example struct
/*
type DBRoot struct {
	embedded DBResource `freezer:"table:dbresource"`
}

type DBResource struct {
	//use column as extend, because of use gorm, so append gorm tag(gorm and sql tag)
	Name string `freezer:"column:name;resoucekey" gorm:"column:name" sql:"type:varchar(100);unique"`
}

you must give out a resource key for rest requst
*/

//TableTag indicate table information in current field
type TableTag struct {
	column      string            //it is a column name in db
	tableName   string            //a table name if this field with table tag
	keyword     map[string]string //the keyword for sql like as resourceKey unique and so on
	structField string
}

//Table extract table from object
type Table struct {
	name string
	obj  reflect.Value

	//json tag value as key,valus is TableTag: it contains column or talbe or  other keyword for sql
	freezerTag map[string]TableTag

	//freezer column as key,the jsonkey as value
	columnToFreezerTagKey map[string]string

	//resourcekey hold a column name,this key as a resouce name in restful url
	//will use this filed value for metadata.name field
	resoucekey string
}

const (
	maxTableCache int = 128
)

var tableCache *cache.LRUExpireCache

func init() {
	tableCache = cache.NewLRUExpireCache(maxTableCache)
}

//FindTableTag scan object field,extract it into TableTag
func FindTableTag(typ reflect.Type, index int, t *Table) bool {
	field := typ.Field(index)

	value, ok := field.Tag.Lookup(freezerTag)
	findTable := false

	if ok {
		tagMap := parseTag(value)

		//specific process for rawobj, column and freezer tag value is rawobj.
		//but json tag may be -
		if len(tagMap.column) != 0 && 0 == strings.Compare(tagMap.column, columnRawObj) {
			if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Uint8 {
				tagMap.structField = field.Name
				t.freezerTag[jsonTagRawObj] = tagMap
				t.columnToFreezerTagKey[tagMap.column] = jsonTagRawObj
				goto out
			}
		}

		var jsonKey string
		if value, ok := field.Tag.Lookup(jsonTag); ok {
			jsonKey = stripJSONTagValue(value)
			tagMap.structField = field.Name
			t.freezerTag[jsonKey] = tagMap
		}

		if len(tagMap.tableName) != 0 {
			t.columnToFreezerTagKey[tagTable] = jsonKey
			t.name = tagMap.tableName
			findTable = true
		} else if len(tagMap.column) != 0 {
			t.columnToFreezerTagKey[tagMap.column] = jsonKey
		}

		_, ok := tagMap.keyword[tableTagKeyWordResourceKey]
		if ok {
			t.resoucekey = tagMap.column
		}
		// for _, v := range tagMap.keyword {
		// 	if strings.Compare(tableTagKeyWordResourceKey, v) == 0 {
		// 		t.resoucekey = tagMap.column
		// 	}
		// }

	}

out:
	return findTable
}

//BuildTable search tag in obj
//return the reflect.value of tag
//return error if has a error
func BuildTable(obj reflect.Value, t *Table) error {

	vType := obj.Type()
	for i := 0; i < vType.NumField(); i++ {
		embV := obj.Field(i)

		if FindTableTag(vType, i, t) {
			t.obj = reflect.Indirect(reflect.New(embV.Type()))
			t.obj.Set(embV)
		}

		switch embV.Kind() {
		case reflect.Struct:
			if err := BuildTable(embV, t); err != nil {
				return err
			}
		}
	}

	return nil
}

func stripJSONTagValue(origin string) string {
	vals := strings.Split(origin, ",")
	return vals[0]
}

func parseTag(origin string) TableTag {
	vals := strings.Split(origin, ";")

	tag := TableTag{
		keyword: make(map[string]string),
	}
	for _, query := range vals {
		key := query
		if i := strings.IndexAny(key, ":"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else if len(key) != 0 {
			singleTag := key[:]
			key, query = singleTag, singleTag
		} else {
			key, query = "", ""
		}

		switch key {
		case tagTable:
			tag.tableName = query
		case tagColumn:
			tag.column = query
		case tableTagKeyWordResourceKey:
			//we used column value in  this tag line
			tag.keyword[query] = tableTagKeyWordDefaultValue
		case tableTagKeyWordConstant:
			tag.keyword[query] = tableTagKeyWordDefaultValue
		}
	}

	return tag
}

//GetTable scan obj tags and get table in obj if exist
func GetTable(ctx context.Context, obj runtime.Object) (*Table, error) {
	v, err := conversion.EnforcePtr(obj)
	if err != nil {
		return nil, err
	}

	kind := pluginstorage.GetObjKind(obj)
	cacheVal, found := tableCache.Get(kind)
	if found {
		table := cacheVal.(*Table)
		return table, nil
	}

	table := &Table{
		freezerTag:            make(map[string]TableTag),
		columnToFreezerTagKey: make(map[string]string),
	}

	err = BuildTable(v, table)
	if err != nil {
		return nil, err
	}

	if len(table.name) == 0 {
		return nil, fmt.Errorf("not find tag('table') in struct")
	}

	if len(table.resoucekey) == 0 {
		return nil, fmt.Errorf("not find resource key in struct")
	}

	_, ok := table.freezerTag[jsonTagRawObj]
	if !ok {
		return nil, fmt.Errorf("in %v tableTag(%v) not find rawobj field in struct or rawobj type is not a []byte", table.name, table.freezerTag)
	}

	glog.V(5).Infof("find table %+v", table)
	tableCache.Add(kind, table, 24*time.Hour)

	WithTable(ctx, table)

	return table, err
}

//ObjSelectField return talbe column array
func (t *Table) ObjSelectField(tableObj reflect.Value) []interface{} {
	selects := []interface{}{}

	for _, v := range t.freezerTag {
		selects = append(selects, v.column)
	}

	return selects
}

//ObjMapField convert obj field into map by filed,if field is nil return all field
//if ignoreConstField==true. will ignore filed if it has const keyword
func (t *Table) ObjMapField(obj reflect.Value, field []string, ignoreConstField bool) map[string]interface{} {
	update := make(map[string]interface{})
	if len(field) == 0 {
		for _, v := range t.freezerTag {
			_, ok := v.keyword[tableTagKeyWordConstant]
			if ok {
				continue
			}

			itemVal := obj.FieldByName(v.structField)
			if itemVal.IsValid() && itemVal.CanInterface() {
				update[v.column] = itemVal.Interface()
			}
		}
	} else {
		//TODO field is struct filed?
		for _, v := range field {
			itemVal := obj.FieldByName(v)
			if itemVal.IsValid() && itemVal.CanInterface() {
				update[v] = itemVal.Interface()
			}
		}
	}

	return update
}

//AfterFindTable find tableObj in object then call this function
type AfterFindTable func(tableObj reflect.Value) error

//ExtractTableObj extract table field in obj. passthrough tableObj with afterFunc
func (t *Table) ExtractTableObj(obj runtime.Object, afterFunc AfterFindTable) error {
	//tObj := reflect.Value{}
	//var tObj uintptr
	v, err := conversion.EnforcePtr(obj)
	if err != nil {
		return err
	}

	if !v.CanInterface() {
		return fmt.Errorf("object(%v) cannt interface by reflect", v.Type())
	}

	//find in struct root
	tableFiledName := t.freezerTag[t.columnToFreezerTagKey[tagTable]].structField
	val := v.FieldByName(tableFiledName)
	if val.IsValid() {
		return afterFunc(val)
	}

	vType := v.Type()
	for i := 0; i < vType.NumField(); i++ {
		embV := v.Field(i)

		switch embV.Kind() {
		case reflect.Struct:
			val := embV.FieldByName(tableFiledName)
			if val.IsValid() {
				return afterFunc(val)
			}
			vType = embV.Type()
		}
	}

	return fmt.Errorf("runtime object(%v) not found table", v.Type())
}

//SetTable find talbe field in obj. then set obj field with table value
func (t *Table) SetTable(obj runtime.Object, table reflect.Value) (string, error) {
	resourceKeyValue := string("")
	err := t.ExtractTableObj(obj, func(tObj reflect.Value) error {
		if !tObj.CanSet() {
			return fmt.Errorf("object(%v) cannt set by reflect", tObj.Type())
		}
		tObj.Set(table)
		val := tObj.FieldByName(t.freezerTag[t.columnToFreezerTagKey[t.resoucekey]].structField)
		resourceKeyValue = fmt.Sprintf("%v", val.Interface())
		return nil
	})
	if err != nil {
		return "", err
	}

	return resourceKeyValue, nil
}

//CovertRowsToObject update table(reflect.value) into obj(runtime.Object).
//and  Marshal obj  into row(RowResult)
func (t *Table) CovertRowsToObject(row *RowResult, obj runtime.Object, table reflect.Value) error {
	resourceKeyValue, err := t.SetTable(obj, table)
	if err != nil {
		return err
	}

	//row.data, err = json.Marshal(obj)
	rawobjFieldVal := table.FieldByName(t.freezerTag[jsonTagRawObj].structField).Bytes()
	row.data = make([]byte, len(rawobjFieldVal))
	copy(row.data[:], rawobjFieldVal[:])
	// row.data, err = json.Marshal(rawobjFieldVal)
	//
	// if err != nil {
	// 	return err
	// }
	row.resourceKey = resourceKeyValue

	return nil
}

//GetColumnByField get sql column name by struct filed name
func (t *Table) GetColumnByField(filed string) (column string) {
	i := strings.LastIndexAny(filed, ".")
	if i >= 0 {
		column = filed[i+1:]
	} else {
		column = filed
	}

	//check is a invalid column
	_, ok := t.freezerTag[column]
	if !ok {
		column = ""
	} else {
		tag := t.freezerTag[column]
		column = tag.column
	}

	return
}

//ConvertFieldsValue convert struct value to sqlvale
func (t *Table) ConvertFieldsValue(value string) (sqlValue string) {

	switch value {
	case "true":
		sqlValue = fmt.Sprintf("1")
	case "false":
		sqlValue = fmt.Sprintf("0")
	default:
		sqlValue = value
	}

	return
}

//appendQuoteToField append quote into filed for every member。
// eg 'spec.test.name' ====> '"spec"."test"."name"'
func appendQuoteToField(input string) (field string) {
	members := strings.Split(input, ".")
	if len(members) > 0 {
		var quoteMember []string
		for _, v := range members {
			m := fmt.Sprintf("\"%s\"", v)
			quoteMember = append(quoteMember, m)
		}
		field = strings.Join(quoteMember, ".")
	}

	return
}

//Fields build gorm select condition by storage.SelectionPredicate
//selectionFeild contains what field will be select for query
func (t *Table) Fields(dbHandle *gorm.DB, p storage.SelectionPredicate, selectionFeild []string) *gorm.DB {
	if p.Field == nil || (p.Field != nil && p.Field.Empty()) {
		return dbHandle
	}

	selectAllField := false
	selectCountField := false
	for _, v := range selectionFeild {
		if strings.Compare(v, queryAllField) == 0 {
			selectAllField = true
		} else if strings.Compare(v, queryCount) == 0 {
			selectCountField = true
		}
	}
	_ = selectAllField

	fieldsCondition := p.Field.Requirements()
	for _, v := range fieldsCondition {
		column := t.GetColumnByField(v.Field)
		switch v.Operator {
		case selection.Equals:
			fallthrough
		case selection.DoubleEquals:
			if column == "" {
				query := fmt.Sprintf("JSON_CONTAINS(%s, '%s', '$.%s') = ?",
					t.freezerTag[jsonTagRawObj].column, v.Value, appendQuoteToField(v.Field))
				queryArgs := "1"

				dbHandle = dbHandle.Where(query, queryArgs)
			} else {
				query := fmt.Sprintf("%s = ?", column)
				queryArgs := t.ConvertFieldsValue(v.Value)
				dbHandle = dbHandle.Where(query, queryArgs)
			}
		case selection.NotEquals:
			if column == "" {
				query := fmt.Sprintf("JSON_CONTAINS(%s, '%s', '$.%s') = ?",
					t.freezerTag[jsonTagRawObj].column, v.Value, appendQuoteToField(v.Field))
				queryArgs := "0"

				dbHandle = dbHandle.Where(query, queryArgs)
			} else {
				query := fmt.Sprintf("%s != ?", column)
				queryArgs := t.ConvertFieldsValue(v.Value)
				dbHandle = dbHandle.Where(query, queryArgs)
			}

		//TODO: this can't be support if have specific Requirements
		case selection.In:
			fallthrough
		case selection.NotIn:
			glog.Warningf("strange operator. check value in field in talbe(%v)", t.name)
		case selection.DoesNotExist:
			fallthrough
		case selection.Exists:
			//only search rawobj filed for this operator
			if selectCountField {
				continue
			}
			if column == "" {
				query := fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.%s') = ?",
					t.freezerTag[jsonTagRawObj].column, appendQuoteToField(v.Field))
				queryArgs := "1"
				if selection.DoesNotExist == v.Operator {
					queryArgs = "0"
				}

				dbHandle = dbHandle.Where(query, queryArgs)
			}
		}
	}

	return dbHandle
}

//BaseCondition build select condition
func (t *Table) BaseCondition(dbHandle *gorm.DB, p storage.SelectionPredicate, selectionFeild []string) *gorm.DB {
	dbHandle = t.Fields(dbHandle, p, selectionFeild)
	dbHandle = dbHandle.Order(t.resoucekey)

	return dbHandle
}

//PageCondition build gorm selection by PageCondition
func (t *Table) PageCondition(dbHandle *gorm.DB, p storage.SelectionPredicate, totalCount uint64) *gorm.DB {

	hasPage, perPage, skip := p.BuildPagerCondition(uint64(totalCount))
	if hasPage {
		limitVal := perPage
		if limitVal != 0 {
			dbHandle = dbHandle.Limit(int(limitVal))
		}

		skipVal := skip
		if skipVal != 0 {
			dbHandle = dbHandle.Offset(int(skipVal))
		}
	}

	return dbHandle
}
