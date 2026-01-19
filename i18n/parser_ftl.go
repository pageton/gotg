package i18n

import (
	"bufio"
	"fmt"
	"strings"
)

// FTLParser parses Fluent Translation List (FTL) format
type FTLParser struct{}

// NewFTLParser creates a new FTL parser
func NewFTLParser() *FTLParser {
	return &FTLParser{}
}

// Parse parses FTL content and returns messages
func (p *FTLParser) Parse(content string) (map[string]*Message, error) {
	messages := make(map[string]*Message)
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentKey string
	var currentValue strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for key definition
		if strings.Contains(line, "=") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			// Save previous message
			if currentKey != "" {
				messages[currentKey] = &Message{
					Key:   currentKey,
					Value: strings.TrimSpace(currentValue.String()),
				}
			}

			// Start new message
			parts := strings.SplitN(line, "=", 2)
			currentKey = strings.TrimSpace(parts[0])
			currentValue.Reset()
			currentValue.WriteString(strings.TrimSpace(parts[1]))
		} else if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			// Multiline value continuation
			if currentKey != "" {
				if currentValue.Len() > 0 {
					currentValue.WriteString(" ")
				}
				currentValue.WriteString(strings.TrimSpace(line))
			}
		} else if strings.HasPrefix(line, "  ") {
			// Attribute
			if currentKey != "" {
				attrLine := strings.TrimSpace(line)
				if strings.Contains(attrLine, "=") {
					attrParts := strings.SplitN(attrLine, "=", 2)
					attrName := strings.TrimSpace(attrParts[0])
					attrValue := strings.TrimSpace(attrParts[1])

					msg := messages[currentKey]
					if msg == nil {
						msg = &Message{Key: currentKey}
						messages[currentKey] = msg
					}
					if msg.Attrs == nil {
						msg.Attrs = make(map[string]string)
					}
					msg.Attrs[attrName] = attrValue
				}
			}
		}
	}

	// Save last message
	if currentKey != "" {
		messages[currentKey] = &Message{
			Key:   currentKey,
			Value: strings.TrimSpace(currentValue.String()),
		}
	}

	return messages, nil
}

// ParseVariants parses FTL variants (for pluralization, gender, etc.)
func (p *FTLParser) ParseVariants(content string) (map[string]string, error) {
	variants := make(map[string]string)
	lines := strings.Split(content, "\n")

	var currentVariant string
	var valueBuilder strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			// Save previous variant
			if currentVariant != "" {
				variants[currentVariant] = strings.TrimSpace(valueBuilder.String())
			}

			// Start new variant
			currentVariant = strings.Trim(trimmed, "[]")
			valueBuilder.Reset()
		} else if trimmed != "" && currentVariant != "" {
			if valueBuilder.Len() > 0 {
				valueBuilder.WriteString(" ")
			}
			valueBuilder.WriteString(trimmed)
		}
	}

	// Save last variant
	if currentVariant != "" {
		variants[currentVariant] = strings.TrimSpace(valueBuilder.String())
	}

	return variants, nil
}

// FormatMessage formats an FTL message with variables
func (p *FTLParser) FormatMessage(template string, args map[string]string) string {
	result := template

	// Replace {variable} placeholders
	for key, value := range args {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Replace {$variable} (local variable syntax)
	for key, value := range args {
		placeholder := fmt.Sprintf("{$%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// FTLEncoder generates FTL content from messages
type FTLEncoder struct{}

// NewFTLEncoder creates a new FTL encoder
func NewFTLEncoder() *FTLEncoder {
	return &FTLEncoder{}
}

// Encode converts messages to FTL format
func (e *FTLEncoder) Encode(messages map[string]*Message) string {
	var builder strings.Builder

	for _, msg := range messages {
		builder.WriteString(fmt.Sprintf("%s = %s\n", msg.Key, msg.Value))

		// Write attributes
		for attrName, attrValue := range msg.Attrs {
			builder.WriteString(fmt.Sprintf("  .%s = %s\n", attrName, attrValue))
		}

		// Write variants
		if len(msg.Variants) > 0 {
			for variantName, variantValue := range msg.Variants {
				builder.WriteString(fmt.Sprintf("    [%s]\n", variantName))
				builder.WriteString(fmt.Sprintf("    %s\n", variantValue))
			}
		}

		builder.WriteString("\n")
	}

	return builder.String()
}
