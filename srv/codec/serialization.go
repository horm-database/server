// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codec

import (
	"context"
	"errors"

	"github.com/horm-database/common/codec"
)

// Serializer defines body serialization interface.
type Serializer interface {
	// Deserialize the input bytes into body
	Deserialize(in []byte, body interface{}) error

	// Serialize the body to output byte
	Serialize(body interface{}) (out []byte, err error)
}

// SerializationType defines the type of different serializers, such as json, xml
const (
	// SerializationTypeJSON is json serialization code.
	SerializationTypeJSON = 0

	// SerializationTypeXML is used to export xml body
	SerializationTypeXML = 1
)

var serializers = make(map[int]Serializer)

// RegisterSerializer registers serializer
func RegisterSerializer(serializationType int, s Serializer) {
	serializers[serializationType] = s
}

// GetSerializer returns the serializer by type.
func GetSerializer(serializationType int) Serializer {
	return serializers[serializationType]
}

// Deserialize the input bytes into body.
// The specific serialization mode is defined by type, json is default mode.
func Deserialize(ctx context.Context, in []byte, body interface{}) error {
	if body == nil || len(in) == 0 {
		return nil
	}

	msg := codec.Message(ctx)
	if msg == nil {
		return errors.New("not find serializationType")
	}

	s := GetSerializer(msg.SerializationType())
	if s == nil {
		return errors.New("serializer not registered")
	}

	return s.Deserialize(in, body)
}

// Serialize the body to output byte
// The specific serialization mode is defined by type, json is default mode.
func Serialize(ctx context.Context, body interface{}) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	msg := codec.Message(ctx)
	if msg == nil {
		return nil, errors.New("not find serializationType")
	}

	s := GetSerializer(msg.SerializationType())
	if s == nil {
		return nil, errors.New("serializer not registered")
	}
	return s.Serialize(body)
}
