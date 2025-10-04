package security

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// URLValidator validates and sanitizes URLs
type URLValidator struct {
	allowedSchemes []string
	allowedDomains []string
	maxURLLength   int
}

// NewURLValidator creates a new URL validator with security defaults
func NewURLValidator() *URLValidator {
	return &URLValidator{
		allowedSchemes: []string{"http", "https"},
		allowedDomains: []string{
			"techcrunch.com",
			"edition.cnn.com",
			"cnn.com",
			"example.com", // For testing
		},
		maxURLLength: 2048, // Reasonable limit
	}
}

// ValidateURL validates and sanitizes a URL
func (v *URLValidator) ValidateURL(rawURL string) (string, error) {
	// Check length
	if len(rawURL) > v.maxURLLength {
		return "", fmt.Errorf("URL too long: %d characters (max: %d)", len(rawURL), v.maxURLLength)
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if !v.isAllowedScheme(parsedURL.Scheme) {
		return "", fmt.Errorf("disallowed URL scheme: %s (allowed: %v)", parsedURL.Scheme, v.allowedSchemes)
	}

	// Check domain
	if !v.isAllowedDomain(parsedURL.Host) {
		return "", fmt.Errorf("disallowed domain: %s (allowed: %v)", parsedURL.Host, v.allowedDomains)
	}

	// Sanitize URL by reconstructing it
	sanitizedURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)
	if parsedURL.RawQuery != "" {
		sanitizedURL += "?" + parsedURL.RawQuery
	}
	if parsedURL.Fragment != "" {
		sanitizedURL += "#" + parsedURL.Fragment
	}

	return sanitizedURL, nil
}

// isAllowedScheme checks if the scheme is allowed
func (v *URLValidator) isAllowedScheme(scheme string) bool {
	for _, allowed := range v.allowedSchemes {
		if scheme == allowed {
			return true
		}
	}
	return false
}

// isAllowedDomain checks if the domain is allowed
func (v *URLValidator) isAllowedDomain(host string) bool {
	// Remove port if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	for _, allowed := range v.allowedDomains {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}

// ValidateQuery validates and sanitizes search queries
func ValidateQuery(query string) (string, error) {
	// Remove potentially dangerous characters
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		return "", fmt.Errorf("query cannot be empty")
	}
	if len(query) > 1000 {
		return "", fmt.Errorf("query too long: %d characters (max: 1000)", len(query))
	}

	// Remove SQL injection patterns
	dangerousPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"union", "select", "insert", "update", "delete", "drop",
		"create", "alter", "exec", "execute",
	}

	lowerQuery := strings.ToLower(query)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return "", fmt.Errorf("query contains potentially dangerous content: %s", pattern)
		}
	}

	return query, nil
}

// SanitizeString removes potentially dangerous characters from strings
func SanitizeString(input string) string {
	// Remove control characters and normalize whitespace
	re := regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`)
	sanitized := re.ReplaceAllString(input, "")

	// Normalize whitespace
	sanitized = strings.TrimSpace(sanitized)
	sanitized = regexp.MustCompile(`\s+`).ReplaceAllString(sanitized, " ")

	return sanitized
}
