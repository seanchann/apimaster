/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/seachann/apimaster/plugin/storage/awsdynamodb"
	pluginstorage "github.com/seanchann/apimaster/plugin/storage"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsdb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/golang/glog"
)

const (
	//give a resourceversion with 1 if resource exist
	resourceVersion = 1
)

type store struct {
	codec     runtime.Codec
	versioner APIObjectVersioner
	dbHandler *awsdb.DynamoDB
	table     string
}

//New create a mongo store
func New(sess *session.Session, table string, codec runtime.Codec) *store {
	versioner := APIObjectVersioner{}
	db := awsdb.New(sess)

	if len(table) == 0 {
		table = defaultTable
	}
	desc, err := CreateTable(db, table)
	if err != nil {
		glog.Fatalf("table not active, error : %v", err)
		return nil
	}
	glog.V(5).Infof("Got table(%v) description: %v", table, desc)

	return &store{
		codec:     codec,
		versioner: versioner,
		dbHandler: db,
		table:     table,
	}
}

func (s *store) Type() string {
	return string("dynamodb")
}

// Versioner implements storage.Interface.Versioner.
func (s *store) Versioner() storage.Versioner {
	return s.versioner
}

func (s *store) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	glog.V(9).Infof("dynamodb create resource  %v \r\n", key)

	//check item with this key if exist
	_, err := s.queryObjByKey(key, out, true)
	if err != nil {
		return storage.NewInternalErrorf("key %v, object search error %v", err.Error())
	}

	data, err := runtime.Encode(s.codec, obj)
	if err != nil {
		return storage.NewInternalErrorf("key %v, object encode error %v", key, err.Error())
	}

	mapObj, err := ConvertByteToMap(data)
	if err != nil {
		return storage.NewInternalErrorf("key %v, object encode error %v", key, err.Error())
	}

	item, err := dynamodbattribute.MarshalMap(mapObj)
	if err != nil {
		return err
	}
	item[primaryKey] = &awsdb.AttributeValue{S: aws.String(key)}
	item[sortKey] = &awsdb.AttributeValue{S: aws.String(key)}

	_, err = s.dbHandler.PutItem(&awsdb.PutItemInput{
		Item:      item,
		TableName: aws.String(s.table),
		//ReturnValues: aws.String("ALL_OLD"),
	})
	if err != nil {
		return storage.NewInternalErrorf("key %v, put error %v\r\n", key, err)
	}

	return decode(s.codec, s.versioner, data, out)
}

func (s *store) Delete(ctx context.Context, key string, out runtime.Object, preconditions *storage.Preconditions) error {
	glog.V(9).Infof("dynamodb delete resource  %v \r\n", key)
	_, err := s.queryObjByKey(key, out, false)
	if err != nil {
		return err
	}

	params := &awsdb.DeleteItemInput{
		Key: map[string]*awsdb.AttributeValue{
			primaryKey: &awsdb.AttributeValue{
				S: aws.String(key),
			},
			sortKey: &awsdb.AttributeValue{
				S: aws.String(key),
			},
		},
		ReturnValues: aws.String("ALL_OLD"),
		TableName:    aws.String(s.table),
	}

	resp, err := s.dbHandler.DeleteItem(params)
	if err != nil {
		return storage.NewInternalErrorf("key %v delete error %v\r\n", err.Error())
	}
	glog.V(9).Infof("got result %v err %v\r\n", resp.Attributes, err)

	return s.getObject(key, out, false, resp.Attributes)
}

func (s *store) Get(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool) error {
	_, err := s.queryObjByKey(key, out, ignoreNotFound)
	return err
}

