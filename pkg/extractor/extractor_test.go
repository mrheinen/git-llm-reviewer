package extractor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTempFile creates a temporary file with the given content and extension
func createTempFile(t *testing.T, content string, ext string) string {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test*"+ext)
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}

	// Write content to the file
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Could not write to temp file: %v", err)
	}

	// Close the file
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Could not close temp file: %v", err)
	}

	// Return the file path
	return tmpfile.Name()
}

func TestExtractFunctionAtLine(t *testing.T) {
	// Test cases for Go language
	t.Run("Go Language Tests", func(t *testing.T) {
		// Create a Go language extractor
		extractor, err := NewCodeExtractor(Go, ".")
		if err != nil {
			t.Fatalf("Failed to create Go extractor: %v", err)
		}

		// Sample Go code with functions
		goCode := `package test

import "fmt"

// SimpleFunction is a test function
func SimpleFunction() {
	fmt.Println("This is a simple function")
}

// FunctionWithArgs has parameters and return values
func FunctionWithArgs(name string, age int) string {
	result := fmt.Sprintf("Name: %s, Age: %d", name, age)
	return result
}

// StructMethod is a method on a struct
type TestStruct struct {
	Name string
}

func (ts *TestStruct) StructMethod() {
	fmt.Println("Hello from", ts.Name)
}
`

		// Create a temporary file
		filePath := createTempFile(t, goCode, ".go")
		defer os.Remove(filePath) // Clean up after test

		// Test scenarios
		testCases := []struct {
			lineNumber   int
			shouldSucceed bool
			expectedSubstr string
			description  string
		}{
			{7, true, "SimpleFunction", "line inside first function"},
			{13, true, "FunctionWithArgs", "line inside second function"},
			{23, true, "StructMethod", "line inside struct method"},
			{2, false, "", "line outside any function"},
			{100, false, "", "line beyond file bounds"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result, err := extractor.ExtractFunctionAtLine(filePath, tc.lineNumber)
				
				if tc.shouldSucceed {
					if err != nil {
						t.Errorf("Expected to extract function at line %d but got error: %v", tc.lineNumber, err)
					}
					if result == "" {
						t.Errorf("Expected non-empty result for line %d", tc.lineNumber)
					}
					if !strings.Contains(result, tc.expectedSubstr) {
						t.Errorf("Expected result to contain %q but got: %s", tc.expectedSubstr, result)
					}
				} else {
					if err == nil {
						t.Errorf("Expected error for line %d but got none", tc.lineNumber)
					}
					if result != "" {
						t.Errorf("Expected empty result for line %d but got: %s", tc.lineNumber, result)
					}
				}
			})
		}
	})

	// Test cases for JavaScript language
	t.Run("JavaScript Language Tests", func(t *testing.T) {
		// Create a JavaScript language extractor
		extractor, err := NewCodeExtractor(JavaScript, ".")
		if err != nil {
			t.Fatalf("Failed to create JavaScript extractor: %v", err)
		}

		// Sample JavaScript code with functions
		jsCode := `// Regular function
function greet(name) {
  console.log("Hello, " + name);
  return "Greeting sent";
}

// Arrow function
const multiply = (a, b) => {
  return a * b;
};

// Class with method
class Calculator {
  constructor(initialValue = 0) {
    this.value = initialValue;
  }
  
  add(x) {
    this.value += x;
    return this.value;
  }
}
`

		// Create a temporary file
		filePath := createTempFile(t, jsCode, ".js")
		defer os.Remove(filePath) // Clean up after test

		// Test scenarios
		testCases := []struct {
			lineNumber   int
			shouldSucceed bool
			expectedSubstr string
			description  string
		}{
			{3, true, "greet", "line inside regular function"},
			{9, true, "return a * b", "line inside arrow function"}, // Arrow functions may not include the variable name in extraction
			{18, true, "add", "line inside class method"},
			{1, false, "", "line outside any function"},
			{50, false, "", "line beyond file bounds"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result, err := extractor.ExtractFunctionAtLine(filePath, tc.lineNumber)
				
				if tc.shouldSucceed {
					if err != nil {
						t.Errorf("Expected to extract function at line %d but got error: %v", tc.lineNumber, err)
					}
					if result == "" {
						t.Errorf("Expected non-empty result for line %d", tc.lineNumber)
					}
					if !strings.Contains(result, tc.expectedSubstr) {
						t.Errorf("Expected result to contain %q but got: %s", tc.expectedSubstr, result)
					}
				} else {
					if err == nil {
						t.Errorf("Expected error for line %d but got none", tc.lineNumber)
					}
					if result != "" {
						t.Errorf("Expected empty result for line %d but got: %s", tc.lineNumber, result)
					}
				}
			})
		}
	})

	// Test cases for Python language
	t.Run("Python Language Tests", func(t *testing.T) {
		// Create a Python language extractor
		extractor, err := NewCodeExtractor(Python, ".")
		if err != nil {
			t.Fatalf("Failed to create Python extractor: %v", err)
		}

		// Sample Python code with functions
		pyCode := `def greet(name):
    """Say hello to someone"""
    print(f"Hello, {name}")
    return f"Greeted {name}"

# Class with methods
class Person:
    def __init__(self, name, age):
        self.name = name
        self.age = age
        
    def introduce(self):
        return f"My name is {self.name} and I am {self.age} years old"
        
    def celebrate_birthday(self):
        self.age += 1
        return f"Happy birthday! Now I am {self.age} years old"
`

		// Create a temporary file
		filePath := createTempFile(t, pyCode, ".py")
		defer os.Remove(filePath) // Clean up after test

		// Test scenarios
		testCases := []struct {
			lineNumber   int
			shouldSucceed bool
			expectedSubstr string
			description  string
		}{
			{3, true, "greet", "line inside function"},
			{9, true, "__init__", "line inside constructor method"},
			{13, true, "introduce", "line inside class method"},
			{6, false, "", "line outside any function"},
			{50, false, "", "line beyond file bounds"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result, err := extractor.ExtractFunctionAtLine(filePath, tc.lineNumber)
				
				if tc.shouldSucceed {
					if err != nil {
						t.Errorf("Expected to extract function at line %d but got error: %v", tc.lineNumber, err)
					}
					if result == "" {
						t.Errorf("Expected non-empty result for line %d", tc.lineNumber)
					}
					if !strings.Contains(result, tc.expectedSubstr) {
						t.Errorf("Expected result to contain %q but got: %s", tc.expectedSubstr, result)
					}
				} else {
					if err == nil {
						t.Errorf("Expected error for line %d but got none", tc.lineNumber)
					}
					if result != "" {
						t.Errorf("Expected empty result for line %d but got: %s", tc.lineNumber, result)
					}
				}
			})
		}
	})

	// Test with non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		extractor, _ := NewCodeExtractor(Go, ".")
		nonExistentPath := filepath.Join(os.TempDir(), "non_existent_file.go")
		
		result, err := extractor.ExtractFunctionAtLine(nonExistentPath, 1)
		
		if err == nil {
			t.Error("Expected error for non-existent file but got none")
		}
		if result != "" {
			t.Errorf("Expected empty result for non-existent file but got: %s", result)
		}
	})
}

