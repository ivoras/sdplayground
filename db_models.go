package main

import (
	"time"

	"github.com/uptrace/bun"
)

type DbHistory struct {
	bun.BaseModel `bun:"table:history"`
	ID            int         `bun:"id,pk,autoincrement"`
	Timestamp     time.Time   `bun:"ts"`
	Username      string      `bun:"username"`
	Prompt        string      `bun:"prompt"`
	Result        SDAPIResult `bun:"result"`
}
