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
	"github.com/horm-database/common/json"
)

func init() {
	RegisterSerializer(SerializationTypeJSON, &JSONSerialization{})
}

// JSONSerialization provides json serialization mode.
type JSONSerialization struct{}

// Deserialize json unmarshal the input bytes into body.
func (s *JSONSerialization) Deserialize(in []byte, body interface{}) error {
	return json.Api.Unmarshal(in, body)
}

// Serialize json marshal the body into output bytes
func (s *JSONSerialization) Serialize(body interface{}) ([]byte, error) {
	return json.MarshalBase(body)
}
