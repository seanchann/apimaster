package mongodb

import (
	"context"
	"reflect"

	"gofreezer/pkg/storage/mongodbs"
	"gofreezer/pkg/storage/mongodbs/client"

	pluginstorage "github.com/seanchann/apimaster/plugin/storage"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage"

	"github.com/golang/glog"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type store struct {
	codec     runtime.Codec
	versioner APIObjectVersioner
	session   *mgo.Session
	dbname    string
}

//New create a mongo store
func New(sess *mgo.Session, dbName string, codec runtime.Codec) *store {
	versioner := APIObjectVersioner{}
	return &store{
		codec:     codec,
		versioner: versioner,
		dbname:    dbName,
		session:   sess,
	}
}

func (s *store) Type() string {
	return string("mongo")
}

// Versioner implements storage.Interface.Versioner.
func (s *store) Versioner() storage.Versioner {
	return s.versioner
}

func (s *store) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {

	c, err := GetCollection(s.dbname, s.session, obj)
	if err != nil {
		return err
	}

	meta := c.GetRequestMeta(s.session)
	if meta == nil {
		return storage.NewInternalErrorf("key %v, object can't convent into collection", key)
	}
	//first check resource exist
	err = s.getObject(meta, key, out, true)
	if err != nil {
		return err
	}

	doc := NewDocument(ttl, obj)
	data, err := doc.Encode(s.codec, key)
	if err != nil {
		return storage.NewInternalErrorf("key %v, object encode error %v", key, err.Error())
	}

	err = client.MongInsertOne(meta, data)
	if err != nil {
		return storage.NewInternalErrorf("create resource error %v", err.Error())
	}
	return nil
}

func (s *store) Delete(ctx context.Context, key string, out runtime.Object, preconditions *storage.Preconditions) error {
	c, err := GetCollection(s.dbname, s.session, out)
	if err != nil {
		return err
	}

	meta := c.GetRequestMeta(s.session)
	if meta == nil {
		return storage.NewInternalErrorf("key %v, object can't convent into collection", key)
	}

	//first check resource exist
	err = s.getObject(meta, key, out, false)
	if err != nil {
		return err
	}

	err = client.MongDeleteOne(meta, bson.M{"key": key})
	if err != nil {
		return storage.NewInternalErrorf("key %v delete error %v", key, err.Error())
	}

	return nil
}

func (s *store) Get(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool) error {

	c, err := GetCollection(s.dbname, s.session, out)
	if err != nil {
		return err
	}

	meta := c.GetRequestMeta(s.session)
	if meta == nil {
		return storage.NewInternalErrorf("key %v, object can't convent into collection", key)
	}

	return s.getObject(meta, key, out, ignoreNotFound)
}

func (s *store) GetToList(ctx context.Context, key string, p storage.SelectionPredicate, listObj runtime.Object) error {
	listPtr, itemPtrObj, err := pluginstorage.GetListItemObj(listObj)
	if err != nil {
		return storage.NewInvalidObjError(key, err.Error())
	}

	c, err := GetCollection(s.dbname, s.session, itemPtrObj)
	if err != nil {
		return err
	}

	meta := c.GetRequestMeta(s.session)
	if meta == nil {
		return storage.NewInternalErrorf("key %v, object can't convent into collection", key)
	}

	//query  document in collection
	query := &client.QueryMetaData{}
	query.Condition = bson.M{}
	Condition(meta, query, p)

	var result []*DocObject
	err = client.MongoNormalQuery(meta, query, &result)

	if len(result) == 0 {
		return nil
	}

	return decodeList(result, listPtr, s.codec, s.versioner)
}

func (s *store) GuaranteedUpdate(ctx context.Context, key string, out runtime.Object, ignoreNotFound bool, precondtions *storage.Preconditions, tryUpdate mongodbs.UpdateFunc) error {

	c, err := GetCollection(s.dbname, s.session, out)
	if err != nil {
		return err
	}

	meta := c.GetRequestMeta(s.session)
	if meta == nil {
		return storage.NewInternalErrorf("key %v, object can't convent into collection", key)
	}

	err = s.getObject(meta, key, out, false)
	if err != nil {
		glog.Infof("not found %v", err)
		return err
	}

	ret, ttl, err := userUpdate(out, tryUpdate)
	if err != nil {
		return storage.NewInternalErrorf("key %s error:%v", key, err.Error())
	}

	doc := NewDocument(*ttl, ret)
	data, err := doc.Encode(s.codec, key)
	if err != nil {
		return storage.NewInternalErrorf("key %v, object encode error %v", key, err.Error())
	}

	dataObj := data.(DocObject)

	selector := bson.M{"key": key}
	updateDoc := bson.M{"$set": bson.M{"obj": dataObj.Object}}
	changeInfo, err := client.MongUpsertOne(meta, selector, updateDoc)
	if err != nil {
		return storage.NewInternalErrorf("key %v, update(%v) error %v", key, changeInfo, err.Error())
	}

	return nil
}

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
	//versioner.UpdateObject(objPtr, uint64(rev))
	return nil
}

// decodeList decodes a list of values into a list of objects, with resource version set to corresponding rev.
// On success, ListPtr would be set to the list of objects.
func decodeList(elems []*DocObject, ListPtr interface{}, codec runtime.Codec, versioner storage.Versioner) error {
	v, err := conversion.EnforcePtr(ListPtr)
	if err != nil || v.Kind() != reflect.Slice {
		panic("need ptr to slice")
	}
	for _, elem := range elems {
		obj, _, err := codec.Decode(elem.Object, nil, reflect.New(v.Type().Elem()).Interface().(runtime.Object))
		if err != nil {
			return err
		}
		// being unable to set the version does not prevent the object from being extracted
		// versioner.UpdateObject(obj, elem.rev)
		// if filter(obj) {
		v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
		// }
	}
	return nil
}

func (s *store) getObject(meta *client.RequestMeta, key string, out runtime.Object, ignoreNotFound bool) error {
	query := &client.QueryMetaData{
		Condition: bson.M{"key": key},
	}

	var docObj DocObject
	err := client.MongoQueryOne(meta, query, &docObj)
	if err != nil && err.Error() != "not found" {
		return storage.NewInternalErrorf("key %v, %v", key, err)
	}

	if len(docObj.Key) == 0 {
		if ignoreNotFound {
			return runtime.SetZeroValue(out)
		}
		return storage.NewItemNotFoundError(key)
	}

	return decode(s.codec, s.versioner, docObj.Object, out)
}

func getDoc(meta *client.RequestMeta, key string) (*DocObject, error) {
	query := &client.QueryMetaData{
		Condition: bson.M{"key": key},
	}

	var docObj DocObject
	err := client.MongoQueryOne(meta, query, &docObj)
	if err != nil {
		return nil, storage.NewInternalErrorf("key %v, %v", key, err.Error())
	}

	return &docObj, nil
}

func userUpdate(input runtime.Object, userUpdate mongodbs.UpdateFunc) (output runtime.Object, ttl *uint64, err error) {
	ret, ttl, err := userUpdate(input)
	if err != nil {
		return nil, nil, err
	}
	return ret, ttl, nil
}
