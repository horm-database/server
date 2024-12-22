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
	"fmt"

	"github.com/horm-database/common/types"
)

func init() {
	RegisterSerializer(SerializationTypeXML, &XMLSerialization{})
}

// XMLSerialization export xml body
type XMLSerialization struct{}

// Deserialize the input xml into body
func (j *XMLSerialization) Deserialize(in []byte, body interface{}) error {
	body = in
	return nil
}

// Serialize the body to output xml.
func (j *XMLSerialization) Serialize(body interface{}) ([]byte, error) {
	return types.StringToBytes(fmt.Sprintf("%v", body)), nil
}
