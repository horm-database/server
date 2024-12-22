// Copyright (c) 2024 The horm-database Authors (such as CaoHao <18500482693@163.com>). All rights reserved.
//
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
package main

import "fmt"

const (
	MajorVersion  = 0     //MajorVersion when you make incompatible changes .
	MinorVersion  = 0     //MinorVersion when you add functionality in a backwards-compatible manner .
	PatchVersion  = 1     //PatchVersion  when you fix bugs .
	VersionSuffix = "dev" // -alpha -alpha.1 -beta -rc -rc.1
)

// Version returns the version of server.
func Version() string {
	return fmt.Sprintf("v%d.%d.%d-%s", MajorVersion, MinorVersion, PatchVersion, VersionSuffix)
}
