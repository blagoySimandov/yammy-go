package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Struct definitions for YAML data
type Skill struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
}

type University struct {
	Name    string              `yaml:"name"`
	Years   []int               `yaml:"years"`
	Courses map[string][]string `yaml:"courses"`
}

type Details struct {
	Address string   `yaml:"address"`
	City    string   `yaml:"city"`
	Country string   `yaml:"country"`
	Phones  []string `yaml:"phones"`
}

type SkillSet struct {
	Programming []Skill `yaml:"programming"`
	Languages   []Skill `yaml:"languages"`
}

type Education struct {
	Universities []University `yaml:"universities"`
}

type Person struct {
	Name      string    `yaml:"name"`
	Age       int       `yaml:"age"`
	Hobbies   []string  `yaml:"hobbies"`
	Details   Details   `yaml:"details"`
	Skills    SkillSet  `yaml:"skills"`
	Education Education `yaml:"education"`
}

func main() {
	// Process both 2-space and 4-space files
	files := []string{"test.yaml"}
	for _, file := range files {
		if err := processFile(file); err != nil {
			fmt.Printf("Error processing %s: %v\n", file, err)
			continue
		}
		fmt.Printf("Updated YAML has been written to updated_%s\n", file)
	}
}

// processFile reads a YAML file, updates its contents, and writes the result to a new file
func processFile(file string) error {
	// Read the YAML file
	yamlData, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Detect indentation from the file content
	lines := bytes.Split(yamlData, []byte("\n"))
	baseIndent := 2
	for _, line := range lines {
		if len(line) > 0 && line[0] == ' ' {
			// Count leading spaces
			spaces := 0
			for _, c := range line {
				if c != ' ' {
					break
				}
				spaces++
			}
			if spaces > 0 {
				baseIndent = spaces
				break
			}
		}
	}

	// Create new data with minimal changes
	newData := Person{
		Name:    "John",                                  // Same name
		Age:     31,                                      // Just increment age
		Hobbies: []string{"reading", "gaming", "hiking"}, // Only change last hobby
		Details: Details{
			Address: "123 Elm Street",                 // Same address
			City:    "Gotham",                         // Same city
			Country: "Wonderland",                     // Same country
			Phones:  []string{"555-0123", "555-9999"}, // Only change second phone
		},
		Skills: SkillSet{
			Programming: []Skill{
				{Name: "Go", Level: "Expert"},           // Only change level
				{Name: "Python", Level: "Intermediate"}, // Same as original
			},
			Languages: []Skill{
				{Name: "English", Level: "Native"},   // Same as original
				{Name: "Spanish", Level: "Beginner"}, // Same as original
			},
		},
		Education: Education{
			Universities: []University{
				{
					Name:  "Tech University", // Same name
					Years: []int{2015, 2020}, // Only change end year
					Courses: map[string][]string{
						"CS101": {"A", "B+", "A"}, // Only change last grade
						"CS102": {"B+", "A"},      // Same as original
					},
				},
			},
		},
	}

	// Parse and update the YAML
	var root yaml.Node
	if err := yaml.Unmarshal(yamlData, &root); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := updateYamlFromStruct(&root, newData, baseIndent); err != nil {
		return fmt.Errorf("failed to update YAML: %w", err)
	}

	// Encode the updated YAML
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(baseIndent)
	if err := enc.Encode(&root); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	// Write to output file
	outputFile := "updated_" + file
	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// updateYamlFromStruct updates a YAML node tree with values from a struct while preserving
// comments and formatting. The function handles document nodes and mapping nodes.
func updateYamlFromStruct(node *yaml.Node, data interface{}, baseIndent int) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct or pointer to struct")
	}

	// Get the mapping node
	mappingNode := node
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) != 1 {
			return fmt.Errorf("invalid YAML structure: document node should have exactly one child")
		}
		mappingNode = node.Content[0]
	}

	// Initialize or validate mapping node
	if mappingNode.Kind != yaml.MappingNode {
		mappingNode.Kind = yaml.MappingNode
		mappingNode.Tag = "!!map"
	}
	if mappingNode.Content == nil {
		mappingNode.Content = []*yaml.Node{}
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		if err := updateField(mappingNode, typ.Field(i), val.Field(i), baseIndent); err != nil {
			return fmt.Errorf("failed to update field %s: %w", typ.Field(i).Name, err)
		}
	}

	return nil
}

// updateField updates a single field in a mapping node
func updateField(mappingNode *yaml.Node, fieldType reflect.StructField, fieldValue reflect.Value, baseIndent int) error {
	yamlTag := fieldType.Tag.Get("yaml")
	if yamlTag == "" {
		yamlTag = fieldType.Name
	}

	// Find or create nodes
	keyNode, valueNode, found := findNodes(mappingNode, yamlTag)
	if !found {
		keyNode = &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: yamlTag,
		}
		valueNode = &yaml.Node{}
		mappingNode.Content = append(mappingNode.Content, keyNode, valueNode)
	}

	// Set indentation for new nodes
	if !found {
		keyNode.Column = mappingNode.Column + baseIndent
		valueNode.Column = mappingNode.Column + baseIndent
	}

	return updateNode(valueNode, fieldValue, baseIndent)
}

// findNodes looks for a key-value pair in a mapping node
func findNodes(mappingNode *yaml.Node, key string) (keyNode, valueNode *yaml.Node, found bool) {
	for i := 0; i < len(mappingNode.Content); i += 2 {
		if mappingNode.Content[i].Value == key {
			return mappingNode.Content[i], mappingNode.Content[i+1], true
		}
	}
	return nil, nil, false
}