func TestExtractTypeByName(t *testing.T) {
	// Test cases for Go language
	t.Run("Go Language Tests", func(t *testing.T) {
		// Create a Go language extractor
		extractor, err := NewCodeExtractor(Go, ".")
		if err != nil {
			t.Fatalf("Failed to create Go extractor: %v", err)
		}

		// Read test_types.go as our test file
		content, err := os.ReadFile("test_types.go")
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		// Test cases
		testCases := []struct {
			typeName    string
			shouldExist bool
			description string
		}{
			{"Request", true, "struct type that exists"},
			{"Response", true, "struct type that exists"},
			{"Server", true, "interface type that exists"},
			{"SimpleServer", true, "struct type that exists"},
			{"NonExistentType", false, "type that doesn't exist"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result, err := extractor.ExtractTypeByName(content, tc.typeName)
				
				if tc.shouldExist {
					if err != nil {
						t.Errorf("Expected to find type %s but got error: %v", tc.typeName, err)
					}
					if result == "" {
						t.Errorf("Expected non-empty result for type %s", tc.typeName)
					}
				} else {
					if err == nil {
						t.Errorf("Expected error for non-existent type %s but got none", tc.typeName)
					}
					if result != "" {
						t.Errorf("Expected empty result for non-existent type %s but got: %s", tc.typeName, result)
					}
				}
			})
		}
	})

	// Test cases for JavaScript language
	t.Run("JavaScript Language Tests", func(t *testing.T) {
		// Create a JavaScript language extractor
		extractor, err := NewCodeExtractor(JavaScript, ".")
		if err != nil {
			t.Fatalf("Failed to create JavaScript extractor: %v", err)
		}

		// Sample JavaScript code with types - using only class definitions since that's what our extractor supports best
		jsCode := []byte(`
// User class for authentication
class User {
  constructor(name, email) {
    this.name = name;
    this.email = email;
  }
  
  getInfo() {
    return this.name + ' (' + this.email + ')';
  }
}

// Authentication provider class
class AuthProvider {
  login(credentials) {
    // Login implementation
    return new User(credentials.username, 'user@example.com');
  }
  
  logout() {
    // Logout implementation
  }
}

// User credentials class
class UserCredentials {
  constructor(username, password) {
    this.username = username;
    this.password = password;
  }
}
`)

		// Test cases
		testCases := []struct {
			typeName    string
			shouldExist bool
			description string
		}{
			{"User", true, "class that exists"},
			{"AuthProvider", true, "class that exists"},
			{"UserCredentials", true, "class that exists"},
			{"NonExistentType", false, "type that doesn't exist"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result, err := extractor.ExtractTypeByName(jsCode, tc.typeName)
				
				if tc.shouldExist {
					if err != nil {
						t.Errorf("Expected to find type %s but got error: %v", tc.typeName, err)
					}
					if result == "" {
						t.Errorf("Expected non-empty result for type %s", tc.typeName)
					}
				} else {
					if err == nil {
						t.Errorf("Expected error for non-existent type %s but got none", tc.typeName)
					}
					if result != "" {
						t.Errorf("Expected empty result for non-existent type %s but got: %s", tc.typeName, result)
					}
				}
			})
		}
	})

	// Test cases for Python language
	t.Run("Python Language Tests", func(t *testing.T) {
		// Create a Python language extractor
		extractor, err := NewCodeExtractor(Python, ".")
		if err != nil {
			t.Fatalf("Failed to create Python extractor: %v", err)
		}

		// Sample Python code with classes
		pyCode := []byte(`
class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        pass

class Dog(Animal):
    def speak(self):
        return "Woof!"

class Cat(Animal):
    def __init__(self, name, color):
        super().__init__(name)
        self.color = color
        
    def speak(self):
        return "Meow!"
`)

		// Test cases
		testCases := []struct {
			typeName    string
			shouldExist bool
			description string
		}{
			{"Animal", true, "class that exists"},
			{"Dog", true, "derived class that exists"},
			{"Cat", true, "derived class with constructor that exists"},
			{"NonExistentType", false, "class that doesn't exist"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result, err := extractor.ExtractTypeByName(pyCode, tc.typeName)
				
				if tc.shouldExist {
					if err != nil {
						t.Errorf("Expected to find class %s but got error: %v", tc.typeName, err)
					}
					if result == "" {
						t.Errorf("Expected non-empty result for class %s", tc.typeName)
					}
				} else {
					if err == nil {
						t.Errorf("Expected error for non-existent class %s but got none", tc.typeName)
					}
					if result != "" {
						t.Errorf("Expected empty result for non-existent class %s but got: %s", tc.typeName, result)
					}
				}
			})
		}
	})

	// Test with empty content
	t.Run("Empty Content Test", func(t *testing.T) {
		extractor, _ := NewCodeExtractor(Go, ".")
		result, err := extractor.ExtractTypeByName([]byte(""), "AnyType")
		
		if err == nil {
			t.Error("Expected error for empty content but got none")
		}
		if result != "" {
			t.Errorf("Expected empty result for empty content but got: %s", result)
		}
	})
}
