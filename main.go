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
	indent := detectIndentation(string(yamlData))
	fmt.Printf("Detected indentation: %d spaces\n", indent)

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

	fmt.Printf("Root node before: kind=%v, style=%v, column=%v\n", root.Kind, root.Style, root.Column)
	if len(root.Content) > 0 {
		fmt.Printf("Root content[0] before: kind=%v, style=%v, column=%v\n",
			root.Content[0].Kind, root.Content[0].Style, root.Content[0].Column)
	}

	if err := updateYamlFromStruct(&root, newData); err != nil {
		return fmt.Errorf("failed to update YAML: %w", err)
	}

	fmt.Printf("Root node after: kind=%v, style=%v, column=%v\n", root.Kind, root.Style, root.Column)
	if len(root.Content) > 0 {
		fmt.Printf("Root content[0] after: kind=%v, style=%v, column=%v\n",
			root.Content[0].Kind, root.Content[0].Style, root.Content[0].Column)
	}

	// Ensure root node and its content have no indentation
	root.Column = 0
	if len(root.Content) > 0 {
		root.Content[0].Column = 0
	}

	// Encode the updated YAML
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
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

// detectIndentation analyzes the YAML content to determine the indentation level
func detectIndentation(content string) int {
	lines := bytes.Split([]byte(content), []byte("\n"))
	for _, line := range lines {
		// Skip empty lines and lines without indentation
		if len(line) == 0 || line[0] != ' ' {
			continue
		}

		// Count leading spaces
		spaces := 0
		for _, ch := range line {
			if ch == ' ' {
				spaces++
			} else {
				break
			}
		}

		// Return the first non-zero indentation found
		if spaces > 0 {
			return spaces
		}
	}

	// Default to 2 spaces if no indentation is found
	return 2
}

// updateYamlFromStruct updates a YAML node tree with values from a struct while preserving
// comments and formatting
func updateYamlFromStruct(node *yaml.Node, data interface{}) error {
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
		// Get the root indentation offset
		rootOffset := mappingNode.Column
		// Reset root nodes to have no indentation
		node.Column = 0
		mappingNode.Column = 0
		// Adjust all child nodes relative to root
		adjustNodeColumns(mappingNode, rootOffset)
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
		if err := updateField(mappingNode, typ.Field(i), val.Field(i)); err != nil {
			return fmt.Errorf("failed to update field %s: %w", typ.Field(i).Name, err)
		}
	}

	return nil
}

// adjustNodeColumns adjusts all node columns by subtracting the offset
func adjustNodeColumns(node *yaml.Node, offset int) {
	if node.Column > offset {
		node.Column -= offset
	}
	for _, child := range node.Content {
		adjustNodeColumns(child, offset)
	}
}

// updateField updates a single field in a mapping node
func updateField(mappingNode *yaml.Node, fieldType reflect.StructField, fieldValue reflect.Value) error {
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
		// Copy style and column from an existing pair
		if len(mappingNode.Content) > 0 {
			keyNode.Style = mappingNode.Content[0].Style
			keyNode.Column = mappingNode.Content[0].Column
			valueNode.Style = mappingNode.Content[1].Style
			valueNode.Column = mappingNode.Content[1].Column
		} else {
			// If no existing content, use parent's column + 2
			keyNode.Column = mappingNode.Column + 2
			valueNode.Column = mappingNode.Column + 2
		}
		mappingNode.Content = append(mappingNode.Content, keyNode, valueNode)
	}

	return updateNode(valueNode, fieldValue)
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

