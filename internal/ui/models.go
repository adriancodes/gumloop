package ui

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"
)

const modelsAPIURL = "https://models.dev/api.json"

// agentToProvider maps gumloop agent IDs to models.dev provider keys
var agentToProvider = map[string][]string{
	"claude":   {"anthropic"},
	"codex":    {"openai"},
	"gemini":   {"google"},
	"opencode": {"opencode"},
	"ollama":   {"ollama-cloud"},
	"cursor":   {"anthropic", "openai", "google"}, // Combined
}

// modelsAPIResponse represents the structure of the models.dev API
type modelsAPIResponse map[string]providerData

type providerData struct {
	ID     string                `json:"id"`
	Name   string                `json:"name"`
	Models map[string]modelData  `json:"models"`
}

type modelData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// fetchModelsFromAPI fetches models from the models.dev API
// Returns nil if fetch fails (caller should use fallback)
func fetchModelsFromAPI(agentID string) []modelOption {
	providers, ok := agentToProvider[agentID]
	if !ok {
		return nil
	}

	// Fetch with timeout
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(modelsAPIURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var apiResp modelsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil
	}

	// Collect models from all relevant providers
	var models []modelOption
	seen := make(map[string]bool)

	for _, providerKey := range providers {
		provider, ok := apiResp[providerKey]
		if !ok {
			continue
		}

		for _, model := range provider.Models {
			// Filter to "latest" models only
			if !isLatestModel(model.ID, model.Name) {
				continue
			}

			// Skip duplicates (for cursor which combines providers)
			if seen[model.ID] {
				continue
			}
			seen[model.ID] = true

			// Skip embedding models
			if strings.Contains(strings.ToLower(model.ID), "embedding") {
				continue
			}

			models = append(models, modelOption{
				ID:          model.ID,
				Name:        model.Name,
				Desc: "", // API doesn't provide descriptions
			})
		}
	}

	// Sort models alphabetically by name
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	return models
}

// isLatestModel returns true if the model should be shown (is a "latest" version)
func isLatestModel(id, name string) bool {
	idLower := strings.ToLower(id)
	nameLower := strings.ToLower(name)

	// Include models with "latest" in name/id
	if strings.Contains(idLower, "latest") || strings.Contains(nameLower, "latest") {
		return true
	}

	// Include specific model families without version dates
	// These are typically the "current" versions

	// Anthropic: claude-opus-4-5, claude-sonnet-4-5, claude-haiku-4-5 (without dates)
	if strings.HasPrefix(idLower, "claude-") {
		// Exclude dated versions like claude-opus-4-5-20251101
		if !containsDate(idLower) {
			return true
		}
		return false
	}

	// OpenAI: Include main model names without specific dates
	if strings.HasPrefix(idLower, "gpt-") || strings.HasPrefix(idLower, "o1") ||
	   strings.HasPrefix(idLower, "o3") || strings.HasPrefix(idLower, "o4") {
		// Exclude dated versions like gpt-4o-2024-05-13
		if !containsDate(idLower) {
			return true
		}
		return false
	}

	// Google Gemini: Include stable versions without "preview"
	if strings.HasPrefix(idLower, "gemini-") {
		// Include stable releases, exclude previews and dated versions
		if !strings.Contains(idLower, "preview") && !containsDate(idLower) {
			return true
		}
		// Also include gemini-flash-latest and similar
		if strings.Contains(idLower, "latest") {
			return true
		}
		return false
	}

	// Ollama: Include most models (they're already curated)
	// Exclude very large models that most users can't run
	if strings.Contains(idLower, ":") {
		// This is an Ollama model with size specification
		return true
	}

	// OpenCode: Include all (it's a curated list)
	return true
}

// containsDate checks if a string contains a date pattern (YYYY-MM-DD or YYYYMMDD)
func containsDate(s string) bool {
	// Check for patterns like 2024-05-13 or 20241022
	for i := 0; i < len(s)-7; i++ {
		// Check for YYYY-MM-DD pattern
		if i+9 < len(s) && s[i+4] == '-' && s[i+7] == '-' {
			if isDigits(s[i:i+4]) && isDigits(s[i+5:i+7]) && isDigits(s[i+8:i+10]) {
				return true
			}
		}
		// Check for YYYYMMDD pattern
		if isDigits(s[i : i+8]) {
			year := s[i : i+4]
			if year >= "2020" && year <= "2030" {
				return true
			}
		}
	}
	return false
}

// isDigits checks if a string contains only digits
func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// fallbackModels returns hardcoded models when API fetch fails
func fallbackModels(agentID string) []modelOption {
	switch agentID {
	case "claude":
		return []modelOption{
			{ID: "sonnet", Name: "Sonnet", Desc: "Fast and capable (recommended)"},
			{ID: "opus", Name: "Opus", Desc: "Most powerful, best for complex tasks"},
			{ID: "haiku", Name: "Haiku", Desc: "Fastest, good for simple tasks"},
		}
	case "codex":
		return []modelOption{
			{ID: "gpt-4o", Name: "GPT-4o", Desc: "Latest multimodal model (recommended)"},
			{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", Desc: "Fast and powerful"},
			{ID: "o1", Name: "o1", Desc: "Advanced reasoning model"},
			{ID: "o1-mini", Name: "o1-mini", Desc: "Smaller reasoning model"},
		}
	case "gemini":
		return []modelOption{
			{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Desc: "Latest fast model (recommended)"},
			{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", Desc: "Advanced reasoning"},
			{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", Desc: "Fast and efficient"},
		}
	case "ollama":
		return []modelOption{
			{ID: "deepseek-coder", Name: "DeepSeek Coder", Desc: "Specialized for code"},
			{ID: "codellama", Name: "Code Llama", Desc: "Meta's code model"},
			{ID: "qwen2.5-coder", Name: "Qwen 2.5 Coder", Desc: "Alibaba's code model"},
			{ID: "llama3.1", Name: "Llama 3.1", Desc: "General purpose"},
		}
	case "cursor":
		return []modelOption{
			{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", Desc: "Anthropic"},
			{ID: "gpt-4o", Name: "GPT-4o", Desc: "OpenAI"},
			{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Desc: "Google"},
		}
	case "opencode":
		return []modelOption{
			{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", Desc: "Recommended"},
			{ID: "gpt-5-codex", Name: "GPT-5 Codex", Desc: "OpenAI coding model"},
		}
	default:
		return []modelOption{}
	}
}
