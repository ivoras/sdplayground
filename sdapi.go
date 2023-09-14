package main

type SDAPIResult struct {
	Status         string         `json:"status"`
	GenerationTime float64        `json:"generationTime"`
	ID             int            `json:"id"`
	Output         []string       `json:"output"`
	WebhookStatus  string         `json:"webhook_status"`
	Meta           map[string]any `json:"meta"`
}

type SDAPIRequest struct {
	Key               string  `json:"key"`
	ModelID           string  `json:"model_id"`
	Prompt            string  `json:"prompt"`
	NegativePrompt    string  `json:"negative_prompt,omitempty"`
	Width             int     `json:"width,omitempty"`
	Height            int     `json:"height,omitempty"`
	Samples           int     `json:"samples,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
	GuidanceScale     float64 `json:"guidance_scale,omitempty"`
}