func (s *store) GetToList(ctx context.Context, key string, p storage.SelectionPredicate, listObj runtime.Object) error {
	listPtr, itemPtr, err := pluginstorage.GetListItemObj(listObj)
	if err != nil {
		return storage.NewInvalidObjError(key, err.Error())
	}

	//always need to get list count,
	//prevent return a lots of item from backend
	scanParam := &awsdb.ScanInput{
		TableName: aws.String(s.table),
		Select:    aws.String("COUNT"),
	}
	output, err := s.dbHandler.Scan(scanParam)
	if err != nil {
		return storage.NewInternalErrorf("key %v, scan list count error %v", err.Error())
	}

	//start scan item
	scanParam = &awsdb.ScanInput{
		TableName: aws.String(s.table),
	}
	hasPage, perPage, skip := p.BuildPagerCondition(uint64(*output.Count))
	var requirePage uint64 = 1
	if hasPage && perPage != 0 {
		limit := int64(perPage)
		scanParam.Limit = &(limit)
		requirePage = skip/perPage + 1
		glog.V(9).Infof("require pagination limit(%v) skip(%v) perpage(%v) require page(%v)\r\n", limit, skip, perPage, requirePage)
	}

	filter, expressionAttrName, expressionAttrValue := BuildScanFilterAttr(itemPtr, p)

	if len(filter) > 0 {
		scanParam.FilterExpression = aws.String(filter)
	}
	if len(expressionAttrName) > 0 {
		scanParam.ExpressionAttributeNames = expressionAttrName
	}
	if len(expressionAttrValue) > 0 {
		scanParam.ExpressionAttributeValues = expressionAttrValue
	}

	// output, err = s.dbHandler.Scan(scanParam)
	// if err != nil {
	// 	return storage.NewInternalErrorf("key %v, scan list  error %v", err.Error())
	// }
	var pageNum uint64 = 0
	err = s.dbHandler.ScanPages(scanParam, func(page *awsdb.ScanOutput, last bool) bool {
		pageNum++
		glog.V(9).Infof("Get page number %v require page %v\r\n", pageNum, requirePage)
		if pageNum == requirePage {
			output = page
			return false
		}
		if last {
			panic("until last page not found require page\r\n")
		}
		return true
	})
	if err != nil {
		return storage.NewInternalErrorf("key %v, scan list  error %v", err.Error())
	}

	glog.V(9).Infof("Get query output count %+v\r\n", *output.Count)

	jsonData, cnt, err := ConvertTOJson(&output.Items)
	if cnt == 0 {
		return nil
	}

	return decodeList(jsonData, listPtr, s.codec, s.versioner)
}

func (s *store) GuaranteedUpdate(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool, precondtions *storage.Preconditions, tryUpdate awsdynamodb.UpdateFunc) error {
	//check item with this key exist,we need replace this by aws PutItem
	_, err := s.queryObjByKey(key, out, false)
	if err != nil {
		if !storage.IsNotFound(err) {
			return storage.NewInternalErrorf("key %s, search error %v", key, err.Error())
		}
	}

	attrValue := make(map[string]interface{})
	ret, _, err := userUpdate(out, tryUpdate, attrValue)
	if err != nil {
		return storage.NewInternalErrorf("key %s, update by user error:%v", key, err.Error())
	}

	data, err := runtime.Encode(s.codec, ret)
	if err != nil {
		return storage.NewInternalErrorf("key %v, object encode error %v", key, err.Error())
	}

	mapObj, err := ConvertByteToMap(data)
	if err != nil {
		return storage.NewInternalErrorf("key %v, object encode error %v", key, err.Error())
	}

	item, err := dynamodbattribute.MarshalMap(mapObj)
	if err != nil {
		return err
	}
	item[primaryKey] = &awsdb.AttributeValue{S: aws.String(key)}
	item[sortKey] = &awsdb.AttributeValue{S: aws.String(key)}

	_, err = s.dbHandler.PutItem(&awsdb.PutItemInput{
		Item:      item,
		TableName: aws.String(s.table),
		//ReturnValues: aws.String("ALL_OLD"),
	})
	if err != nil {
		return storage.NewInternalErrorf("key %v, put error %v\r\n", key, err)
	}

	return decode(s.codec, s.versioner, data, out)
}

//can't got attr values from apistack registry store use putitem instead of
// func (s *store) GuaranteedUpdate(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool, precondtions *storage.Preconditions, tryUpdate awsdynamodb.UpdateFunc) error {
// 	//check item with this key exist,we need replace this by aws PutItem
// 	_, err := s.queryObjByKey(key, out, false)
// 	if err != nil {
// 		return storage.NewInternalErrorf("key %s, search error %v", err.Error())
// 	}
//
// 	attrValue := make(map[string]interface{})
// 	ret, _, err := userUpdate(out, tryUpdate, attrValue)
// 	if err != nil {
// 		return storage.NewInternalErrorf("key %s, update by user error:%v", key, err.Error())
// 	}
//
// 	updateExpression, expressionAttributeNames, expressionAttributeValues, err := BuildUpdateAttr(ret, out, attrValue)
// 	if err != nil {
// 		return storage.NewInternalErrorf("key %v, update error %v\r\n", key, err.Error())
// 	}
// 	glog.V(5).Infof("build attr UpdateExpression: %v expressionAttributeNames:%v expressionAttributeValues:%v",
// 		updateExpression, expressionAttributeNames, expressionAttributeValues)
//
// 	params := &awsdb.UpdateItemInput{
// 		Key: map[string]*awsdb.AttributeValue{
// 			"key": &awsdb.AttributeValue{
// 				S: aws.String(key),
// 			},
// 		},
// 		ReturnValues:              aws.String("ALL_NEW"),
// 		TableName:                 aws.String(s.table),
// 		UpdateExpression:          aws.String(updateExpression),
// 		ExpressionAttributeNames:  expressionAttributeNames,
// 		ExpressionAttributeValues: expressionAttributeValues,
// 	}
// 	resp, err := s.dbHandler.UpdateItem(params)
// 	glog.V(5).Infof("got result %v err %v\r\n", resp.Attributes, err)
// 	if err != nil {
// 		return storage.NewInternalErrorf("key %v update error %v\r\n", err.Error())
// 	}
//
// 	return s.getObject(key, out, false, resp.Attributes)
// }

