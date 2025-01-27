# yammy-go

A Go library for parsing YAML files while preserving comments and key order. This library is particularly useful when you need to read, modify, and write YAML files without losing the original formatting, comments, and order of keys.

## Features

- Preserves comments in YAML files
- Maintains the original order of keys
- Supports nested YAML structures
- Simple and intuitive API
- Includes marshaling and unmarshaling functionality

## Installation

```bash
go get github.com/blagoySimandov/yammy-go
```

## Usage

Here's a simple example of how to use the library:

```go
package main

import (
    "fmt"
    "strings"
    "github.com/blagoySimandov/yammy-go"
)

func main() {
    yamlContent := `# Configuration file
name: test-app  # Application name
version: 1.0.0  # Current version

# Database settings
database:
  host: localhost  # DB host
  port: 5432      # Default PostgreSQL port`

    parser := yammy.NewParser()
    root, err := parser.Parse(strings.NewReader(yamlContent))
    if err != nil {
        panic(err)
    }

    // Modify the YAML structure if needed

    // Marshal back to YAML
    output, err := yammy.Marshal(root)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(output))
}
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
