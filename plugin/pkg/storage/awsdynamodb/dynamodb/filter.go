package dynamodb

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apiserver/pkg/storage"

	awsdb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"k8s.io/klog"
)

const (
	DynamoDBOpEqual        = "="
	DynamoDBOpNotEqual     = "!="
	DynamoDBOpAttrExist    = "attribute_exists"
	DynamoDBOpAttrNotExist = "attribute_not_exists"
	DynamoDBOpContains     = "contains"    //"CONTAINS"
	DynamoDBOpNotContains  = "not_contain" //"NOT_CONTAINS"
)

//TODO: instead of our case statement
type dynamodbOP struct {
	op        string
	condition func(expressionNameList []string, op string,
		value interface{},
		expressionAttributeNames map[string]*string, expressionAttributeValues map[string]*awsdb.AttributeValue) (condition string, err error)
}

var selectionToDynamo = map[string]string{
	string(selection.Equals):       DynamoDBOpEqual,
	string(selection.DoubleEquals): DynamoDBOpEqual,
	string(selection.DoesNotExist): DynamoDBOpAttrNotExist,
	string(selection.Exists):       DynamoDBOpAttrExist,
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//fieldToCondition convert filed into dynamodb conditions
/*
	input filed(metadata.name) and value in filed
	output:
	 	condition(metadata.#name = :name)
		expressionAttributeNames:
			{
				"#metadata.#name":"metadata.name"
			}
		expressionAttributeValues:
			{
				":name": value
			}
*/
func fieldToCondition(field string, operator string, value interface{},
	expressionAttributeNames map[string]*string,
	expressionAttributeValues map[string]*awsdb.AttributeValue) (condition string, err error) {

	fieldNames := strings.Split(field, ".")

	var expression []string

	var expressionNameKey []string
	for _, v := range fieldNames {
		actualname := fmt.Sprintf("%s", v)
		name := fmt.Sprintf("#%s", v)

		expressionNameKey = append(expressionNameKey, ".")
		expressionNameKey = append(expressionNameKey, name)
		expressionAttributeNames[name] = &actualname

		klog.V(9).Infof("append name:%v actualName:%v\r\n", expressionNameKey, actualname)
	}

	var expressionValName string
	switch operator {
	case DynamoDBOpEqual:
		fallthrough
	case DynamoDBOpNotEqual:
		expression = append(expression, expressionNameKey...)
		expression = append(expression, operator)
		//use last filed as a valname
		expressionValName = fmt.Sprintf(":%sval%s", fieldNames[len(fieldNames)-1], RandStringRunes(6))
		expression = append(expression, expressionValName)

		//need to ignore first '.' in slice
		condition = strings.Join(expression[1:], "")
	case DynamoDBOpContains:
		fallthrough
	case DynamoDBOpNotContains:
		//use last filed as a valname
		expressionValName = fmt.Sprintf(":%sval%s", fieldNames[len(fieldNames)-1], RandStringRunes(6))

		containsName := strings.Join(expressionNameKey[1:], "")
		containsExpression := fmt.Sprintf("%s(%s,%s)", operator, containsName, expressionValName)
		expression = append(expression, containsExpression)

		condition = strings.Join(expression, "")
	case DynamoDBOpAttrExist:
		fallthrough
	case DynamoDBOpAttrNotExist:
		attrName := strings.Join(expressionNameKey[1:], "")
		attrExpression := fmt.Sprintf("%s(%s)", operator, attrName)
		expression = append(expression, attrExpression)

		//need to ignore first '.' in slice
		condition = strings.Join(expression, "")

		//not has attr value, return
		return condition, nil
	}

	dynaAttr, err := dynamodbattribute.ConvertTo(value)
	if err != nil {
		return "", err
	}
	expressionAttributeValues[expressionValName] = dynaAttr

	return condition, nil
}

//ScanFilterWithFileds range FieldSelector convert into dynamodb scan
//return string is a slice constain of filter expression for dynamodb
func ScanFilterWithFileds(p storage.SelectionPredicate,
	expressionAttributeNames map[string]*string,
	expressionAttributeValues map[string]*awsdb.AttributeValue) (expression []string) {

	fieldsCondition := p.Field.Requirements()

	for _, v := range fieldsCondition {
		operator := ""
		var value interface{}
		switch v.Operator {
		case selection.Equals:
			fallthrough
		case selection.DoubleEquals:
			fallthrough
		case selection.NotEquals:
			operator = fmt.Sprintf("%s", selectionToDynamo[string(v.Operator)])
			value = v.Value
		case selection.Exists:
			fallthrough
		case selection.DoesNotExist:
			operator = fmt.Sprintf("%s", selectionToDynamo[string(v.Operator)])
			valList := strings.Split(v.Field, ".")
			if len(valList) > 0 {
				value = valList[len(valList)-1]
			} else {
				value = v.Field
			}
		default:
			klog.Errorf("not support %v", v.Operator)
			continue
		}

		field := fmt.Sprintf("%s", v.Field)

		condition, err := fieldToCondition(field, operator, value, expressionAttributeNames, expressionAttributeValues)
		if err != nil {
			klog.Errorf("convert selector(%+v) filed to condition error(%v)", p, err)
			continue
		}
		expression = append(expression, condition)
	}

	klog.V(9).Infof("field condition filterexpression:(%v) namemap:(%#v) valuemap:(%+v)", expression, expressionAttributeNames, expressionAttributeValues)

	return expression
}

//ScanFilterWithFileds range LablesSelector convert into dynamodb scan
//return string is a slice constain of filter expression for dynamodb
func ScanFilterWithLables(p storage.SelectionPredicate,
	expressionAttributeNames map[string]*string,
	expressionAttributeValues map[string]*awsdb.AttributeValue) (expression []string) {

	lables, selectable := p.Label.Requirements()
	if !selectable {
		return
	}

	for _, v := range lables {
		operator := ""
		var value interface{}
		switch v.Operator() {
		case selection.Equals:
			fallthrough
		case selection.DoubleEquals:
			fallthrough
		case selection.NotEquals:
			operator = fmt.Sprintf("%s", selectionToDynamo[string(v.Operator())])
			valList := v.Values().List()
			value = valList[0]
		case selection.Exists:
			fallthrough
		case selection.DoesNotExist:
			operator = fmt.Sprintf("%s", selectionToDynamo[string(v.Operator())])
			valList := strings.Split(v.Key(), ".")
			if len(valList) > 0 {
				value = valList[len(valList)-1]
			} else {
				value = v.Key()
			}
		default:
			klog.Errorf("not support %v", v.Operator())
			continue
		}

		field := fmt.Sprintf("metadata.lables.%s", v.Key())

		klog.V(9).Infof("convert filed(%v) operator(%v) value(%v) into filters", field, operator, value)
		condition, err := fieldToCondition(field, operator, value, expressionAttributeNames, expressionAttributeValues)
		if err != nil {
			klog.Errorf("convert selector(%+v) filed to condition error(%v)", p, err)
			continue
		}
		expression = append(expression, condition)
	}

	klog.V(9).Infof("lables condition filterexpression:(%v) namemap:(%#v) valuemap:(%+v)", expression, expressionAttributeNames, expressionAttributeValues)

	return
}