// updateNode updates a node with a new value, handling different types appropriately
func updateNode(node *yaml.Node, value reflect.Value, baseIndent int) error {
	// Store original style and column
	originalStyle := node.Style
	originalColumn := node.Column

	switch value.Kind() {
	case reflect.Struct:
		if err := updateYamlFromStruct(node, value.Interface(), baseIndent); err != nil {
			return err
		}
	case reflect.Slice, reflect.Array:
		if err := updateSequence(node, value, baseIndent); err != nil {
			return err
		}
	case reflect.Map:
		if err := updateMapping(node, value, baseIndent); err != nil {
			return err
		}
	default:
		node.Kind = yaml.ScalarNode
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			node.Tag = "!!int"
		case reflect.Float32, reflect.Float64:
			node.Tag = "!!float"
		case reflect.Bool:
			node.Tag = "!!bool"
		default:
			node.Tag = "!!str"
		}
		node.Value = fmt.Sprintf("%v", value.Interface())
	}

	// Restore original style and column
	node.Style = originalStyle
	node.Column = originalColumn
	return nil
}

// updateSequence updates a sequence node with values from a slice or array
func updateSequence(node *yaml.Node, value reflect.Value, baseIndent int) error {
	// Store original properties
	originalStyle := node.Style
	originalColumn := node.Column
	originalContent := node.Content

	// Set up sequence node
	node.Kind = yaml.SequenceNode
	node.Tag = "!!seq"
	if node.Content == nil {
		node.Content = []*yaml.Node{}
	}

	// Create new content
	newContent := make([]*yaml.Node, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		elemNode := createOrReuseNode(node, i, originalContent, baseIndent)
		if err := updateNode(elemNode, value.Index(i), baseIndent); err != nil {
			return fmt.Errorf("error updating sequence element %d: %w", i, err)
		}
		newContent = append(newContent, elemNode)
	}

	// Update node
	node.Content = newContent
	node.Style = originalStyle
	node.Column = originalColumn
	return nil
}

// createOrReuseNode creates a new node or reuses an existing one from a sequence
func createOrReuseNode(node *yaml.Node, index int, originalContent []*yaml.Node, baseIndent int) *yaml.Node {
	if index < len(originalContent) {
		return originalContent[index]
	}

	elemNode := &yaml.Node{}
	if len(originalContent) > 0 {
		// Copy style and column from the last node of the original content
		lastNode := originalContent[len(originalContent)-1]
		elemNode.Style = lastNode.Style
		elemNode.Column = lastNode.Column
	} else if len(node.Content) > 0 {
		// Fallback to current content if no original content
		lastNode := node.Content[len(node.Content)-1]
		elemNode.Style = lastNode.Style
		elemNode.Column = lastNode.Column
	} else {
		// If no reference nodes exist, use parent's column + baseIndent
		elemNode.Column = node.Column + baseIndent
	}
	return elemNode
}

// updateMapping updates a mapping node with values from a map
func updateMapping(node *yaml.Node, value reflect.Value, baseIndent int) error {
	// Store original properties
	originalStyle := node.Style
	originalColumn := node.Column
	originalContent := node.Content

	// Set up mapping node
	node.Kind = yaml.MappingNode
	node.Tag = "!!map"
	if node.Content == nil {
		node.Content = []*yaml.Node{}
	}

	// Create new content
	newContent := []*yaml.Node{}
	iter := value.MapRange()
	for iter.Next() {
		keyNode, valueNode := createOrReusePair(node, fmt.Sprintf("%v", iter.Key().Interface()), originalContent, baseIndent)
		if err := updateNode(valueNode, iter.Value(), baseIndent); err != nil {
			return fmt.Errorf("error updating map value: %w", err)
		}
		newContent = append(newContent, keyNode, valueNode)
	}

	// Update node
	node.Content = newContent
	node.Style = originalStyle
	node.Column = originalColumn
	return nil
}

// createOrReusePair creates a new key-value pair or reuses existing nodes from a mapping
func createOrReusePair(node *yaml.Node, key string, originalContent []*yaml.Node, baseIndent int) (*yaml.Node, *yaml.Node) {
	// First try to find and reuse existing nodes from original content
	for i := 0; i < len(originalContent); i += 2 {
		if originalContent[i].Value == key {
			return originalContent[i], originalContent[i+1]
		}
	}

	// Then try current content
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i], node.Content[i+1]
		}
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
		Tag:   "!!str",
	}
	valueNode := &yaml.Node{}

	if len(originalContent) > 0 {
		// Copy style and column from the last pair in original content
		lastKey := originalContent[len(originalContent)-2]
		lastValue := originalContent[len(originalContent)-1]
		keyNode.Style = lastKey.Style
		keyNode.Column = lastKey.Column
		valueNode.Style = lastValue.Style
		valueNode.Column = lastValue.Column
	} else if len(node.Content) > 0 {
		// Fallback to current content
		lastKey := node.Content[len(node.Content)-2]
		lastValue := node.Content[len(node.Content)-1]
		keyNode.Style = lastKey.Style
		keyNode.Column = lastKey.Column
		valueNode.Style = lastValue.Style
		valueNode.Column = lastValue.Column
	} else {
		// If no reference nodes exist, use parent's column + baseIndent
		keyNode.Column = node.Column + baseIndent
		valueNode.Column = node.Column + baseIndent
	}

	return keyNode, valueNode
}
