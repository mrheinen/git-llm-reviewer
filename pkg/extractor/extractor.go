package extractor

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
)

// Language types supported by this extractor
type Language string

const (
	Go         Language = "go"
	JavaScript Language = "javascript"
	Python     Language = "python"
)

// Configure function node types based on language
var functionNodeTypes = map[Language][]string{
	Go:         {"function_declaration", "method_declaration"},
	JavaScript: {"function", "function_declaration", "method_definition", "arrow_function"},
	Python:     {"function_definition", "class_definition", "method_definition"},
}

// Configure type node types based on language
var typeNodeTypes = map[Language][]string{
	Go:         {"type_declaration", "type_spec"},
	JavaScript: {"class_declaration", "interface_declaration", "type_alias_declaration"},
	Python:     {"class_definition"},
}

// CodeExtractor extracts code elements from source code files
type CodeExtractor struct {
	Language Language
	parser   *sitter.Parser
	dirPath  string
}

// NewCodeExtractor creates a new code extractor for the specified language
func NewCodeExtractor(lang Language, dirPath string) (*CodeExtractor, error) {
	parser := sitter.NewParser()

	// Set language parser based on language type
	switch lang {
	case Go:
		parser.SetLanguage(golang.GetLanguage())
	case JavaScript:
		parser.SetLanguage(javascript.GetLanguage())
	case Python:
		parser.SetLanguage(python.GetLanguage())
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	return &CodeExtractor{
		Language: lang,
		parser:   parser,
		dirPath:  dirPath,
	}, nil
}

// ExtractFunctionAtLine extracts a function containing the specified line number from a code file
func (ce *CodeExtractor) ExtractFunctionAtLine(filePath string, lineNumber int) (string, error) {
	// Read the file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Count lines in the file
	lines := bytes.Count(content, []byte{'\n'}) + 1

	// Validate line number
	if lineNumber < 1 || lineNumber > lines {
		return "", fmt.Errorf("line number %d is out of bounds (file has %d lines)", lineNumber, lines)
	}

	// Parse the code into an AST
	tree := ce.parser.Parse(nil, content)
	rootNode := tree.RootNode()

	// Convert 1-indexed line number to 0-indexed for internal processing
	zeroBasedLineNumber := lineNumber - 1

	// Find the function node that contains the specified line
	var functionNode *sitter.Node
	var smallestSize int = -1

	// Recursive function to find a node at a specific position
	var findNodeAtLine func(node *sitter.Node)
	findNodeAtLine = func(node *sitter.Node) {
		// Get the node's start and end positions
		startRow := node.StartPoint().Row
		endRow := node.EndPoint().Row

		// Check if the node contains the target line
		if int(startRow) <= zeroBasedLineNumber && zeroBasedLineNumber <= int(endRow) {
			// Check if this is a function node
			nodeType := node.Type()

			isFunctionNode := false
			for _, funcType := range functionNodeTypes[ce.Language] {
				if nodeType == funcType {
					isFunctionNode = true
					break
				}
			}

			if isFunctionNode {
				// Calculate node size
				nodeSize := int(endRow - startRow)

				// Keep the smallest (most specific) function node
				if functionNode == nil || nodeSize < smallestSize {
					functionNode = node
					smallestSize = nodeSize
				}
			}

			// Continue searching in children
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				findNodeAtLine(child)
			}
		}
	}

	// Start the search from the root node
	findNodeAtLine(rootNode)

	// If we found a function node, extract its text
	if functionNode != nil {
		startIndex := functionNode.StartByte()
		endIndex := functionNode.EndByte()
		return string(content[startIndex:endIndex]), nil
	}

	return "", fmt.Errorf("no function found containing line %d", lineNumber)
}

// DetectLanguage attempts to detect the language from the file extension
func DetectLanguage(filePath string) Language {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		return Go
	case ".js", ".jsx", ".ts", ".tsx":
		return JavaScript
	case ".py":
		return Python
	default:
		// Default to Go if extension not recognized
		return Go
	}
}

