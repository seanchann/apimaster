/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package client

import (
	"fmt"
	"time"

	"k8s.io/apiserver/pkg/storage"

	"k8s.io/klog"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type RequestMeta struct {
	Sess       *mgo.Session
	DBName     string
	Collection string
}

type QueryMetaData struct {
	Condition bson.M
	Limit     int
	Skip      int
	Sort      []string
}

func mongoQueryCondition(query *QueryMetaData) bson.M {
	return query.Condition
}

func mongoGetCollection(meta *RequestMeta) (*mgo.Collection, error) {

	var err error
	c := meta.Sess.DB(meta.DBName).C(meta.Collection)
	if c == nil {
		err = storage.NewInternalErrorf("collection(%v) not found in db(%v)", meta.Collection, meta.DBName)
	}

	return c, err
}

func mongoGetDatabase(meta *RequestMeta) (*mgo.Database, error) {

	var err error
	db := meta.Sess.DB(meta.DBName)
	if db == nil {
		err = storage.NewInternalErrorf(meta.DBName, fmt.Sprintf("open error"))
	}

	return db, err
}

func cloneRequestMeta(meta *RequestMeta) *RequestMeta {
	return &RequestMeta{
		Sess:       meta.Sess.Copy(),
		DBName:     meta.DBName,
		Collection: meta.Collection,
	}
}

func isCollectionExist(collection string, db *mgo.Database) (bool, error) {
	var isexist bool = false

	names, err := db.CollectionNames()
	if err != nil {
		err = storage.NewInternalErrorf("search collection(%v): %v", collection, err.Error())
	} else {
		for _, item := range names {
			if item == collection {
				isexist = true
				break
			}
		}
	}

	return isexist, err
}

func MongoEnsureIndex(req *RequestMeta, indexes []string) error {

	found := false
	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()

	collectExist := false
	db, err := mongoGetDatabase(meta)
	if err != nil {
		klog.Errorf("Got database:%s err:%s", meta.DBName, err)
		return err
	}
	collectExist, err = isCollectionExist(meta.Collection, db)
	if err != nil {
		klog.Errorf("Check Collection:%s exist err:%s", meta.Collection, err)
		return err
	}

	c, err := mongoGetCollection(meta)
	if err != nil {
		klog.Errorf("Got Collection:%s err:%s", meta.Collection, err)
		return err
	}

	if !collectExist {
		klog.Infof("Collection:%s not exist", meta.Collection)
	} else {
		var indexAll []mgo.Index
		indexAll, err = c.Indexes()
		if err != nil {
			klog.Errorf("Get index with Collection:%s err:%s", meta.Collection, err)
			return err
		}

		for _, indexItem := range indexAll {
			if len(indexItem.Key) != len(indexes) {
				continue
			}

			for i, item := range indexItem.Key {
				klog.V(5).Infof("Traversal index:%v, custom Index:%v", item, indexes[i])
				if item != indexes[i] {
					found = false
					break
				} else {
					found = true
				}
			}

			if found {
				break
			}
		}
	}

	klog.V(5).Infof("Collection index found=%v", found)
	//index not exist,make a index
	if !found {
		newIndex := mgo.Index{
			Key:        indexes,
			Unique:     true,
			DropDups:   true,
			Background: true, // See notes.
			Sparse:     true,
		}
		err = c.EnsureIndex(newIndex)
	}

	return err
}

func MongoEnsureIndexWithExpire(req *RequestMeta, indexes []string, afterTime time.Duration) error {

	found := false
	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()

	collectExist := false
	db, err := mongoGetDatabase(meta)
	if err != nil {
		klog.Errorf("Got database:%s err:%s", meta.DBName, err)
		return err
	}
	collectExist, err = isCollectionExist(meta.Collection, db)
	if err != nil {
		klog.Errorf("Check Collection:%s exist err:%s", meta.Collection, err)
		return err
	}

	c, err := mongoGetCollection(meta)
	if err != nil {
		klog.Errorf("Got Collection:%s err:%s", meta.Collection, err)
		return err
	}

	if !collectExist {
		klog.Infof("Collection:%s not exist", meta.Collection)
	} else {
		var indexAll []mgo.Index
		indexAll, err = c.Indexes()
		if err != nil {
			klog.Errorf("Get index with Collection:%s err:%s", meta.Collection, err)
			return err
		}

		for _, indexItem := range indexAll {
			if len(indexItem.Key) != len(indexes) {
				continue
			}

			for i, item := range indexItem.Key {
				klog.V(5).Infof("Traversal index:%v, custom Index:%v", item, indexes[i])
				if item != indexes[i] {
					found = false
					break
				} else {
					found = true
				}
			}

			if found {
				break
			}
		}
	}

	klog.V(5).Infof("Collection index found=%v", found)
	//index not exist,make a index
	if !found {
		newIndex := mgo.Index{
			Key:         indexes,
			Unique:      false,
			DropDups:    false,
			Background:  true, // See notes.
			Sparse:      true,
			ExpireAfter: afterTime,
		}
		err = c.EnsureIndex(newIndex)
	}

	return err
}

func MongoQueryCount(req *RequestMeta, query *QueryMetaData) (int, error) {
	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()
	klog.V(5).Infof("Mongo query all with collection:%s", meta.Collection)

	c, err := mongoGetCollection(meta)
	if err != nil {
		return 0, err
	}

	condition := mongoQueryCondition(query)
	klog.V(5).Infof("Mongo query build condition:%+v", condition)

	return c.Find(condition).Count()
}

func MongoNormalQuery(req *RequestMeta, query *QueryMetaData, result interface{}) error {
	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()

	klog.V(5).Infof("Mongo query all with collection:%s", meta.Collection)

	c, err := mongoGetCollection(meta)
	if err != nil {
		return err
	}

	klog.V(5).Infoln("Mongo query build condition...")
	condition := mongoQueryCondition(query)

	resultSets := c.Find(condition)

	if len(query.Sort) > 0 {
		resultSets = resultSets.Sort(query.Sort...)
	}

	if query.Limit > 0 {
		resultSets = resultSets.Limit(query.Limit)
	}

	if query.Skip > 0 {
		resultSets = resultSets.Skip(query.Skip)
	}

	return resultSets.All(result)

}

func MongoQueryOne(req *RequestMeta, query *QueryMetaData, result interface{}) error {
	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()

	klog.V(5).Infof("Mongo query with collection:%s", meta.Collection)

	c, err := mongoGetCollection(meta)
	if err != nil {
		return err
	}

	condition := mongoQueryCondition(query)
	klog.V(5).Infof("Mongo query build condition:%+v", condition)

	return c.Find(condition).One(result)

}

func MongoQueryOneNoCondition(req *RequestMeta, result interface{}) error {

	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()
	klog.V(5).Infof("Mongo query no condition with collection:%s", meta.Collection)

	c, err := mongoGetCollection(meta)
	if err != nil {
		klog.Errorf("MongoQueryOneNoCondition err %v", err)
		return err
	}

	return c.Find(nil).One(result)

}

func MongInsertOne(req *RequestMeta, docs interface{}) error {
	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()

	c, err := mongoGetCollection(meta)
	if err != nil {
		klog.V(5).Infoln("MongInsertOne err ")
		return err
	}

	return c.Insert(docs)
}

func MongDeleteOne(req *RequestMeta, selector interface{}) error {

	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()

	c, err := mongoGetCollection(meta)
	if err != nil {
		return err
	}

	return c.Remove(selector)
}

func MongUpsertOne(req *RequestMeta, selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {

	meta := cloneRequestMeta(req)
	defer meta.Sess.Close()

	c, err := mongoGetCollection(meta)
	if err != nil {
		return nil, err
	}

	return c.Upsert(selector, update)
}
