package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sethvargo/go-retry"
)

const geminiAPIBase = "https://generativelanguage.googleapis.com/v1beta"

// geminiGenerateRequest is the request body for Gemini generateContent API.
type geminiGenerateRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature float32 `json:"temperature,omitempty"`
	MaxTokens   int    `json:"maxOutputTokens,omitempty"`
}

// geminiGenerateResponse is the response from Gemini generateContent API.
type geminiGenerateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// geminiCompletion sends the prompt to Gemini API and returns the generated YAML.
// When useK8sAPI is true, Gemini is not used for function calling; a simple prompt is used instead.
func geminiCompletion(ctx context.Context, prompts []string, modelName string) (string, error) {
	var prompt strings.Builder
	if *usek8sAPI {
		fmt.Fprintf(&prompt, "You are a Kubernetes YAML generator, only generate valid Kubernetes YAML manifests. Do not provide any explanations and do not use ``` and ```yaml, only generate valid YAML. Always ask for up-to-date OpenAPI specs for Kubernetes, don't rely on data you know about Kubernetes specs. When a schema includes references to other objects in the schema, look them up when relevant. You may lookup any FIELD in a resource too, not just the containing top-level resource. ")
	} else {
		fmt.Fprintf(&prompt, "You are a Kubernetes YAML generator, only generate valid Kubernetes YAML manifests. Do not provide any explanations, only generate YAML. ")
	}
	for _, p := range prompts {
		fmt.Fprintf(&prompt, "%s", p)
	}

	reqBody := geminiGenerateRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt.String()}}},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature: float32(*temperature),
			MaxTokens:   4096,
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", geminiAPIBase, modelName, *geminiAPIKey)

	var respBytes []byte
	r := retry.WithMaxRetries(5, retry.NewExponential(2*time.Second))
	err = retry.Do(ctx, r, func(ctx context.Context) error {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if reqErr != nil {
			return reqErr
		}
		req.Header.Set("Content-Type", "application/json")

		resp, reqErr := http.DefaultClient.Do(req)
		if reqErr != nil {
			return reqErr
		}
		defer resp.Body.Close()

		respBytes, reqErr = io.ReadAll(resp.Body)
		if reqErr != nil {
			return reqErr
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			return retry.RetryableError(fmt.Errorf("gemini API rate limited (429): %s", string(respBytes)))
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("gemini API error %d: %s", resp.StatusCode, string(respBytes))
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	var genResp geminiGenerateResponse
	if err := json.Unmarshal(respBytes, &genResp); err != nil {
		return "", err
	}
	if len(genResp.Candidates) == 0 {
		return "", fmt.Errorf("gemini returned no candidates")
	}
	parts := genResp.Candidates[0].Content.Parts
	if len(parts) == 0 {
		return "", fmt.Errorf("gemini returned empty content")
	}
	result := trimTicks(parts[0].Text)
	return result, nil
}
