package main

import (
	"time"

	"github.com/uptrace/bun"
)

type DbHistory struct {
	bun.BaseModel `bun:"table:history"`
	ID            int         `bun:"id,pk,autoincrement" json:"id"`
	Timestamp     time.Time   `bun:"ts" json:"ts"`
	Username      string      `bun:"username" json:"username"`
	Model         string      `bun:"model" json:"model"`
	Prompt        string      `bun:"prompt" json:"prompt"`
	ImageFilename string      `bun:"image_filename" json:"image_filename"`
	Result        SDAPIResult `bun:"result" json:"result"`
}

type DbGrades struct {
	bun.BaseModel `bun:"table:grades"`
	WebGradeRequest
}
