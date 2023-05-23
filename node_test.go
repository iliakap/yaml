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
	"bytes"
	"fmt"
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"
)

var nodeTests = []struct {
	yaml string
	node yaml.Node
}{
	{
		// multiline literal starting with multiple newlines.
		"a: |2-\n\n\n  str\n",
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "a",
					Line:   1,
					Column: 1,
				}, {
					Kind:        yaml.ScalarNode,
					Style:       yaml.LiteralStyle,
					Tag:         "!!str",
					Value:       "\n\nstr",
					LineComment: "",
					Line:        1,
					Column:      4,
				}},
			}},
		},
	},
}

func (s *S) TestNodeRoundtrip(c *C) {
	defer os.Setenv("TZ", os.Getenv("TZ"))
	os.Setenv("TZ", "UTC")
	for i, item := range nodeTests {
		c.Logf("test %d: %q", i, item.yaml)

		if strings.Contains(item.yaml, "#") {
			var buf bytes.Buffer
			fprintComments(&buf, &item.node, "    ")
			c.Logf("  expected comments:\n%s", buf.Bytes())
		}

		decode := true
		encode := true

		testYaml := item.yaml
		if s := strings.TrimPrefix(testYaml, "[decode]"); s != testYaml {
			encode = false
			testYaml = s
		}
		if s := strings.TrimPrefix(testYaml, "[encode]"); s != testYaml {
			decode = false
			testYaml = s
		}

		if decode {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(testYaml), &node)
			c.Assert(err, IsNil)
			if strings.Contains(item.yaml, "#") {
				var buf bytes.Buffer
				fprintComments(&buf, &node, "    ")
				c.Logf("  obtained comments:\n%s", buf.Bytes())
			}
			c.Assert(&node, DeepEquals, &item.node)
		}
		if encode {
			node := deepCopyNode(&item.node, nil)
			buf := bytes.Buffer{}
			enc := yaml.NewEncoder(&buf)
			enc.SetIndent(2)
			err := enc.Encode(node)
			c.Assert(err, IsNil)
			err = enc.Close()
			c.Assert(err, IsNil)
			c.Assert(buf.String(), Equals, testYaml)

			// Ensure there were no mutations to the tree.
			c.Assert(node, DeepEquals, &item.node)
		}
	}
}

func deepCopyNode(node *yaml.Node, cache map[*yaml.Node]*yaml.Node) *yaml.Node {
	if n, ok := cache[node]; ok {
		return n
	}
	if cache == nil {
		cache = make(map[*yaml.Node]*yaml.Node)
	}
	copy := *node
	cache[node] = &copy
	copy.Content = nil
	for _, elem := range node.Content {
		copy.Content = append(copy.Content, deepCopyNode(elem, cache))
	}
	if node.Alias != nil {
		copy.Alias = deepCopyNode(node.Alias, cache)
	}
	return &copy
}

var savedNodes = make(map[string]*yaml.Node)

func saveNode(name string, node *yaml.Node) *yaml.Node {
	savedNodes[name] = node
	return node
}

func peekNode(name string) *yaml.Node {
	return savedNodes[name]
}

func dropNode(name string) *yaml.Node {
	node := savedNodes[name]
	delete(savedNodes, name)
	return node
}

var setStringTests = []struct {
	str  string
	yaml string
	node yaml.Node
}{
	{
		"something simple",
		"something simple\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "something simple",
			Tag:   "!!str",
		},
	}, {
		`"quoted value"`,
		"'\"quoted value\"'\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: `"quoted value"`,
			Tag:   "!!str",
		},
	}, {
		"multi\nline",
		"|-\n  multi\n  line\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "multi\nline",
			Tag:   "!!str",
			Style: yaml.LiteralStyle,
		},
	}, {
		"123",
		"\"123\"\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "123",
			Tag:   "!!str",
		},
	}, {
		"multi\nline\n",
		"|\n  multi\n  line\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "multi\nline\n",
			Tag:   "!!str",
			Style: yaml.LiteralStyle,
		},
	}, {
		"\x80\x81\x82",
		"!!binary gIGC\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "gIGC",
			Tag:   "!!binary",
		},
	},
}

func (s *S) TestSetString(c *C) {
	defer os.Setenv("TZ", os.Getenv("TZ"))
	os.Setenv("TZ", "UTC")
	for i, item := range setStringTests {
		c.Logf("test %d: %q", i, item.str)

		var node yaml.Node

		node.SetString(item.str)

		c.Assert(node, DeepEquals, item.node)

		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		err := enc.Encode(&item.node)
		c.Assert(err, IsNil)
		err = enc.Close()
		c.Assert(err, IsNil)
		c.Assert(buf.String(), Equals, item.yaml)

		var doc yaml.Node
		err = yaml.Unmarshal([]byte(item.yaml), &doc)
		c.Assert(err, IsNil)

		var str string
		err = node.Decode(&str)
		c.Assert(err, IsNil)
		c.Assert(str, Equals, item.str)
	}
}

