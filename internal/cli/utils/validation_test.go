package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "test-project", false},
		{"valid with underscore", "test_project", false},
		{"valid with numbers", "project123", false},
		{"valid mixed", "my-awesome_project123", false},
		{"empty name", "", true},
		{"too short", "a", true},
		{"too long", "this_is_a_very_long_project_name_that_exceeds_fifty_characters", true},
		{"starts with number", "123project", true},
		{"starts with special char", "-project", true},
		{"contains spaces", "test project", true},
		{"contains dots", "test.project", true},
		{"reserved name", "main", true},
		{"reserved name case", "MAIN", true},
		{"reserved name", "go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateGoModule(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid github module", "github.com/user/project", false},
		{"valid gitlab module", "gitlab.com/user/project", false},
		{"valid custom domain", "example.com/project", false},
		{"simple name", "project", true}, // Warning, should not be fatal
		{"empty name", "", true},
		{"invalid characters", "github.com/user/project!", true},
		{"starts with special", "/invalid", true},
		{"ends with special", "invalid/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGoModule(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGoModule(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDirectoryName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "project", false},
		{"valid with dash", "my-project", false},
		{"valid with underscore", "my_project", false},
		{"empty name", "", true},
		{"contains invalid char", "project<test", true},
		{"contains pipe", "project|test", true},
		{"only dots", "...", true},
		{"only dot", ".", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectoryName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectoryName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "project", "project"},
		{"name with spaces", "my project", "my_project"},
		{"name with dashes", "my-project", "my_project"},
		{"name with special chars", "my@project#test", "my_project_test"},
		{"starts with number", "123project", "Aproject"},
		{"multiple spaces", "my   project", "my_project"},
		{"mixed special chars", "my-project_test 123", "my_project_test_123"},
		{"empty string", "", "project"},
		{"only special chars", "@#$%", "project"},
		{"ends with underscore", "test_", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckDirectoryExists(t *testing.T) {
	// 创建临时目录和文件
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")
	testFile := filepath.Join(tmpDir, "testfile")

	// 创建目录和文件
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
		wantErr  bool
	}{
		{"existing directory", testDir, true, false},
		{"non-existing directory", filepath.Join(tmpDir, "nonexistent"), false, false},
		{"existing file", testFile, false, true}, // File exists but is not a directory
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := CheckDirectoryExists(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDirectoryExists(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if exists != tt.expected {
				t.Errorf("CheckDirectoryExists(%q) = %v, expected %v", tt.path, exists, tt.expected)
			}
		})
	}
}

func TestIsDirectoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建空目录
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	// 创建有文件的目录
	nonEmptyDir := filepath.Join(tmpDir, "nonempty")
	if err := os.Mkdir(nonEmptyDir, 0755); err != nil {
		t.Fatalf("Failed to create non-empty directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建只有隐藏文件的目录
	hiddenDir := filepath.Join(tmpDir, "hidden")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, ".hidden"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
		wantErr  bool
	}{
		{"empty directory", emptyDir, true, false},
		{"non-empty directory", nonEmptyDir, false, false},
		{"directory with only hidden files", hiddenDir, true, false},
		{"non-existent directory", filepath.Join(tmpDir, "nonexistent"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty, err := IsDirectoryEmpty(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDirectoryEmpty(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if isEmpty != tt.expected {
				t.Errorf("IsDirectoryEmpty(%q) = %v, expected %v", tt.path, isEmpty, tt.expected)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid OpenAI key", "sk-1234567890abcdef1234567890abcdef", false},
		{"valid OpenRouter key", "sk-or-v1-1234567890abcdef1234567890abcdef", false},
		{"generic valid key", "abcdef1234567890", false},
		{"empty key", "", true},
		{"too short OpenAI key", "sk-123", true},
		{"too short OpenRouter key", "sk-or-123", true},
		{"too short generic key", "123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKey(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
		})
	}
}

func TestGenerateGoModule(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "myproject", "github.com/username/myproject"},
		{"name with dashes", "my-project", "github.com/username/my-project"},
		{"name with underscores", "my_project", "github.com/username/my_project"},
		{"name with spaces", "my project", "github.com/username/my_project"},
		{"complex name", "My@Project#123", "github.com/username/my_project_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateGoModule(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateGoModule(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple word", "hello", "Hello"},
		{"snake case", "hello_world", "HelloWorld"},
		{"kebab case", "hello-world", "HelloWorld"},
		{"spaces", "hello world", "HelloWorld"},
		{"mixed", "hello_world-test 123", "HelloWorldTest123"},
		{"already pascal", "HelloWorld", "HelloWorld"},
		{"single char", "a", "A"},
		{"empty", "", ""},
		{"numbers", "test123", "Test123"},
		{"multiple separators", "hello__world--test", "HelloWorldTest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple word", "hello", "hello"},
		{"snake case", "hello_world", "helloWorld"},
		{"kebab case", "hello-world", "helloWorld"},
		{"spaces", "hello world", "helloWorld"},
		{"mixed", "hello_world-test", "helloWorldTest"},
		{"single char", "a", "a"},
		{"empty", "", ""},
		{"already camel", "helloWorld", "helloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToCamelCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple word", "hello", "hello"},
		{"pascal case", "HelloWorld", "hello_world"},
		{"camel case", "helloWorld", "hello_world"},
		{"kebab case", "hello-world", "hello_world"},
		{"spaces", "hello world", "hello_world"},
		{"mixed", "HelloWorld-Test 123", "hello_world_test_123"},
		{"single char", "A", "a"},
		{"empty", "", ""},
		{"already snake", "hello_world", "hello_world"},
		{"multiple caps", "XMLHttpRequest", "xml_http_request"},
		{"numbers", "Test123ABC", "test123_abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "Project", "project"},
		{"with spaces", "My Project", "my-project"},
		{"with underscores", "my_project", "my-project"},
		{"mixed special chars", "My@Project#123", "my-project-123"},
		{"multiple spaces", "my   project", "my-project"},
		{"leading/trailing special", "@project@", "project"},
		{"only special chars", "@#$", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建有效的YAML文件
	validYAML := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validYAML, []byte("test: value"), 0644); err != nil {
		t.Fatalf("Failed to create valid YAML file: %v", err)
	}

	// 创建yml扩展名的文件
	validYML := filepath.Join(tmpDir, "valid.yml")
	if err := os.WriteFile(validYML, []byte("test: value"), 0644); err != nil {
		t.Fatalf("Failed to create valid YML file: %v", err)
	}

	// 创建非YAML文件
	nonYAML := filepath.Join(tmpDir, "invalid.txt")
	if err := os.WriteFile(nonYAML, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create non-YAML file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid yaml file", validYAML, false},
		{"valid yml file", validYML, false},
		{"non-yaml file", nonYAML, true},
		{"non-existent file", filepath.Join(tmpDir, "nonexistent.yaml"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateYAMLFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateYAMLFile(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestFormatPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"current directory", ".", false},
		{"relative path", "./test", false},
		{"absolute path", "/tmp", false},
		{"home path", "~/test", false}, // 注意：filepath.Abs不会展开~
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatPath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if err == nil && result == "" {
				t.Errorf("FormatPath(%q) returned empty path", tt.input)
			}
		})
	}
}
