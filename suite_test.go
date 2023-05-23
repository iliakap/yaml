//
// Copyright (c) 2011-2019 Canonical Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yaml_test

import (
	_ "embed"
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

//go:embed a.yaml
var yml string

type Yml struct {
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
	Other  string `yaml:"other"`
}

//func (s S) TestFoo(c *C) {
//	//var m1 = make(map[string]string)
//	//var m2 = make(map[string]string)
//	var m1 Yml
//	var m2 Yml
//	var m3 Yml
//	err := yaml.Unmarshal(([]byte)(yml), &m1)
//	out1, err := yaml.Marshal(m1)
//	err = yaml.Unmarshal(out1, &m2)
//	out2, err := yaml.Marshal(m2)
//	err = yaml.Unmarshal(out2, &m3)
//	out3, err := yaml.Marshal(m3)
//	os.WriteFile("a1.yaml", out1, 0644)
//	os.WriteFile("a2.yaml", out2, 0644)
//	os.WriteFile("a3.yaml", out3, 0644)
//	fmt.Println(err)
//	fmt.Println(string(out1))
//	fmt.Println(string(out2))
//	fmt.Println(string(out3))
//}

type S struct{}

var _ = Suite(&S{})