var nodeEncodeDecodeTests = []struct {
	value interface{}
	yaml  string
	node  yaml.Node
}{{
	"something simple",
	"something simple\n",
	yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "something simple",
		Tag:   "!!str",
	},
}, {
	`"quoted value"`,
	"'\"quoted value\"'\n",
	yaml.Node{
		Kind:  yaml.ScalarNode,
		Style: yaml.SingleQuotedStyle,
		Value: `"quoted value"`,
		Tag:   "!!str",
	},
}, {
	123,
	"123",
	yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: `123`,
		Tag:   "!!int",
	},
}, {
	[]interface{}{1, 2},
	"[1, 2]",
	yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
		Content: []*yaml.Node{{
			Kind:  yaml.ScalarNode,
			Value: "1",
			Tag:   "!!int",
		}, {
			Kind:  yaml.ScalarNode,
			Value: "2",
			Tag:   "!!int",
		}},
	},
}, {
	map[string]interface{}{"a": "b"},
	"a: b",
	yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{{
			Kind:  yaml.ScalarNode,
			Value: "a",
			Tag:   "!!str",
		}, {
			Kind:  yaml.ScalarNode,
			Value: "b",
			Tag:   "!!str",
		}},
	},
}}

func (s *S) TestNodeEncodeDecode(c *C) {
	for i, item := range nodeEncodeDecodeTests {
		c.Logf("Encode/Decode test value #%d: %#v", i, item.value)

		var v interface{}
		err := item.node.Decode(&v)
		c.Assert(err, IsNil)
		c.Assert(v, DeepEquals, item.value)

		var n yaml.Node
		err = n.Encode(item.value)
		c.Assert(err, IsNil)
		c.Assert(n, DeepEquals, item.node)
	}
}

func (s *S) TestNodeZeroEncodeDecode(c *C) {
	// Zero node value behaves as nil when encoding...
	var n yaml.Node
	data, err := yaml.Marshal(&n)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "null\n")

	// ... and decoding.
	var v *struct{} = &struct{}{}
	c.Assert(n.Decode(&v), IsNil)
	c.Assert(v, IsNil)

	// ... and even when looking for its tag.
	c.Assert(n.ShortTag(), Equals, "!!null")

	// Kind zero is still unknown, though.
	n.Line = 1
	_, err = yaml.Marshal(&n)
	c.Assert(err, ErrorMatches, "yaml: cannot encode node with unknown kind 0")
	c.Assert(n.Decode(&v), ErrorMatches, "yaml: cannot decode node with unknown kind 0")
}

func (s *S) TestNodeOmitEmpty(c *C) {
	var v struct {
		A int
		B yaml.Node ",omitempty"
	}
	v.A = 1
	data, err := yaml.Marshal(&v)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "a: 1\n")

	v.B.Line = 1
	_, err = yaml.Marshal(&v)
	c.Assert(err, ErrorMatches, "yaml: cannot encode node with unknown kind 0")
}

func fprintComments(out io.Writer, node *yaml.Node, indent string) {
	switch node.Kind {
	case yaml.ScalarNode:
		fmt.Fprintf(out, "%s<%s> ", indent, node.Value)
		fprintCommentSet(out, node)
		fmt.Fprintf(out, "\n")
	case yaml.DocumentNode:
		fmt.Fprintf(out, "%s<DOC> ", indent)
		fprintCommentSet(out, node)
		fmt.Fprintf(out, "\n")
		for i := 0; i < len(node.Content); i++ {
			fprintComments(out, node.Content[i], indent+"  ")
		}
	case yaml.MappingNode:
		fmt.Fprintf(out, "%s<MAP> ", indent)
		fprintCommentSet(out, node)
		fmt.Fprintf(out, "\n")
		for i := 0; i < len(node.Content); i += 2 {
			fprintComments(out, node.Content[i], indent+"  ")
			fprintComments(out, node.Content[i+1], indent+"  ")
		}
	case yaml.SequenceNode:
		fmt.Fprintf(out, "%s<SEQ> ", indent)
		fprintCommentSet(out, node)
		fmt.Fprintf(out, "\n")
		for i := 0; i < len(node.Content); i++ {
			fprintComments(out, node.Content[i], indent+"  ")
		}
	}
}

func fprintCommentSet(out io.Writer, node *yaml.Node) {
	if len(node.HeadComment)+len(node.LineComment)+len(node.FootComment) > 0 {
		fmt.Fprintf(out, "%q / %q / %q", node.HeadComment, node.LineComment, node.FootComment)
	}
}
