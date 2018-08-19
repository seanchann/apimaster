/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package mongodb

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type Document struct {
	TTL    time.Time
	Object runtime.Object
}

type DocObject struct {
	TTL    time.Time `bson:"ttl"`
	UID    types.UID `bson:"uid"`
	Key    string    `bson:"key"`
	Object []byte    `bson:"obj"`
}

var (
	//2200.01.01.00 as a forever ttl
	TTLForever = time.Date(2200, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
)

func NewDocument(ttl uint64, obj runtime.Object) *Document {
	expire := time.Now().Add(time.Duration(ttl) * time.Second)
	if ttl == 0 {
		expire = TTLForever
	}

	return &Document{
		TTL:    expire,
		Object: obj,
	}
}

func (doc *Document) Encode(codec runtime.Codec, key string) (interface{}, error) {
	data, err := runtime.Encode(codec, doc.Object)
	if err != nil {
		return nil, err
	}

	return DocObject{
		TTL:    doc.TTL,
		UID:    uuid.NewUUID(),
		Object: data,
		Key:    key,
	}, nil
}

func (doc *Document) Decode(codec runtime.Codec) (interface{}, error) {
	data, err := runtime.Encode(codec, doc.Object)
	if err != nil {
		return nil, err
	}

	return DocObject{
		TTL:    doc.TTL,
		UID:    uuid.NewUUID(),
		Object: data,
	}, nil
}
