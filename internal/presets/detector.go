package presets

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type DetectionRule struct {
	Patterns  []string
	Templates []string
}

func DetectFiles(repoPath string) ([]string, error) {
	detected := map[string]struct{}{}
	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := strings.ToLower(d.Name())
		if d.IsDir() {
			if name == ".git" {
				return filepath.SkipDir
			}
			detected[name+"/"] = struct{}{}
			return nil
		}

		detected[name] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan repo: %w", err)
	}

	list := make([]string, 0, len(detected))
	for name := range detected {
		list = append(list, name)
	}
	return list, nil
}

func SuggestTemplates(detected []string) ([]string, error) {
	rules := defaultDetectionRules()
	suggestions := make([]string, 0)
	seen := map[string]struct{}{}

	for _, rule := range rules {
		if ruleMatches(rule, detected) {
			for _, tmpl := range rule.Templates {
				key := strings.ToLower(tmpl)
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				suggestions = append(suggestions, tmpl)
			}
		}
	}

	return suggestions, nil
}

func ruleMatches(rule DetectionRule, detected []string) bool {
	for _, pattern := range rule.Patterns {
		pattern = strings.ToLower(pattern)
		for _, candidate := range detected {
			match, err := filepath.Match(pattern, candidate)
			if err == nil && match {
				return true
			}
			if pattern == candidate {
				return true
			}
		}
	}
	return false
}

func defaultDetectionRules() []DetectionRule {
	return []DetectionRule{
		{Patterns: []string{"package.json"}, Templates: []string{"Node"}},
		{Patterns: []string{"go.mod"}, Templates: []string{"Go"}},
		{Patterns: []string{"requirements.txt", "setup.py", "pyproject.toml"}, Templates: []string{"Python"}},
		{Patterns: []string{"cargo.toml"}, Templates: []string{"Rust"}},
		{Patterns: []string{"pom.xml"}, Templates: []string{"Maven"}},
		{Patterns: []string{"build.gradle", "build.gradle.kts"}, Templates: []string{"Gradle"}},
		{Patterns: []string{"*.csproj"}, Templates: []string{"VisualStudio"}},
		{Patterns: []string{"composer.json"}, Templates: []string{"Composer"}},
		{Patterns: []string{"gemfile"}, Templates: []string{"Ruby"}},
		{Patterns: []string{"*.swift"}, Templates: []string{"Swift"}},
		{Patterns: []string{"*.kt", "*.kts"}, Templates: []string{"Kotlin"}},
		{Patterns: []string{"*.dart"}, Templates: []string{"Dart"}},
		{Patterns: []string{"*.ts", "*.tsx"}, Templates: []string{"TypeScript"}},
		{Patterns: []string{".idea/"}, Templates: []string{"IntelliJ"}},
		{Patterns: []string{".vscode/"}, Templates: []string{"VisualStudioCode"}},
		{Patterns: []string{"*.xcodeproj"}, Templates: []string{"Xcode"}},
		{Patterns: []string{"*.sln"}, Templates: []string{"VisualStudio"}},
	}
}