// updateNode updates a node with a new value while preserving formatting
func updateNode(node *yaml.Node, value reflect.Value) error {
	// Store original style and column
	originalStyle := node.Style
	originalColumn := node.Column

	fmt.Printf("updateNode: kind=%v, style=%v, column=%v, value=%v\n", node.Kind, node.Style, node.Column, node.Value)

	switch value.Kind() {
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
	fmt.Printf("updateNode (after): kind=%v, style=%v, column=%v, value=%v\n", node.Kind, node.Style, node.Column, node.Value)
	return nil
}

// updateSequence updates a sequence node with values from a slice or array
func updateSequence(node *yaml.Node, value reflect.Value) error {
	// Store original properties
	originalStyle := node.Style
	originalColumn := node.Column
	originalContent := node.Content

	fmt.Printf("updateSequence: start style=%v, column=%v, content_len=%d\n", originalStyle, originalColumn, len(originalContent))

	// Set up sequence node
	node.Kind = yaml.SequenceNode
	node.Tag = "!!seq"
	if node.Content == nil {
		node.Content = []*yaml.Node{}
	}

	// Calculate base indentation for sequence items
	baseIndent := 2
	if len(originalContent) > 0 {
		baseIndent = originalContent[0].Column - node.Column
	}

	// Create new content
	newContent := make([]*yaml.Node, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		elemNode := createOrReuseNode(node, i, originalContent, baseIndent)
		if err := updateNode(elemNode, value.Index(i)); err != nil {
			return fmt.Errorf("error updating sequence element %d: %w", i, err)
		}
		newContent = append(newContent, elemNode)
	}

	// Update node
	node.Content = newContent
	node.Style = originalStyle
	node.Column = originalColumn
	fmt.Printf("updateSequence: end style=%v, column=%v, content_len=%d\n", node.Style, node.Column, len(node.Content))
	return nil
}

// createOrReuseNode creates a new node or reuses an existing one from a sequence
func createOrReuseNode(node *yaml.Node, index int, originalContent []*yaml.Node, baseIndent int) *yaml.Node {
	if index < len(originalContent) {
		fmt.Printf("createOrReuseNode: reusing node[%d] style=%v, column=%v\n", index, originalContent[index].Style, originalContent[index].Column)
		return originalContent[index]
	}

	elemNode := &yaml.Node{}
	if len(originalContent) > 0 {
		// Copy style and indentation from the last node
		lastNode := originalContent[len(originalContent)-1]
		elemNode.Style = lastNode.Style
		elemNode.Column = lastNode.Column
		elemNode.Line = lastNode.Line
		fmt.Printf("createOrReuseNode: new node copying from last style=%v, column=%v\n", lastNode.Style, lastNode.Column)
	} else {
		// If no reference nodes exist, use parent's column + baseIndent
		elemNode.Column = node.Column + baseIndent
		fmt.Printf("createOrReuseNode: new node using parent column=%v + baseIndent=%v\n", node.Column, baseIndent)
	}
	return elemNode
}

// updateMapping updates a mapping node with values from a map
func updateMapping(node *yaml.Node, value reflect.Value) error {
	// Store original properties
	originalStyle := node.Style
	originalColumn := node.Column
	originalContent := node.Content

	fmt.Printf("updateMapping: start style=%v, column=%v, content_len=%d\n", originalStyle, originalColumn, len(originalContent))

	// Set up mapping node
	node.Kind = yaml.MappingNode
	node.Tag = "!!map"
	if node.Content == nil {
		node.Content = []*yaml.Node{}
	}

	// Calculate base indentation for mapping items
	baseIndent := 2
	if len(originalContent) > 0 {
		baseIndent = originalContent[0].Column - node.Column
	}

	// Create new content
	newContent := []*yaml.Node{}
	iter := value.MapRange()
	for iter.Next() {
		keyNode, valueNode := createOrReusePair(node, fmt.Sprintf("%v", iter.Key().Interface()), originalContent, baseIndent)
		fmt.Printf("updateMapping: key=%v keyStyle=%v keyColumn=%v valueStyle=%v valueColumn=%v\n",
			keyNode.Value, keyNode.Style, keyNode.Column, valueNode.Style, valueNode.Column)
		if err := updateNode(valueNode, iter.Value()); err != nil {
			return fmt.Errorf("error updating map value: %w", err)
		}
		newContent = append(newContent, keyNode, valueNode)
	}

	// Update node
	node.Content = newContent
	node.Style = originalStyle
	node.Column = originalColumn
	fmt.Printf("updateMapping: end style=%v, column=%v, content_len=%d\n", node.Style, node.Column, len(node.Content))
	return nil
}

// createOrReusePair creates a new key-value pair or reuses existing nodes from a mapping
func createOrReusePair(node *yaml.Node, key string, originalContent []*yaml.Node, baseIndent int) (*yaml.Node, *yaml.Node) {
	// First try to find and reuse existing nodes
	for i := 0; i < len(originalContent); i += 2 {
		if originalContent[i].Value == key {
			fmt.Printf("createOrReusePair: reusing pair[%d] key=%v keyStyle=%v keyColumn=%v valueStyle=%v valueColumn=%v\n",
				i/2, key, originalContent[i].Style, originalContent[i].Column, originalContent[i+1].Style, originalContent[i+1].Column)
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
		// Copy style and indentation from the last pair
		lastKey := originalContent[len(originalContent)-2]
		lastValue := originalContent[len(originalContent)-1]
		keyNode.Style = lastKey.Style
		keyNode.Column = lastKey.Column
		keyNode.Line = lastKey.Line
		valueNode.Style = lastValue.Style
		valueNode.Column = lastValue.Column
		valueNode.Line = lastValue.Line
		fmt.Printf("createOrReusePair: new pair copying from last key=%v keyStyle=%v keyColumn=%v valueStyle=%v valueColumn=%v\n",
			key, keyNode.Style, keyNode.Column, valueNode.Style, valueNode.Column)
	} else {
		// If no reference nodes exist, use parent's column + baseIndent
		keyNode.Column = node.Column + baseIndent
		valueNode.Column = node.Column + baseIndent
		fmt.Printf("createOrReusePair: new pair using parent column=%v + baseIndent=%v\n", node.Column, baseIndent)
	}

	return keyNode, valueNode
}
