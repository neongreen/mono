package goutils_test

import (
	"github.com/neongreen/mono/dissect/pkg/goutils"
	"testing"
)

func TestNormalizeImports(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "single non-parenthesized import",
			input: `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}
`,
			expected: `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello")
}
`,
		},
		{
			name: "multiple non-parenthesized imports",
			input: `package main

import "fmt"
import "os"

func main() {
	fmt.Println("Hello")
	os.Exit(0)
}
`,
			expected: `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello")
	os.Exit(0)
}
`,
		},
		{
			name: "parenthesized import",
			input: `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello")
}
`,
			expected: `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello")
}
`,
		},
		{
			name: "mixed import styles",
			input: `package main

import "os"
import (
	"fmt"
)

func main() {
	fmt.Println("Hello")
	os.Exit(0)
}
`,
			expected: `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello")
	os.Exit(0)
}
`,
		},
		{
			name: "reordered imports",
			input: `package main

import (
	"os"
	"fmt"
)

func main() {
	fmt.Println("Hello")
	os.Exit(0)
}
`,
			expected: `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello")
	os.Exit(0)
}
`,
		},
		{
			name: "imports with aliases",
			input: `package main

import (
	myfmt "fmt"
	"os"
)

func main() {
	myfmt.Println("Hello")
	os.Exit(0)
}
`,
			expected: `package main

import (
	myfmt "fmt"
	"os"
)

func main() {
	myfmt.Println("Hello")
	os.Exit(0)
}
`,
		},
		{
			name: "no imports",
			input: `package main

func main() {
	println("Hello")
}
`,
			expected: `package main

func main() {
	println("Hello")
}
`,
		},
		{
			name: "unused imports (should not be removed by NormalizeImports)",
			input: `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello")
}
`,
			expected: `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello")
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := goutils.NormalizeImports(tt.input)
			if err != nil {
				t.Fatalf("NormalizeImports returned an error: %v", err)
			}
			if normalized != tt.expected {
				t.Errorf("NormalizeImports(%q) = %q, expected %q", tt.input, normalized, tt.expected)
			}
		})
	}
}
