package main

type WebResponseHistory struct {
	Ok             bool        `json:"ok"`
	MediaURLPrefix string      `json:"media_url_prefix"`
	History        []DbHistory `json:"history"`
}

type WebGenImgRequest struct {
	Username string `json:"username"`
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
}

type WebGenImgResponse struct {
	Ok       bool        `json:"ok"`
	Result   SDAPIResult `json:"result"`
	ImageURL string      `json:"image_url"`
}

type WebGradeRequest struct {
	GraderName string `json:"grader_name" bun:"grader_name"`
	ProposalID string `json:"proposal_id" bun:"proposal_id"`
	Grade      int    `json:"grade" bun:"grade"`
}

type WebGradeResponse struct {
	Ok bool `json:"ok"`
}
