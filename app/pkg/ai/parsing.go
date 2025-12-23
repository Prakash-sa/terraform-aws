package ai

import (
	"encoding/json"
	"regexp"
	"strings"
)

// parseAnalysisResponse parses the AI response and extracts analysis data
func parseAnalysisResponse(rawResp string) (*AnalysisResponse, error) {
	jsonStr := extractJSON(rawResp)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If JSON parsing fails, return a basic response
		return &AnalysisResponse{
			Summary:            rawResp,
			Findings:           []string{},
			RootCauses:         []string{},
			RecommendedActions: []string{},
			SuggestedSeverity:  "unknown",
			RawResponse:        rawResp,
		}, nil
	}

	return &AnalysisResponse{
		Summary:            getStringValue(data, "summary"),
		Findings:           getStringSlice(data, "findings"),
		RootCauses:         getStringSlice(data, "root_causes"),
		RecommendedActions: getStringSlice(data, "recommended_actions"),
		SuggestedSeverity:  getStringValue(data, "suggested_severity"),
		RawResponse:        rawResp,
	}, nil
}

// parseRCAResponse parses the AI response for RCA generation
func parseRCAResponse(rawResp string) (*RCAResponse, error) {
	jsonStr := extractJSON(rawResp)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If JSON parsing fails, return a basic response
		return &RCAResponse{
			Timeline:            rawResp,
			RootCause:           "",
			Impact:              "",
			ImmediateResolution: "",
			PreventiveMeasures:  []string{},
			LessonsLearned:      []string{},
			RawResponse:         rawResp,
		}, nil
	}

	return &RCAResponse{
		Timeline:            getStringValue(data, "timeline"),
		RootCause:           getStringValue(data, "root_cause"),
		Impact:              getStringValue(data, "impact"),
		ImmediateResolution: getStringValue(data, "immediate_resolution"),
		PreventiveMeasures:  getStringSlice(data, "preventive_measures"),
		LessonsLearned:      getStringSlice(data, "lessons_learned"),
		RawResponse:         rawResp,
	}, nil
}

// parseSummarizeResponse parses the AI response for log summarization
func parseSummarizeResponse(rawResp string) (*SummarizeResponse, error) {
	jsonStr := extractJSON(rawResp)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If JSON parsing fails, return a basic response
		return &SummarizeResponse{
			Summary:     rawResp,
			KeyInsights: []string{},
			Alerts:      []string{},
			RawResponse: rawResp,
		}, nil
	}

	return &SummarizeResponse{
		Summary:     getStringValue(data, "summary"),
		KeyInsights: getStringSlice(data, "key_insights"),
		Alerts:      getStringSlice(data, "alerts"),
		RawResponse: rawResp,
	}, nil
}

// extractJSON extracts JSON object from a string that may be wrapped in markdown
func extractJSON(s string) string {
	// Remove markdown code blocks if present
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}

	// Find the first { and last } to extract JSON
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")

	if start != -1 && end != -1 && start < end {
		return s[start : end+1]
	}

	return s
}

// getStringValue safely extracts a string value from a map
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getStringSlice safely extracts a string slice from a map
func getStringSlice(data map[string]interface{}, key string) []string {
	if val, ok := data[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

// TrimLongText truncates text to a maximum length
func TrimLongText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

// SanitizePrompt removes potentially harmful content from prompts
func SanitizePrompt(prompt string) string {
	// Basic sanitization - add more rules as needed
	re := regexp.MustCompile(`(?i)(api[_-]?key|password|secret|token).*`)
	return re.ReplaceAllString(prompt, "$1=***REDACTED***")
}
