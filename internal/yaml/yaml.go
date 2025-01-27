package yaml

import (
	"bytes"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// UpdateYAML reads a YAML content, updates it with new data while preserving formatting,
// and returns the updated YAML content
func UpdateYAML(content []byte, newData interface{}) ([]byte, error) {
	indent := detectIndentation(string(content))

	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := updateYamlFromStruct(&root, newData); err != nil {
		return nil, fmt.Errorf("failed to update YAML: %w", err)
	}

	root.Column = 0
	if len(root.Content) > 0 {
		root.Content[0].Column = 0
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
	if err := enc.Encode(&root); err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}

	return buf.Bytes(), nil
}

func detectIndentation(content string) int {
	lines := bytes.Split([]byte(content), []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 || line[0] != ' ' {
			continue
		}

		spaces := 0
		for _, ch := range line {
			if ch == ' ' {
				spaces++
			} else {
				break
			}
		}

		if spaces > 0 {
			return spaces
		}
	}

	return 2
}

func updateYamlFromStruct(node *yaml.Node, data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	mappingNode := node
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) != 1 {
			return fmt.Errorf("invalid YAML structure: document node should have exactly one child")
		}
		mappingNode = node.Content[0]
		rootOffset := mappingNode.Column
		node.Column = 0
		mappingNode.Column = 0
		adjustNodeColumns(mappingNode, rootOffset)
	}

	if mappingNode.Kind != yaml.MappingNode {
		mappingNode.Kind = yaml.MappingNode
		mappingNode.Tag = "!!map"
	}
	if mappingNode.Content == nil {
		mappingNode.Content = []*yaml.Node{}
	}

	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			if err := updateField(mappingNode, typ.Field(i), val.Field(i)); err != nil {
				return fmt.Errorf("failed to update field %s: %w", typ.Field(i).Name, err)
			}
		}
	case reflect.Map:
		if val.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("map key must be string")
		}
		for _, key := range val.MapKeys() {
			keyStr := key.String()
			keyNode, valueNode, found := findNodes(mappingNode, keyStr)
			if !found {
				keyNode = &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!str",
					Value: keyStr,
				}
				valueNode = &yaml.Node{}
				if len(mappingNode.Content) > 0 {
					keyNode.Style = mappingNode.Content[0].Style
					keyNode.Column = mappingNode.Content[0].Column
					valueNode.Style = mappingNode.Content[1].Style
					valueNode.Column = mappingNode.Content[1].Column
				} else {
					keyNode.Column = mappingNode.Column + 2
					valueNode.Column = mappingNode.Column + 2
				}
				mappingNode.Content = append(mappingNode.Content, keyNode, valueNode)
			}
			if err := updateNode(valueNode, val.MapIndex(key)); err != nil {
				return fmt.Errorf("failed to update map value for key %s: %w", keyStr, err)
			}
		}
	default:
		return fmt.Errorf("data must be a struct, pointer to struct, or map[string]interface{}")
	}

	return nil
}

func adjustNodeColumns(node *yaml.Node, offset int) {
	if node.Column > offset {
		node.Column -= offset
	}
	for _, child := range node.Content {
		adjustNodeColumns(child, offset)
	}
}

func updateField(mappingNode *yaml.Node, fieldType reflect.StructField, fieldValue reflect.Value) error {
	yamlTag := fieldType.Tag.Get("yaml")
	if yamlTag == "" {
		yamlTag = fieldType.Name
	}

	keyNode, valueNode, found := findNodes(mappingNode, yamlTag)
	if !found {
		keyNode = &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: yamlTag,
		}
		valueNode = &yaml.Node{}
		if len(mappingNode.Content) > 0 {
			keyNode.Style = mappingNode.Content[0].Style
			keyNode.Column = mappingNode.Content[0].Column
			valueNode.Style = mappingNode.Content[1].Style
			valueNode.Column = mappingNode.Content[1].Column
		} else {
			keyNode.Column = mappingNode.Column + 2
			valueNode.Column = mappingNode.Column + 2
		}
		mappingNode.Content = append(mappingNode.Content, keyNode, valueNode)
	}

	return updateNode(valueNode, fieldValue)
}

func findNodes(mappingNode *yaml.Node, key string) (keyNode, valueNode *yaml.Node, found bool) {
	for i := 0; i < len(mappingNode.Content); i += 2 {
		if mappingNode.Content[i].Value == key {
			return mappingNode.Content[i], mappingNode.Content[i+1], true
		}
	}
	return nil, nil, false
}

