package utils

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// PromptString prompts for a string with validation
func PromptString(label string, validator func(string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validator,
	}

	return prompt.Run()
}

// PromptSelect prompts for selection from a list
func PromptSelect(label string, items []string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}

	_, result, err := prompt.Run()
	return result, err
}

// PromptYesNo prompts for a yes/no confirmation
func PromptYesNo(label string) (bool, error) {
	prompt := promptui.Select{
		Label: label,
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return false, err
	}

	return result == "Yes", nil
}

// ValidateProjectName validates a project name
func ValidateProjectName(input string) error {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return fmt.Errorf("project name cannot be empty")
	}
	if len(input) < 3 {
		return fmt.Errorf("project name must be at least 3 characters")
	}
	if len(input) > 50 {
		return fmt.Errorf("project name must be no more than 50 characters")
	}
	// Check for valid characters (alphanumeric, hyphens, underscores)
	for _, r := range input {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("project name can only contain letters, numbers, hyphens, and underscores")
		}
	}
	return nil
}
