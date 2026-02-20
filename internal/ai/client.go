package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// Client is an HTTP client for the AI Engine
type Client struct {
	baseURL    string
	httpClient *http.Client
	enabled    bool
}

// NewClient creates a new AI Engine client
func NewClient(baseURL string, timeout int, enabled bool) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		enabled: enabled,
	}
}

// IsEnabled returns whether AI Engine is enabled
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// HealthCheck checks if AI Engine is healthy
func (c *Client) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	if !c.enabled {
		return &HealthResponse{Status: "disabled"}, nil
	}

	resp, err := c.doRequest(ctx, "GET", "/api/v1/ai/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse[HealthResponse]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if !apiResp.Success {
		errMsg := "Unknown error"
		if apiResp.Error != nil {
			errMsg = *apiResp.Error
		}
		return nil, fmt.Errorf("AI Engine Error: %s", errMsg)
	}

	return &apiResp.Data, nil
}

// DetectAnomaly detects anomalies in sensor readings
func (c *Client) DetectAnomaly(ctx context.Context, req *AnomalyRequest) (*AnomalyResponse, error) {
	if !c.enabled {
		return &AnomalyResponse{IsAnomaly: false}, nil
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/ai/anomaly-detect", req)
	if err != nil {
		logger.Warn("AI anomaly detection failed: %v", err)
		return &AnomalyResponse{IsAnomaly: false}, nil // Graceful degradation
	}
	defer resp.Body.Close()

	var apiResp APIResponse[AnomalyResponse]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if !apiResp.Success {
		errMsg := "Unknown error"
		if apiResp.Error != nil {
			errMsg = *apiResp.Error
		}
		logger.Warn("AI Engine anomaly detection error: %s", errMsg)
		return &AnomalyResponse{IsAnomaly: false}, nil // Graceful degradation
	}
	return &apiResp.Data, nil
}

// PredictGrade predicts harvest grade based on environment
func (c *Client) PredictGrade(ctx context.Context, req *GradePredictionRequest) (*GradePredictionResponse, error) {
	if !c.enabled {
		// Fallback to threshold-based grading
		return c.fallbackGradePredict(req), nil
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/ai/predict-grade", req)
	if err != nil {
		logger.Warn("AI grade prediction failed: %v", err)
		return c.fallbackGradePredict(req), nil // Graceful degradation
	}
	defer resp.Body.Close()

	var apiResp APIResponse[GradePredictionResponse]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if !apiResp.Success {
		errMsg := "Unknown error"
		if apiResp.Error != nil {
			errMsg = *apiResp.Error
		}
		logger.Warn("AI Engine grade prediction error: %s", errMsg)
		return c.fallbackGradePredict(req), nil // Graceful degradation
	}
	return &apiResp.Data, nil
}

// PredictPump recommends pump action based on environment
func (c *Client) PredictPump(ctx context.Context, req *PumpPredictionRequest) (*PumpPredictionResponse, error) {
	if !c.enabled {
		return &PumpPredictionResponse{PumpState: "OFF", Confidence: 0}, nil
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/ai/predict-pump", req)
	if err != nil {
		logger.Warn("AI pump prediction failed: %v", err)
		return &PumpPredictionResponse{PumpState: "OFF", Confidence: 0}, nil
	}
	defer resp.Body.Close()

	var apiResp APIResponse[PumpPredictionResponse]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if !apiResp.Success {
		errMsg := "Unknown error"
		if apiResp.Error != nil {
			errMsg = *apiResp.Error
		}
		logger.Warn("AI Engine pump prediction error: %s", errMsg)
		return &PumpPredictionResponse{PumpState: "OFF", Confidence: 0}, nil
	}
	return &apiResp.Data, nil
}

// Analyze performs comprehensive analysis
func (c *Client) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("AI Engine is disabled")
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/ai/analyze", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse[AnalyzeResponse]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if !apiResp.Success {
		errMsg := "Unknown error"
		if apiResp.Error != nil {
			errMsg = *apiResp.Error
		}
		return nil, fmt.Errorf("AI Engine Error: %s", errMsg)
	}
	return &apiResp.Data, nil
}

// doRequest performs an HTTP request to the AI Engine
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("AI Engine returned status %d", resp.StatusCode)
	}

	return resp, nil
}

// fallbackGradePredict provides threshold-based grade prediction
func (c *Client) fallbackGradePredict(req *GradePredictionRequest) *GradePredictionResponse {
	// Simple threshold-based grading
	score := 100.0

	// Temperature scoring (optimal: 27-30°C)
	if req.Temperature < 20 || req.Temperature > 35 {
		score -= 40
	} else if req.Temperature < 25 || req.Temperature > 32 {
		score -= 20
	}

	// Humidity scoring (optimal: 70-80%)
	if req.Humidity < 50 || req.Humidity > 95 {
		score -= 30
	} else if req.Humidity < 65 || req.Humidity > 85 {
		score -= 15
	}

	// Ammonia scoring (lower is better)
	if req.Ammonia > 25 {
		score -= 40
	} else if req.Ammonia > 15 {
		score -= 20
	}

	var grade string
	if score >= 70 {
		grade = "bagus"
	} else if score >= 40 {
		grade = "sedang"
	} else {
		grade = "buruk"
	}

	return &GradePredictionResponse{
		Grade:      grade,
		Confidence: 0.5, // Low confidence for fallback
		Probabilities: map[string]float64{
			"bagus":  0.33,
			"sedang": 0.34,
			"buruk":  0.33,
		},
	}
}