// decode decodes value of bytes into object. It will also set the object resource version to rev.
// On success, objPtr would be set to the object.
func decode(codec runtime.Codec, versioner storage.Versioner, value []byte, objPtr runtime.Object) error {
	if _, err := conversion.EnforcePtr(objPtr); err != nil {
		panic("unable to convert output object to pointer")
	}
	_, _, err := codec.Decode(value, nil, objPtr)
	if err != nil {
		return err
	}
	// being unable to set the version does not prevent the object from being extracted
	versioner.UpdateObject(objPtr, uint64(resourceVersion))
	return nil
}

// decodeList decodes a list of values into a list of objects, with resource version set to corresponding rev.
// On success, ListPtr would be set to the list of objects.
func decodeList(elems []map[string]interface{}, ListPtr interface{}, codec runtime.Codec, versioner storage.Versioner) error {
	v, err := conversion.EnforcePtr(ListPtr)
	if err != nil || v.Kind() != reflect.Slice {
		panic("need ptr to slice")
	}
	for _, elem := range elems {
		data, err := json.Marshal(elem)
		if err != nil {
			return storage.NewInternalError(err.Error())
		}
		obj, _, err := codec.Decode(data, nil, reflect.New(v.Type().Elem()).Interface().(runtime.Object))
		if err != nil {
			return err
		}
		// being unable to set the version does not prevent the object from being extracted
		versioner.UpdateObject(obj, uint64(resourceVersion))
		v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
	}
	return nil
}

func userUpdate(input runtime.Object, userUpdate awsdynamodb.UpdateFunc, attributeValues map[string]interface{}) (output runtime.Object, ttl *uint64, err error) {
	ret, ttl, err := userUpdate(input, attributeValues)
	if err != nil {
		return nil, nil, err
	}
	return ret, ttl, nil
}

func (s *store) queryObjByKey(key string, out runtime.Object, ignoreNotFound bool) (*awsdb.ScanOutput, error) {
	scanParam := &awsdb.ScanInput{
		ScanFilter: map[string]*awsdb.Condition{
			"key": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*awsdb.AttributeValue{
					{
						S: aws.String(key),
					},
				},
			},
		},
		TableName: aws.String(s.table),
	}

	output, err := s.dbHandler.Scan(scanParam)
	if err != nil {
		return nil, err
	}

	return output, s.getObject(key, out, ignoreNotFound, output.Items...)
}

func (s *store) getObject(key string, out runtime.Object, ignoreNotFound bool, attrs ...map[string]*awsdb.AttributeValue) error {

	jsonData, count, err := ConvertTOJson(&attrs)
	if count == 0 {
		if ignoreNotFound {
			return runtime.SetZeroValue(out)
		}
		return storage.NewKeyNotFoundError(key, resourceVersion)
	} else if count > 1 {
		panic(fmt.Sprintf("resource key(%s) must to be unique", key))
	}

	firstObj := jsonData[0]
	data, err := json.Marshal(firstObj)
	if err != nil {
		return fmt.Errorf("marshal object(%+v) to json error:%v", firstObj, err)
	}
	//glog.V(9).Infof("marshal obj to json: %v\r\n", string(data))
	return decode(s.codec, s.versioner, data, out)
}

// func (s *store) extracKeys(attrs ...map[string]*awsdb.AttributeValue) (patitionKey *awsdb.AttributeValue, sortKey *awsdb.AttributeValue) {
// 	// for var := range attrs {
// 	//
// 	// }
// 	// partitionKeyitem = scanOut.Items[primaryKey]
// 	// sortkeyitem = scanOut.Items[sortKey]
//
// 	return nil, nil
// }
