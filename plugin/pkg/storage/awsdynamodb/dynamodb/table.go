package dynamodb

/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsdb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"k8s.io/klog"
)

var defaultTable = os.Args[0]

const (
	primaryKey = "key"
	sortKey    = "sortKey"
)

func CreateTable(dbHandler *awsdb.DynamoDB, table string) (string, error) {
	descParams := &awsdb.DescribeTableInput{
		TableName: aws.String(table), // Required
	}
	resp, err := dbHandler.DescribeTable(descParams)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			goto createTable
		} else {
			return string(""), nil
		}
	}

	return resp.String(), err

createTable:
	params := &awsdb.CreateTableInput{
		TableName: aws.String(table),
		AttributeDefinitions: []*awsdb.AttributeDefinition{
			{
				AttributeName: aws.String(primaryKey),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(sortKey),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*awsdb.KeySchemaElement{
			{
				AttributeName: aws.String(primaryKey),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(sortKey),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &awsdb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(4),
			WriteCapacityUnits: aws.Int64(1),
		},
	}

	output, err := dbHandler.CreateTable(params)

	return output.String(), err
}

func convertObject(attr map[string]*awsdb.AttributeValue) (map[string]interface{}, error) {
	obj := make(map[string]interface{})

	//not use this key for runtime
	delete(attr, primaryKey)
	delete(attr, sortKey)
	// for k, v := range attr {
	//
	// 	return nil
	// }
	err := dynamodbattribute.UnmarshalMap(attr, &obj)
	if err != nil {
		klog.V(5).Infof("unmarshalMap error %v\r\n", err)
		return nil, err
	}

	return obj, nil
}

func ConvertTOJson(items *[]map[string]*awsdb.AttributeValue) ([]map[string]interface{}, int64, error) {

	var listObj []map[string]interface{}
	var count int64
	for _, item := range *items {
		obj, err := convertObject(item)
		if err != nil {
			return nil, count, err
		} else if len(obj) != 0 {
			count++
			listObj = append(listObj, obj)
		}
	}

	return listObj, count, nil
}

func ConvertByteToMap(data []byte) (map[string]interface{}, error) {
	mapObj := make(map[string]interface{})
	err := json.Unmarshal(data, &mapObj)
	return mapObj, err
}
