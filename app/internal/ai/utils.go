package ai

import (
	"strings"
)

// extractJSON extracts JSON content from a response that may be wrapped in markdown code blocks
func extractJSON(response string) string {
	jsonStr := response
	
	// Try to extract from ```json blocks
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json") + 7
		end := strings.LastIndex(response, "```")
		if start > 7 && end > start {
			jsonStr = strings.TrimSpace(response[start:end])
		}
	} else if strings.Contains(response, "```") {
		// Try to extract from generic ``` blocks
		start := strings.Index(response, "```") + 3
		end := strings.LastIndex(response, "```")
		if start > 3 && end > start {
			jsonStr = strings.TrimSpace(response[start:end])
		}
	}
	
	return jsonStr
}
