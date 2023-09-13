package main

type WebResponseHistory struct {
	Ok      bool        `json:"ok"`
	History []DbHistory `json:"history"`
}
