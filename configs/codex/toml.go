package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

type values map[string]any

func pathID(path []string) string  { b, _ := json.Marshal(path); return string(b) }
func pathParts(id string) []string { var p []string; _ = json.Unmarshal([]byte(id), &p); return p }
func flatten(m map[string]any, prefix []string) values {
	out := values{}
	for k, v := range m {
		path := append(append([]string{}, prefix...), k)
		if child, ok := v.(map[string]any); ok {
			for id, leaf := range flatten(child, path) {
				out[id] = leaf
			}
		} else {
			out[pathID(path)] = v
		}
	}
	return out
}

func decode(data []byte) (map[string]any, error) {
	m := map[string]any{}
	err := toml.Unmarshal(data, &m)
	return m, err
}

// JSON normalization makes legacy Python snapshots and TOML integer values comparable.
func equal(a, b any) bool {
	aa, ea := json.Marshal(a)
	bb, eb := json.Marshal(b)
	return ea == nil && eb == nil && bytes.Equal(aa, bb)
}

func same(a values, b values, key string) bool {
	av, ao := a[key]
	bv, bo := b[key]
	return ao == bo && (!ao || equal(av, bv))
}

func merge(data []byte, source, previous map[string]any, filename string) ([]byte, []string, error) {
	local, err := decode(data)
	if err != nil {
		return nil, nil, err
	}
	current, desired, old := flatten(local, nil), flatten(source, nil), flatten(previous, nil)
	keys := map[string]bool{}
	for k := range desired {
		keys[k] = true
	}
	for k := range old {
		keys[k] = true
	}
	ordered := []string{}
	for k := range keys {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)
	result := bytes.Clone(data)
	conflicts := []string{}
	for _, k := range ordered {
		_, exists := current[k]
		if filename == "config.toml" && k == pathID([]string{"tui", "theme"}) && exists {
			continue
		}
		apply := false
		if previous == nil {
			apply = !exists
		} else if same(desired, old, k) || same(current, desired, k) {
			continue
		} else if same(current, old, k) {
			apply = true
		} else {
			conflicts = append(conflicts, filename+": "+strings.Join(pathParts(k), "."))
			continue
		}
		if apply {
			v, present := desired[k]
			result, err = editTOML(result, pathParts(k), v, present)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return result, conflicts, nil
}

type assignment struct {
	path              []string
	start, value, end int
}
type section struct {
	path      []string
	insertion int
}

func nodeKey(n *unstable.Node) []string {
	parts := []string{}
	keys := n.Key()
	for keys.Next() {
		parts = append(parts, string(keys.Node().Data))
	}
	return parts
}

// Only the requested assignment/value is edited; all unrelated bytes remain intact.
// The pinned parser owns strings, multiline values, arrays and quoted/dotted keys.
func indexTOML(data []byte) ([]assignment, []section, error) {
	var parser unstable.Parser
	parser.Reset(data)
	assignments := []assignment{}
	sections := []section{{insertion: 0}}
	var table []string
	arrayTable := false
	for parser.NextExpression() {
		n := parser.Expression()
		switch n.Kind {
		case unstable.Table, unstable.ArrayTable:
			table = nodeKey(n)
			arrayTable = n.Kind == unstable.ArrayTable
			if !arrayTable {
				keys := n.Key()
				keys.Next()
				start := int(keys.Node().Raw.Offset)
				end := bytes.IndexByte(data[start:], '\n')
				if end < 0 {
					end = len(data) - start
				} else {
					end++
				}
				sections = append(sections, section{append([]string{}, table...), start + end})
			}
		case unstable.KeyValue:
			if arrayTable {
				continue
			}
			keys := n.Key()
			var last *unstable.Node
			for keys.Next() {
				last = keys.Node()
			}
			start := int(n.Raw.Offset)
			end := start + int(n.Raw.Length)
			value := int(last.Raw.Offset + last.Raw.Length)
			for value < end && data[value] != '=' {
				value++
			}
			value++
			for value < end && (data[value] == ' ' || data[value] == '\t') {
				value++
			}
			path := append(append([]string{}, table...), nodeKey(n)...)
			assignments = append(assignments, assignment{path, start, value, end})
		}
	}
	return assignments, sections, parser.Error()
}

func prefix(a, b []string) bool { return len(a) <= len(b) && reflect.DeepEqual(a, b[:len(a)]) }
func quotedPath(path []string) string {
	parts := make([]string, len(path))
	for i, p := range path {
		b, _ := json.Marshal(p)
		parts[i] = string(b)
	}
	return strings.Join(parts, ".")
}

func encodedValue(v any) ([]byte, error) {
	var b bytes.Buffer
	err := toml.NewEncoder(&b).SetTablesInline(true).Encode(map[string]any{"value": v})
	if err != nil {
		return nil, err
	}
	_, value, ok := bytes.Cut(b.Bytes(), []byte("="))
	if !ok {
		return nil, fmt.Errorf("cannot encode value")
	}
	return bytes.TrimSpace(value), nil
}

func editTOML(data []byte, path []string, value any, present bool) ([]byte, error) {
	assignments, sections, err := indexTOML(data)
	if err != nil {
		return nil, err
	}
	var replacement []byte
	begin, end := -1, -1
	for _, a := range assignments {
		if !prefix(a.path, path) {
			continue
		}
		begin, end = a.value, a.end
		if len(a.path) < len(path) {
			wrapper, e := decode(append([]byte("value = "), data[a.value:a.end]...))
			if e != nil {
				return nil, e
			}
			node, ok := wrapper["value"].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("table conflicts with scalar: %s", strings.Join(path, "."))
			}
			rest := path[len(a.path):]
			current := node
			for _, k := range rest[:len(rest)-1] {
				if current[k] == nil {
					current[k] = map[string]any{}
				}
				child, ok := current[k].(map[string]any)
				if !ok {
					return nil, fmt.Errorf("table conflicts with scalar")
				}
				current = child
			}
			if present {
				current[rest[len(rest)-1]] = value
			} else {
				delete(current, rest[len(rest)-1])
			}
			replacement, err = encodedValue(node)
		} else if present {
			replacement, err = encodedValue(value)
		} else {
			begin = a.start
			// Remove the assignment, retaining an inline comment as a standalone comment.
			for begin > 0 && (data[begin-1] == ' ' || data[begin-1] == '\t') {
				begin--
			}
			// A removed assignment that occupied its whole line takes the line break too.
			if begin == 0 || data[begin-1] == '\n' {
				if bytes.HasPrefix(data[end:], []byte("\r\n")) {
					end += 2
				} else if bytes.HasPrefix(data[end:], []byte("\n")) {
					end++
				}
			}
		}
		break
	}
	if err != nil {
		return nil, err
	}
	if begin < 0 {
		if !present {
			return data, nil
		}
		best := section{insertion: 0}
		for _, s := range sections {
			if prefix(s.path, path[:len(path)-1]) && len(s.path) > len(best.path) {
				best = s
			}
		}
		v, e := encodedValue(value)
		if e != nil {
			return nil, e
		}
		newline := "\n"
		if bytes.Contains(data, []byte("\r\n")) {
			newline = "\r\n"
		}
		replacement = []byte(quotedPath(path[len(best.path):]) + " = " + string(v) + newline)
		begin, end = best.insertion, best.insertion
		if begin > 0 && data[begin-1] != '\n' {
			replacement = append([]byte(newline), replacement...)
		}
	}
	out := make([]byte, 0, len(data)+len(replacement))
	out = append(out, data[:begin]...)
	out = append(out, replacement...)
	out = append(out, data[end:]...)
	if _, err = decode(out); err != nil {
		return nil, fmt.Errorf("edited TOML is invalid: %w", err)
	}
	return out, nil
}
