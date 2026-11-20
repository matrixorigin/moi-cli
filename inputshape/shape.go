// Package inputshape describes Go JSON decode structures without business validation.
package inputshape

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Describe derives a structural description from the actual encoding/json type.
func Describe(typ reflect.Type) map[string]any {
	return inputSchema(typ, map[reflect.Type]bool{})
}

func inputSchema(typ reflect.Type, seen map[reflect.Type]bool) map[string]any {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	// Types with custom JSON encoding (e.g. protobuf Value) cannot be inferred
	// from their internal Go fields. Do not claim that those fields are the wire.
	if reflect.PointerTo(typ).Implements(reflect.TypeFor[json.Unmarshaler]()) || typ.Implements(reflect.TypeFor[json.Unmarshaler]()) {
		return map[string]any{"description": "Custom JSON value; consult command/domain documentation"}
	}
	if seen[typ] {
		return map[string]any{"description": "Recursive value: " + typ.String()}
	}
	seen[typ] = true
	defer delete(seen, typ)
	switch typ.Kind() {
	case reflect.Struct:
		properties := map[string]any{}
		for name, fieldType := range inputJSONFields(typ) {
			properties[name] = inputSchema(fieldType, seen)
		}
		return map[string]any{"type": "object", "properties": properties, "additionalProperties": true}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": inputSchema(typ.Elem(), seen)}
	case reflect.Slice, reflect.Array:
		if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Uint8 {
			return map[string]any{"type": "string", "contentEncoding": "base64"}
		}
		return map[string]any{"type": "array", "items": inputSchema(typ.Elem(), seen)}
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	default:
		return map[string]any{}
	}
}

// inputJSONFields follows encoding/json's embedded-field selection: untagged
// anonymous structs are promoted, shallow fields win, and a tagged field wins
// over untagged fields at the same depth. Equal-ranked conflicts are omitted.
func inputJSONFields(typ reflect.Type) map[string]reflect.Type {
	type candidate struct {
		typ       reflect.Type
		depth     int
		tagged    bool
		ambiguous bool
	}
	fields := map[string]candidate{}
	ancestors := map[reflect.Type]bool{}
	var visit func(reflect.Type, int)
	visit = func(typ reflect.Type, depth int) {
		if ancestors[typ] {
			return
		}
		ancestors[typ] = true
		defer delete(ancestors, typ)
		for i := range typ.NumField() {
			field := typ.Field(i)
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Pointer {
				embeddedType = embeddedType.Elem()
			}
			if !field.IsExported() && (!field.Anonymous || embeddedType.Kind() != reflect.Struct) {
				continue
			}
			if field.Tag.Get("json") == "-" {
				continue
			}
			name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "" && field.Anonymous && embeddedType.Kind() == reflect.Struct {
				visit(embeddedType, depth+1)
				continue
			}
			tagged := name != ""
			if name == "" {
				name = field.Name
			}
			previous, exists := fields[name]
			switch {
			case !exists || depth < previous.depth || (depth == previous.depth && tagged && !previous.tagged):
				fields[name] = candidate{typ: field.Type, depth: depth, tagged: tagged}
			case depth == previous.depth && tagged == previous.tagged:
				previous.ambiguous = true
				fields[name] = previous
			}
		}
	}
	visit(typ, 0)
	result := map[string]reflect.Type{}
	for name, field := range fields {
		if !field.ambiguous {
			result[name] = field.typ
		}
	}
	return result
}
