package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[38;5;204m" // Soft pink for strings
	ColorGreen  = "\033[38;5;114m" // Soft green for numbers
	ColorBlue   = "\033[38;5;123m" // Cyan-ish blue for property names
	ColorPurple = "\033[38;5;204m" // Soft pink for punctuation
	ColorGray   = "\033[38;5;246m" // Gray for colons and commas
)

// PrettyJSONList formats a JSON array (list) with indentation and colors.
func PrettyJSONList(value interface{}) (string, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	var tempSlice []interface{}
	if err := json.Unmarshal(jsonBytes, &tempSlice); err != nil {
		return "", fmt.Errorf("input is not a valid JSON array: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(ColorPurple + "[\n" + ColorReset)

	for i, item := range tempSlice {
		itemJSON, err := formatJSONObject(item, "  ") // Indent each object
		if err != nil {
			return "", fmt.Errorf("failed to format array item %d: %w", i, err)
		}

		buf.WriteString(itemJSON)

		if i < len(tempSlice)-1 {
			buf.WriteString(ColorGray + ",\n" + ColorReset) // Comma at the END of the line
		} else {
			buf.WriteString("\n")
		}
	}

	buf.WriteString(ColorPurple + "]" + ColorReset)
	return buf.String(), nil
}

// formatJSONObject formats a JSON object with indentation and colors.
func formatJSONObject(value interface{}, indent string) (string, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(indent + ColorPurple + "{\n" + ColorReset)

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	for i, key := range keys {
		val := obj[key]

		// Format key:  Key is on its own line, indented.
		buf.WriteString(indent + "  " + ColorBlue + "\"" + key + "\"" + ColorReset)

		// Colon and space: On the same line as the key.
		buf.WriteString(ColorGray + " : " + ColorReset)

		// Format value:  Value is formatted correctly, using recursion.
		valueStr, err := formatValue(val, indent+"    ") // Increase indent for nested values
		if err != nil {
			return "", fmt.Errorf("failed to format value for key %s: %w", key, err)
		}

		//If the value is an object or array we need to remove the base indentation
		if _, ok := val.(map[string]interface{}); ok {
			valueStr = removeFirstOccurrence(valueStr, indent+"  ")
		}

		if _, ok := val.([]interface{}); ok {
			valueStr = removeFirstOccurrence(valueStr, indent+"  ")
		}
		buf.WriteString(valueStr)

		if i < len(keys)-1 {
			buf.WriteString(ColorGray + ",\n" + ColorReset) // Comma at the END of the line.
		} else {
			buf.WriteString("\n") // No comma after the last value.
		}
	}

	buf.WriteString(indent + ColorPurple + "}" + ColorReset) // Closing brace, correctly indented.
	return buf.String(), nil
}

// formatValue formats a single JSON value.
func formatValue(value interface{}, indent string) (string, error) {
	switch v := value.(type) {
	case string:
		return ColorRed + "\"" + escapeString(v) + "\"" + ColorReset, nil
	case float64:
		return ColorGreen + strconv.FormatFloat(v, 'f', -1, 64) + ColorReset, nil
	case int:
		return ColorGreen + strconv.Itoa(v) + ColorReset, nil
	case bool:
		return ColorGreen + strconv.FormatBool(v) + ColorReset, nil
	case nil:
		return ColorGreen + "null" + ColorReset, nil
	case map[string]interface{}:
		// Recursive call to formatJSONObject, with increased indent.
		return formatJSONObject(v, indent)
	case []interface{}:
		// Recursive call to formatJSONArray, with increased indent.
		return formatJSONArray(v, indent)
	default:
		return fmt.Sprintf("%v", v), nil // Fallback for unknown types.
	}
}

// formatJSONArray formats a JSON array.
func formatJSONArray(arr []interface{}, indent string) (string, error) {
	var buf bytes.Buffer
	buf.WriteString(ColorPurple + "[\n" + ColorReset)

	for i, item := range arr {
		// Indent each element of the array.
		formattedValue, err := formatValue(item, indent+"  ")
		if err != nil {
			return "", err
		}

		//If the value is an object or array we need to remove the base indentation
		if _, ok := item.(map[string]interface{}); ok {
			formattedValue = removeFirstOccurrence(formattedValue, indent)
		}

		if _, ok := item.([]interface{}); ok {
			formattedValue = removeFirstOccurrence(formattedValue, indent)
		}
		buf.WriteString(indent + "  " + formattedValue)

		if i < len(arr)-1 {
			buf.WriteString(ColorGray + ",\n" + ColorReset) // Comma at the END of the line.
		} else {
			buf.WriteString("\n")
		}
	}

	buf.WriteString(indent + ColorPurple + "]" + ColorReset) // Closing bracket, correctly indented.
	return buf.String(), nil
}

// escapeString escapes special characters in a string.
func escapeString(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '\\', '"', '\n', '\r', '\t', '\b', '\f':
			buf.WriteString(fmt.Sprintf("\\%c", escapedChar(r)))
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// escapedChar returns the escaped character.
func escapedChar(r rune) rune {
	switch r {
	case '\\':
		return '\\'
	case '"':
		return '"'
	case '\n':
		return 'n'
	case '\r':
		return 'r'
	case '\t':
		return 't'
	case '\b':
		return 'b'
	case '\f':
		return 'f'
	default:
		return r
	}
}

func removeFirstOccurrence(original, toRemove string) string {
	location := bytes.Index([]byte(original), []byte(toRemove))
	if location == -1 {
		return original
	}
	return original[:location] + original[location+len(toRemove):]
}