func updateNode(node *yaml.Node, value reflect.Value) error {
	originalStyle := node.Style
	originalColumn := node.Column

	switch value.Kind() {
	case reflect.Interface:
		if !value.IsNil() {
			return updateNode(node, value.Elem())
		}
	case reflect.Struct:
		if err := updateYamlFromStruct(node, value.Interface()); err != nil {
			return err
		}
	case reflect.Slice, reflect.Array:
		if err := updateSequence(node, value); err != nil {
			return err
		}
	case reflect.Map:
		if err := updateMapping(node, value); err != nil {
			return err
		}
	default:
		node.Kind = yaml.ScalarNode
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			node.Tag = "!!int"
			node.Value = fmt.Sprintf("%d", value.Int())
		case reflect.Float32, reflect.Float64:
			node.Tag = "!!float"
			node.Value = fmt.Sprintf("%g", value.Float())
		case reflect.Bool:
			node.Tag = "!!bool"
			node.Value = fmt.Sprintf("%v", value.Bool())
		case reflect.String:
			node.Tag = "!!str"
			node.Value = value.String()
		default:
			node.Tag = "!!str"
			node.Value = fmt.Sprintf("%v", value.Interface())
		}
	}

	node.Style = originalStyle
	node.Column = originalColumn
	return nil
}

func updateSequence(node *yaml.Node, value reflect.Value) error {
	originalStyle := node.Style
	originalColumn := node.Column
	originalContent := node.Content

	node.Kind = yaml.SequenceNode
	node.Tag = "!!seq"
	if node.Content == nil {
		node.Content = []*yaml.Node{}
	}

	baseIndent := 2
	if len(originalContent) > 0 {
		baseIndent = originalContent[0].Column - node.Column
	}

	newContent := make([]*yaml.Node, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		elemNode := createOrReuseNode(node, i, originalContent, baseIndent)
		if err := updateNode(elemNode, value.Index(i)); err != nil {
			return fmt.Errorf("error updating sequence element %d: %w", i, err)
		}
		newContent = append(newContent, elemNode)
	}

	node.Content = newContent
	node.Style = originalStyle
	node.Column = originalColumn
	return nil
}

func createOrReuseNode(node *yaml.Node, index int, originalContent []*yaml.Node, baseIndent int) *yaml.Node {
	if index < len(originalContent) {
		return originalContent[index]
	}

	elemNode := &yaml.Node{}
	if len(originalContent) > 0 {
		lastNode := originalContent[len(originalContent)-1]
		elemNode.Style = lastNode.Style
		elemNode.Column = lastNode.Column
		elemNode.Line = lastNode.Line
	} else {
		elemNode.Column = node.Column + baseIndent
	}
	return elemNode
}

func updateMapping(node *yaml.Node, value reflect.Value) error {
	originalStyle := node.Style
	originalColumn := node.Column
	originalContent := node.Content

	node.Kind = yaml.MappingNode
	node.Tag = "!!map"
	if node.Content == nil {
		node.Content = []*yaml.Node{}
	}

	baseIndent := 2
	if len(originalContent) > 0 {
		baseIndent = originalContent[0].Column - node.Column
	}

	newContent := []*yaml.Node{}
	iter := value.MapRange()
	for iter.Next() {
		keyNode, valueNode := createOrReusePair(node, fmt.Sprintf("%v", iter.Key().Interface()), originalContent, baseIndent)
		if err := updateNode(valueNode, iter.Value()); err != nil {
			return fmt.Errorf("error updating map value: %w", err)
		}
		newContent = append(newContent, keyNode, valueNode)
	}

	node.Content = newContent
	node.Style = originalStyle
	node.Column = originalColumn
	return nil
}

func createOrReusePair(node *yaml.Node, key string, originalContent []*yaml.Node, baseIndent int) (*yaml.Node, *yaml.Node) {
	for i := 0; i < len(originalContent); i += 2 {
		if originalContent[i].Value == key {
			return originalContent[i], originalContent[i+1]
		}
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
		Tag:   "!!str",
	}
	valueNode := &yaml.Node{}

	if len(originalContent) > 0 {
		lastKey := originalContent[len(originalContent)-2]
		lastValue := originalContent[len(originalContent)-1]
		keyNode.Style = lastKey.Style
		keyNode.Column = lastKey.Column
		keyNode.Line = lastKey.Line
		valueNode.Style = lastValue.Style
		valueNode.Column = lastValue.Column
		valueNode.Line = lastValue.Line
	} else {
		keyNode.Column = node.Column + baseIndent
		valueNode.Column = node.Column + baseIndent
	}

	return keyNode, valueNode
}
