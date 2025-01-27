package main

import (
	"fmt"
	"os"

	"github.com/blagoySimandov/yammy-go/internal/yaml"
)

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

func processFile(file string) error {
	yamlData, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	newData := Person{
		Name:    "John",
		Age:     31,
		Hobbies: []string{"reading", "gaming", "hiking"},
		Details: Details{
			Address: "123 Elm Street",
			City:    "Gotham",
			Country: "Wonderland",
			Phones:  []string{"555-0123", "555-9999"},
		},
		Skills: SkillSet{
			Programming: []Skill{
				{Name: "Go", Level: "Expert"},
				{Name: "Python", Level: "Intermediate"},
			},
			Languages: []Skill{
				{Name: "English", Level: "Native"},
				{Name: "Spanish", Level: "Beginner"},
			},
		},
		Education: Education{
			Universities: []University{
				{
					Name:  "Tech University",
					Years: []int{2015, 2020},
					Courses: map[string][]string{
						"CS101": {"A", "B+", "A"},
						"CS102": {"B+", "A"},
					},
				},
			},
		},
	}

	updatedYAML, err := yaml.UpdateYAML(yamlData, newData)
	if err != nil {
		return fmt.Errorf("failed to update YAML: %w", err)
	}

	outputFile := "updated_" + file
	if err := os.WriteFile(outputFile, updatedYAML, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
