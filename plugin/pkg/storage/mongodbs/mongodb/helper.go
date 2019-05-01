/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package mongodb

import (
	"fmt"
	"sort"
	"strings"

	"github.com/seanchann/apimaster/plugin/storage/mongodbs/client"
	"k8s.io/apiserver/pkg/storage"

	"k8s.io/klog"
	"gopkg.in/mgo.v2/bson"
)

type operator string

const (
	equalsOperator       operator = "="
	doubleEqualsOperator operator = "=="
	inOperator           operator = "in"
	notEqualsOperator    operator = "!="
	notInOperator        operator = "notin"
	existsOperator       operator = "exists"
)

type selectorItem struct {
	key, value string
	opCode     operator
}

func labelsSelectorToCondition(labels string, condition *client.QueryMetaData) {
	klog.V(5).Infof("mongo driver list with filter(%s):labelsSelectorToCondition", labels)
}

func try(selectorPiece, op string) (lhs, rhs string, ok bool) {
	pieces := strings.Split(selectorPiece, op)
	if len(pieces) == 2 {
		keyslice := strings.Split(pieces[0], ".")
		key := keyslice[len(keyslice)-1]
		return key, pieces[1], true
	}
	return "", "", false
}

func parseFields(selector string, selectorItems *[]selectorItem) error {
	parts := strings.Split(selector, ",")
	sort.StringSlice(parts).Sort()
	for _, part := range parts {
		if part == "" {
			continue
		}
		klog.V(5).Infof("Parse filed:%v", part)
		if lhs, rhs, ok := try(part, string(notEqualsOperator)); ok {
			*selectorItems = append(*selectorItems, selectorItem{key: lhs, value: rhs, opCode: notEqualsOperator})
		} else if lhs, rhs, ok := try(part, string(doubleEqualsOperator)); ok {
			*selectorItems = append(*selectorItems, selectorItem{key: lhs, value: rhs, opCode: doubleEqualsOperator})
		} else if lhs, rhs, ok := try(part, string(equalsOperator)); ok {
			*selectorItems = append(*selectorItems, selectorItem{key: lhs, value: rhs, opCode: equalsOperator})
		} else {
			return fmt.Errorf("invalid selector: '%s'; can't understand '%s'", selector, part)
		}
	}
	return nil
}

func fieldsSelectorToCondition(fields string, selector []selectorItem, condition *client.QueryMetaData) {
	klog.V(5).Infof("mongo driver list with filter(%s):fieldsSelectorToCondition", fields)

	err := parseFields(fields, &selector)
	if err != nil {
		klog.Errorf("parse fields err:%v", err)
		return
	}

	items := selector
	if len(items) > 0 {
		condition.Condition["$and"] = []bson.M{}
		klog.V(5).Infof("Convert selector items(%+v) to condition", items)
		for _, item := range items {
			equalRegex := fmt.Sprintf("\"%v\":\"%v\"", item.key, item.value)
			equalRegexBson := bson.M{"value": bson.M{"$regex": bson.RegEx{equalRegex, ""}}}

			notEqualRegexBson := bson.M{"value": bson.M{"$not": bson.RegEx{equalRegex, ""}}}

			switch item.opCode {
			case equalsOperator:
				fallthrough
			case doubleEqualsOperator:
				condition.Condition["$and"] = append(condition.Condition["$and"].([]bson.M), equalRegexBson)
			case notEqualsOperator:
				klog.V(5).Infof("Convert selector item(%+v) to condition", item)
				condition.Condition["$and"] = append(condition.Condition["$and"].([]bson.M), notEqualRegexBson)
			default:
				klog.Warningln("invalid selector operator")
			}
		}
	}
}

func pagerToCondition(meta *client.RequestMeta, pager storage.SelectionPredicate, condition *client.QueryMetaData) {
	klog.V(5).Infof("mongo driver list with filter:pagerToCondition")

	itemSum, err := client.MongoQueryCount(meta, condition)
	if err != nil {
		klog.Errorf("Request Document Count err:%v", err)
		return
	}
	klog.V(5).Infof("Query Count is:%v", itemSum)
	//update current item sum
	pager.SetItemTotal(uint64(itemSum))

	//if there have not present page do nothing
	has, _, perPage := pager.PresentPage()
	if !has {
		return
	}

	var skip int
	hasPrev, prevPage, prevPerPage := pager.PreviousPage()
	if hasPrev {
		skip = int(prevPage * prevPerPage)
	} else {
		skip = 0
	}

	condition.Limit = int(perPage)
	condition.Skip = skip
	condition.Sort = append(condition.Sort, "lastmodifytime")
}

func Condition(meta *client.RequestMeta, condition *client.QueryMetaData, p storage.SelectionPredicate) error {
	var selector []selectorItem

	if p.Label != nil && !p.Label.Empty() {
		labelsSelectorToCondition(p.Label.String(), condition)
	}

	if p.Field != nil && !p.Field.Empty() {
		fieldsSelectorToCondition(p.Field.String(), selector, condition)
	}

	if p.Page != nil && !p.Page.Empty() {
		pagerToCondition(meta, p.Page, condition)
	}
	return nil
}
