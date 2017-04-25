package main

import "github.com/nlopes/slack"

// Message - message instance for slack.
type Message struct {
	name       string
	channel    string
	attachment slack.Attachment
}
