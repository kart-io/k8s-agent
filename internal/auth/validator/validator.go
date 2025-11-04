package validator

import (
	"fmt"
	"regexp"
	"unicode"
)

var (
	// Email regex pattern.
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	// Username regex: alphanumeric, underscore, hyphen.
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	// Phone regex: basic international format.
	phoneRegex = regexp.MustCompile(`^[+]?[0-9]{10,15}$`)
)

// ValidateUsername validates username format and length.
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if len(username) > 50 {
		return fmt.Errorf("username must not exceed 50 characters")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores and hyphens")
	}
	return nil
}

// ValidatePassword validates password strength.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateEmail validates email format.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if len(email) > 255 {
		return fmt.Errorf("email must not exceed 255 characters")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidatePhone validates phone number format (optional).
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone number format")
	}
	return nil
}

// ValidateRequired checks if a field is not empty.
func ValidateRequired(field, fieldName string) error {
	if field == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateLength validates string length.
func ValidateLength(value, fieldName string, min, max int) error {
	length := len(value)
	if min > 0 && length < min {
		return fmt.Errorf("%s must be at least %d characters", fieldName, min)
	}
	if max > 0 && length > max {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, max)
	}
	return nil
}

// ValidatePermissionType validates permission type.
func ValidatePermissionType(permType string) error {
	validTypes := map[string]bool{
		"menu":   true,
		"button": true,
		"api":    true,
	}
	if !validTypes[permType] {
		return fmt.Errorf("invalid permission type: must be one of [menu, button, api]")
	}
	return nil
}

// ValidateStatus validates status value.
func ValidateStatus(status int) error {
	if status != 0 && status != 1 {
		return fmt.Errorf("invalid status: must be 0 (disabled) or 1 (active)")
	}
	return nil
}

// ValidateUUID validates UUID format (basic check).
func ValidateUUID(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("invalid UUID format")
	}
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	if !uuidRegex.MatchString(id) {
		return fmt.Errorf("invalid UUID format")
	}
	return nil
}
