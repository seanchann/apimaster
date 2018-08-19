package dynamodb

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage"

	awsdb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/golang/glog"
)

func BuildScanFilterAttr(objPtr runtime.Object, p storage.SelectionPredicate) (filterExpression string, expressionAttributeNames map[string]*string, expressionAttributeValues map[string]*awsdb.AttributeValue) {

	var filterExpressionList []string
	expressionAttributeNames = make(map[string]*string)
	expressionAttributeValues = make(map[string]*awsdb.AttributeValue)

	var fieldsFilterExpression []string
	if p.Field != nil && !p.Field.Empty() {
		fieldsFilterExpression = ScanFilterWithFileds(p, expressionAttributeNames, expressionAttributeValues)
	}

	var lablesFilterExpression []string
	if p.Label != nil && !p.Label.Empty() {
		lablesFilterExpression = ScanFilterWithLables(p, expressionAttributeNames, expressionAttributeValues)
	}

	for _, v := range fieldsFilterExpression {
		filterExpressionList = append(filterExpressionList, v)
		filterExpressionList = append(filterExpressionList, " AND ")
	}

	for _, v := range lablesFilterExpression {
		filterExpressionList = append(filterExpressionList, v)
		filterExpressionList = append(filterExpressionList, " AND ")
	}

	if len(filterExpressionList) > 0 {
		//ignore last AND  in slice, not anymore condition to append
		filterExpressionList = filterExpressionList[0 : len(filterExpressionList)-1]

		filterExpression = strings.Join(filterExpressionList, "")

		glog.V(9).Infof("build scan filter by selectionPredicate %+v \r\n"+
			"result: filterExpression(%+v)  expressionAttributeNames(%+v) expressionAttributeValues(%+v)\r\n", p,
			filterExpression, expressionAttributeNames, expressionAttributeValues)

		for k, v := range expressionAttributeNames {
			glog.V(9).Infof("got attr name k:%v,v:%v", k, *v)
		}
	}

	return filterExpression, expressionAttributeNames, expressionAttributeValues
}

//BuildUpdateAttr input attr like as:
/*
	type nested Struct{
		key1 string
		key2 string
	}

	type Test Struct{
		key1 string `json:"key1,omitempty"`
		key2 string
		nested nested `json:"nested,omitempty"`
	}

if you want update Test.key1 and nested.key1 will be have map:
	attr:=map[string]interface{}{
		"key1":value1,
		"nested.#key1":value2,
	}
*/
func BuildUpdateAttr(newObj runtime.Object, oldObj runtime.Object, attr map[string]interface{}) (updateExpression string, expressionAttributeNames map[string]*string, expressionAttributeValues map[string]*awsdb.AttributeValue, err error) {

	updateExpression = string("SET ")
	expressionAttributeNames = make(map[string]*string)
	expressionAttributeValues = make(map[string]*awsdb.AttributeValue)
	for k, v := range attr {
		fieldName := k
		var tagName string
		if i := strings.LastIndexAny(fieldName, "."); i >= 0 {
			fieldName = fieldName[i+1:]
			if i := strings.IndexAny(fieldName, "#"); i >= 0 {
				tagName = fieldName[i+1:]
			} else {
				err = fmt.Errorf("not found '#' with %v", k)
				return
			}
		} else {
			tagName = fieldName
		}

		if updateExpression != string("SET ") {
			updateExpression += string(",")
		}
		updateExpression += fmt.Sprintf("%s= :%s ", k, tagName)

		items := strings.Split(k, ".")

		for _, query := range items[1:] {
			if i := strings.IndexAny(query, "#"); i >= 0 {
				name := fmt.Sprintf("%s", query[i+1:])
				expressionAttributeNames[query] = &name
			} else {
				err = fmt.Errorf("not found '#' with %v", query)
				return
			}
		}

		dynaAttr, dynaErr := dynamodbattribute.ConvertTo(v)
		if dynaErr != nil {
			err = dynaErr
			return
		}
		expressionAttributeValues[fmt.Sprintf(":%s", tagName)] = dynaAttr
	}

	return
}
