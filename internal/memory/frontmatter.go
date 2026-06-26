package memory

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter wraps an ordered YAML mapping node. Using yaml.Node (rather than a
// Go map) preserves key order and nested structure — e.g. the index's nested
// tracker block — so the vault stays legible in Obsidian.
type Frontmatter struct {
	Node *yaml.Node
}

// NewFrontmatter returns an empty mapping.
func NewFrontmatter() Frontmatter {
	return Frontmatter{Node: &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}}
}

func scalarNode(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
}

func (f *Frontmatter) ensure() {
	if f.Node == nil {
		f.Node = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	}
}

// IsEmpty reports whether the frontmatter has no keys.
func (f Frontmatter) IsEmpty() bool {
	return f.Node == nil || len(f.Node.Content) == 0
}

// Get returns the value of a top-level scalar key.
func (f Frontmatter) Get(key string) (string, bool) {
	if f.Node == nil {
		return "", false
	}
	c := f.Node.Content
	for i := 0; i+1 < len(c); i += 2 {
		if c[i].Value == key {
			return c[i+1].Value, true
		}
	}
	return "", false
}

// Set assigns a top-level scalar key, replacing it in place if present and
// appending it (preserving insertion order) otherwise.
func (f *Frontmatter) Set(key, val string) {
	f.ensure()
	c := f.Node.Content
	for i := 0; i+1 < len(c); i += 2 {
		if c[i].Value == key {
			c[i+1] = scalarNode(val)
			return
		}
	}
	f.Node.Content = append(f.Node.Content, scalarNode(key), scalarNode(val))
}

// SetNode assigns a top-level key to an arbitrary node (e.g. a nested mapping
// such as tracker), replacing in place if present.
func (f *Frontmatter) SetNode(key string, node *yaml.Node) {
	f.ensure()
	c := f.Node.Content
	for i := 0; i+1 < len(c); i += 2 {
		if c[i].Value == key {
			c[i+1] = node
			return
		}
	}
	f.Node.Content = append(f.Node.Content, scalarNode(key), node)
}

// ensureMap returns the nested mapping node at key, creating (or replacing a
// non-mapping value with) an empty mapping if needed.
func (f *Frontmatter) ensureMap(key string) *yaml.Node {
	f.ensure()
	c := f.Node.Content
	for i := 0; i+1 < len(c); i += 2 {
		if c[i].Value == key {
			if c[i+1].Kind != yaml.MappingNode {
				c[i+1] = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			}
			return c[i+1]
		}
	}
	m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	f.Node.Content = append(f.Node.Content, scalarNode(key), m)
	return m
}

// setMapScalar sets key=val within a mapping node, replacing in place if present.
func setMapScalar(m *yaml.Node, key, val string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = scalarNode(val)
			return
		}
	}
	m.Content = append(m.Content, scalarNode(key), scalarNode(val))
}

// ToMap decodes the frontmatter into a plain map for JSON output.
func (f Frontmatter) ToMap() map[string]any {
	if f.IsEmpty() {
		return map[string]any{}
	}
	var m map[string]any
	if err := f.Node.Decode(&m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

// ParseDoc splits a markdown document into its frontmatter and trimmed body.
// A document without a leading --- block yields empty frontmatter and the whole
// (trimmed) input as the body.
func ParseDoc(raw string) (Frontmatter, string) {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	if strings.HasPrefix(raw, "---\n") {
		rest := raw[len("---\n"):]
		if fmText, after, found := strings.Cut(rest, "\n---"); found {
			return parseFM(fmText), strings.TrimSpace(after)
		}
	}
	return NewFrontmatter(), strings.TrimSpace(raw)
}

func parseFM(text string) Frontmatter {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil || len(doc.Content) == 0 {
		return NewFrontmatter()
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return NewFrontmatter()
	}
	return Frontmatter{Node: root}
}

// SerializeDoc renders frontmatter + body back to a markdown document. With no
// frontmatter it emits just the trimmed body plus a trailing newline.
func SerializeDoc(f Frontmatter, body string) string {
	body = strings.TrimSpace(body)
	if f.IsEmpty() {
		return body + "\n"
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(f.Node); err != nil {
		// Should not happen for a well-formed mapping node; fall back to body.
		return body + "\n"
	}
	_ = enc.Close()
	return "---\n" + buf.String() + "---\n\n" + body + "\n"
}
