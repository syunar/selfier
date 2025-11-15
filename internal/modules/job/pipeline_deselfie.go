package job

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"selfier/pkg/config"
	"selfier/pkg/middleware"
)

// deselfiePipelineAPI handles communication with the Deselfie pipeline service.
type deselfiePipelineAPI struct {
	endpoint    string
	tokenID     string
	tokenSecret string
	httpClient  *http.Client
}

// NewDeselfiePipelineAPI creates a new client for the Deselfie pipeline.
// It reuses a single http.Client for efficiency.
func NewDeselfiePipelineAPI(cfg *config.ModalConfig) DeselfiePipeline {
	return &deselfiePipelineAPI{
		endpoint:    cfg.EndpointDeselfie,
		tokenID:     cfg.TokenID,
		tokenSecret: cfg.TokenSecret,
		httpClient:  &http.Client{
			// A 5-minute timeout is a reasonable default for long-running generation jobs.
			// Timeout: time.Second * 300,
		},
	}
}

// apiRequest defines the structure of the JSON payload sent to the API.
type apiRequest struct {
	ImageURL string `json:"image_url"`
	Prompt   string `json:"prompt"`
}

// apiResponse defines the structure of the expected JSON response from the API.
type apiResponse struct {
	ImageURL string `json:"image_url"`
}

// Generate sends an image URL and prompt to the Deselfie pipeline,
// downloads the resulting image, and returns it as a base64 encoded string.
func (p *deselfiePipelineAPI) Generate(ctx context.Context, imageURL string, options map[string]any) (string, error) {
	log := middleware.GetLogger(ctx)
	log.Info("starting deselfie pipeline")

	// 1. Safely extract prompt from options.
	prompt, ok := options["prompt"].(string)
	if !ok || prompt == "" {
		return "", fmt.Errorf("prompt option is missing, empty, or not a string")
	}

	// 2. Marshal the request body using a struct for type safety.
	payload := apiRequest{
		ImageURL: imageURL,
		Prompt:   prompt,
	}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Error("failed to marshal request body", "error", err)
		return "", fmt.Errorf("could not create request payload: %w", err)
	}

	// 3. Create and execute the POST request, respecting the context.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("failed to create API request", "error", err)
		return "", fmt.Errorf("could not create API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Modal-Key", p.tokenID)
	req.Header.Set("Modal-Secret", p.tokenSecret)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		log.Error("failed to execute API request", "error", err)
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// 4. Check for non-successful status codes.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Read body for debugging, ignore error here.
		log.Error("API returned non-200 status", "status", resp.Status, "body", string(body))
		return "", fmt.Errorf("API returned an error: %s", resp.Status)
	}

	// 5. Decode the successful response.
	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("failed to unmarshal API response", "error", err)
		return "", fmt.Errorf("could not parse API response: %w", err)
	}
	if result.ImageURL == "" {
		log.Error("API response is missing image_url")
		return "", fmt.Errorf("API response did not contain an image URL")
	}

	log.Info("successfully received output image url", "url", result.ImageURL)

	// 6. Download the generated image.
	imageBase64, err := p.downloadAndEncodeImage(ctx, result.ImageURL)
	if err != nil {
		// The helper function already logs, so we just wrap the error.
		return "", fmt.Errorf("failed to process generated image: %w", err)
	}

	return imageBase64, nil
}

// downloadAndEncodeImage fetches an image from a URL and returns it as a base64 string.
func (p *deselfiePipelineAPI) downloadAndEncodeImage(ctx context.Context, url string) (string, error) {
	log := middleware.GetLogger(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error("failed to create image download request", "url", url, "error", err)
		return "", fmt.Errorf("could not create image download request: %w", err)
	}

	imageResp, err := p.httpClient.Do(req)
	if err != nil {
		log.Error("failed to download image", "url", url, "error", err)
		return "", fmt.Errorf("could not download image: %w", err)
	}
	defer imageResp.Body.Close()

	if imageResp.StatusCode != http.StatusOK {
		log.Error("image host returned non-200 status", "url", url, "status", imageResp.Status)
		return "", fmt.Errorf("image host returned an error: %s", imageResp.Status)
	}

	imageBytes, err := io.ReadAll(imageResp.Body)
	if err != nil {
		log.Error("failed to read image bytes", "url", url, "error", err)
		return "", fmt.Errorf("could not read image data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(imageBytes), nil
}