// ExtractTypeByName extracts a type definition by its name
func (ce *CodeExtractor) ExtractTypeByName(content []byte, typeName string) (string, error) {
	// Parse the code into an AST
	tree := ce.parser.Parse(nil, content)
	rootNode := tree.RootNode()

	// Find a type node with the specified name
	var typeNode *sitter.Node

	// Recursive function to traverse the AST
	var findTypeByName func(node *sitter.Node)
	findTypeByName = func(node *sitter.Node) {
		if typeNode != nil {
			return // Stop if we've already found the type
		}

		// Check if this is a type node
		nodeType := node.Type()

		isTypeNode := false
		for _, typeNodeType := range typeNodeTypes[ce.Language] {
			if nodeType == typeNodeType {
				isTypeNode = true
				break
			}
		}

		if isTypeNode {
			// Type declarations can be complex; we need to find the identifier with the name we're looking for
			var nameNode *sitter.Node

			// Different languages have different node structures
			switch ce.Language {
			case Go:
				if nodeType == "type_declaration" {
					// For Go type_declaration, we need to find the type_spec with the right name
					for i := 0; i < int(node.ChildCount()); i++ {
						child := node.Child(i)
						if child.Type() == "type_spec" {
							// Look for the name in the children
							for j := 0; j < int(child.ChildCount()); j++ {
								grandchild := child.Child(j)
								if grandchild.Type() == "type_identifier" && string(content[grandchild.StartByte():grandchild.EndByte()]) == typeName {
									typeNode = node
									return
								}
							}
						}
					}
				} else if nodeType == "type_spec" {
					// Look directly for the identifier
					for i := 0; i < int(node.ChildCount()); i++ {
						child := node.Child(i)
						if child.Type() == "type_identifier" && string(content[child.StartByte():child.EndByte()]) == typeName {
							typeNode = node.Parent() // Get the whole type_declaration
							return
						}
					}
				}
			case JavaScript, Python:
				// For JavaScript and Python, find the identifier/name node
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if (child.Type() == "identifier" || child.Type() == "property_identifier" ||
						child.Type() == "identifier" || child.Type() == "class_name") &&
						string(content[child.StartByte():child.EndByte()]) == typeName {

						nameNode = child
						break
					}
				}

				if nameNode != nil {
					typeNode = node
					return
				}
			}
		}

		// Continue search in children
		for i := 0; i < int(node.ChildCount()) && typeNode == nil; i++ {
			child := node.Child(i)
			findTypeByName(child)
		}
	}

	// Start search from root
	findTypeByName(rootNode)

	// If we found the type node, extract its text
	if typeNode != nil {
		startIndex := typeNode.StartByte()
		endIndex := typeNode.EndByte()
		return string(content[startIndex:endIndex]), nil
	}

	return "", fmt.Errorf("type '%s' not found", typeName)
}

// FindDefinitionForType searches for and extracts a type definition by name in the directory specified in the CodeExtractor
func (ce *CodeExtractor) FindDefinitionForType(typeName string) (string, string, error) {
	var result string
	var foundFilePath string

	fmt.Println("Searching for type definition:", typeName)

	err := filepath.WalkDir(ce.dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process files that match the language extension
		ext := strings.ToLower(filepath.Ext(path))
		langMatch := false
		switch ce.Language {
		case Go:
			langMatch = ext == ".go"
		case JavaScript:
			langMatch = ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx"
		case Python:
			langMatch = ext == ".py"
		}

		if !langMatch {
			return nil
		}

		// Read the file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		// Look for the type in this file
		extracted, err := ce.ExtractTypeByName(content, typeName)
		if err == nil {
			// Found it
			result = extracted
			foundFilePath = path
			return filepath.SkipAll // Stop walking
		}

		return nil
	})

	if result == "" {
		return "", "", fmt.Errorf("no type named '%s' found in directory %s", typeName, ce.dirPath)
	}

	return result, foundFilePath, err
}
