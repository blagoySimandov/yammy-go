package main

// Skill represents a skill with a name and level
type Skill struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
}

// University represents educational information
type University struct {
	Name    string              `yaml:"name"`
	Years   []int               `yaml:"years"`
	Courses map[string][]string `yaml:"courses"`
}

// Details represents personal details
type Details struct {
	Address string   `yaml:"address"`
	City    string   `yaml:"city"`
	Country string   `yaml:"country"`
	Phones  []string `yaml:"phones"`
}

// SkillSet represents a collection of skills
type SkillSet struct {
	Programming []Skill `yaml:"programming"`
	Languages   []Skill `yaml:"languages"`
}

// Education represents educational history
type Education struct {
	Universities []University `yaml:"universities"`
}

// Person represents a person's complete profile
type Person struct {
	Name      string    `yaml:"name"`
	Age       int       `yaml:"age"`
	Hobbies   []string  `yaml:"hobbies"`
	Details   Details   `yaml:"details"`
	Skills    SkillSet  `yaml:"skills"`
	Education Education `yaml:"education"`
}
