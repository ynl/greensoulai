package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// ValidateProjectName 验证项目名称
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// 检查长度
	if len(name) < 2 {
		return fmt.Errorf("project name must be at least 2 characters long")
	}

	if len(name) > 50 {
		return fmt.Errorf("project name must be less than 50 characters long")
	}

	// 检查字符
	validName := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("project name must start with a letter and contain only letters, numbers, hyphens, and underscores")
	}

	// 检查保留字
	reservedNames := []string{
		"main", "test", "src", "internal", "cmd", "pkg", "vendor",
		"go", "golang", "crewai", "greensoulai",
	}

	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return fmt.Errorf("project name '%s' is reserved", name)
		}
	}

	return nil
}

// ValidateGoModule 验证Go模块名
func ValidateGoModule(module string) error {
	if module == "" {
		return fmt.Errorf("go module name cannot be empty")
	}

	// 基本格式检查
	validModule := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-/]*[a-zA-Z0-9]$`)
	if !validModule.MatchString(module) {
		return fmt.Errorf("invalid go module name format")
	}

	// 检查是否包含域名（推荐格式）
	parts := strings.Split(module, "/")
	if len(parts) >= 2 {
		// 第一部分应该像域名
		domain := parts[0]
		if strings.Contains(domain, ".") {
			return nil // 看起来像是有效的域名格式
		}
	}

	// 如果不是域名格式，给出警告信息
	return fmt.Errorf("go module name should preferably be in domain format (e.g., github.com/username/project)")
}

// ValidateDirectoryName 验证目录名称
func ValidateDirectoryName(name string) error {
	if name == "" {
		return fmt.Errorf("directory name cannot be empty")
	}

	// 检查是否包含无效字符
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("directory name cannot contain '%s'", char)
		}
	}

	// 检查是否全是点
	if strings.Trim(name, ".") == "" {
		return fmt.Errorf("directory name cannot be only dots")
	}

	return nil
}

// SanitizeName 清理名称，使其符合标识符规范
func SanitizeName(name string) string {
	// 移除开头和结尾的空格
	name = strings.TrimSpace(name)

	// 如果全是特殊字符，返回默认名称
	hasValidChar := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasValidChar = true
			break
		}
	}
	if !hasValidChar {
		return "project"
	}

	// 移除所有前导数字
	removedLeadingDigit := false
	for len(name) > 0 && unicode.IsDigit(rune(name[0])) {
		name = name[1:]
		removedLeadingDigit = true
	}

	// 替换空格和特殊字符为下划线
	var result strings.Builder
	var lastWasUnderscore bool

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			lastWasUnderscore = false
		} else if !lastWasUnderscore {
			// 任何非字母数字字符都替换为下划线
			result.WriteRune('_')
			lastWasUnderscore = true
		}
	}

	// 处理首字符不是字母的情况
	sanitized := result.String()
	if len(sanitized) > 0 {
		firstRune := rune(sanitized[0])
		if !unicode.IsLetter(firstRune) {
			// 如果首字符不是字母，添加A前缀
			sanitized = "A" + sanitized
		}
	}

	// 如果移除了前导数字且有有效内容，添加A前缀
	if removedLeadingDigit && len(sanitized) > 0 {
		sanitized = "A" + sanitized
	}

	// 移除结尾的下划线
	sanitized = strings.TrimRight(sanitized, "_")

	// 如果结果为空，使用默认名称
	if sanitized == "" {
		sanitized = "project"
	}

	return sanitized
}

// CheckDirectoryExists 检查目录是否存在
func CheckDirectoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check directory: %w", err)
	}

	if !info.IsDir() {
		return false, fmt.Errorf("path exists but is not a directory")
	}

	return true, nil
}

// IsDirectoryEmpty 检查目录是否为空
func IsDirectoryEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to read directory: %w", err)
	}

	// 忽略隐藏文件
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			return false, nil
		}
	}

	return true, nil
}

// EnsureDirectoryExists 确保目录存在
func EnsureDirectoryExists(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// ValidateAPIKey 验证API密钥格式
func ValidateAPIKey(key string) error {
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// OpenAI API key通常以sk-开头
	if strings.HasPrefix(key, "sk-") {
		if len(key) < 20 {
			return fmt.Errorf("API key appears to be too short")
		}
		return nil
	}

	// OpenRouter API key通常以sk-or-开头
	if strings.HasPrefix(key, "sk-or-") {
		if len(key) < 30 {
			return fmt.Errorf("API key appears to be too short")
		}
		return nil
	}

	// 其他格式的API key
	if len(key) < 10 {
		return fmt.Errorf("API key appears to be too short")
	}

	return nil
}

// GenerateGoModule 生成Go模块名建议
func GenerateGoModule(projectName string) string {
	// 对于Go模块名，保持连字符但替换其他特殊字符
	sanitized := sanitizeForGoModule(projectName)
	return fmt.Sprintf("github.com/username/%s", strings.ToLower(sanitized))
}

// sanitizeForGoModule 为Go模块名清理字符串（保留连字符）
func sanitizeForGoModule(name string) string {
	name = strings.TrimSpace(name)

	var result strings.Builder
	var lastWasDash bool

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			lastWasDash = false
		} else if r == '-' && !lastWasDash {
			result.WriteRune('-') // 连字符保持不变
			lastWasDash = true
		} else if r == '_' && !lastWasDash {
			result.WriteRune('_') // 下划线保持不变
			lastWasDash = true
		} else if !lastWasDash && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			// 其他特殊字符替换为下划线
			result.WriteRune('_')
			lastWasDash = true
		}
	}

	// 移除开头和结尾的连字符和下划线
	sanitized := strings.Trim(result.String(), "-_")

	// 如果结果为空，使用默认名称
	if sanitized == "" {
		return "project"
	}

	return sanitized
}

// FormatPath 格式化路径
func FormatPath(path string) (string, error) {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 清理路径
	cleanPath := filepath.Clean(absPath)

	return cleanPath, nil
}

// ValidateYAMLFile 检查YAML文件是否存在且有效
func ValidateYAMLFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("YAML file does not exist: %s", path)
		}
		return fmt.Errorf("failed to access YAML file: %w", err)
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("file does not have YAML extension: %s", path)
	}

	return nil
}

// NormalizeName 标准化名称（用于文件名、目录名等）
func NormalizeName(name string) string {
	// 转换为小写
	name = strings.ToLower(name)

	// 替换空格和特殊字符为连字符
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "-")

	// 移除开头和结尾的连字符
	name = strings.Trim(name, "-")

	return name
}

// ToPascalCase 转换为帕斯卡命名法
func ToPascalCase(s string) string {
	if s == "" {
		return s
	}

	// 检查是否已经是帕斯卡命名法
	if isPascalCase(s) {
		return s
	}

	// 分割字符串
	parts := regexp.MustCompile(`[^a-zA-Z0-9]+`).Split(s, -1)

	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}

// isPascalCase 检查字符串是否已经是帕斯卡命名法
func isPascalCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	// 必须以大写字母开头
	if !unicode.IsUpper(rune(s[0])) {
		return false
	}

	// 只能包含字母和数字，且无分隔符
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	// 检查是否符合帕斯卡命名法的模式：大写字母后跟小写字母或数字
	inWord := false
	for i, r := range s {
		if i == 0 && unicode.IsUpper(r) {
			inWord = true
			continue
		}

		if unicode.IsUpper(r) {
			// 新单词开始
			inWord = true
		} else if inWord && (unicode.IsLower(r) || unicode.IsDigit(r)) {
			// 在单词中
			continue
		} else {
			// 不符合模式
			return false
		}
	}

	return true
}

// ToCamelCase 转换为驼峰命名法
func ToCamelCase(s string) string {
	if s == "" {
		return s
	}

	// 检查是否已经是驼峰命名法
	if isCamelCase(s) {
		return s
	}

	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}

	return strings.ToLower(pascal[:1]) + pascal[1:]
}

// isCamelCase 检查字符串是否已经是驼峰命名法
func isCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	// 必须以小写字母开头
	if !unicode.IsLower(rune(s[0])) {
		return false
	}

	// 只能包含字母和数字，且单词间无分隔符
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

// ToSnakeCase 转换为蛇形命名法
func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	// 处理连续大写字母的情况，如XMLHttpRequest -> XML_Http_Request
	snake := regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`).ReplaceAllString(s, "${1}_${2}")

	// 在小写字母和大写字母间插入下划线
	snake = regexp.MustCompile(`([a-z0-9])([A-Z])`).ReplaceAllString(snake, "${1}_${2}")

	// 替换空格和特殊字符为下划线
	snake = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(snake, "_")

	// 转换为小写
	snake = strings.ToLower(snake)

	// 移除重复的下划线
	snake = regexp.MustCompile(`_+`).ReplaceAllString(snake, "_")

	// 移除开头和结尾的下划线
	snake = strings.Trim(snake, "_")

	return snake
}
