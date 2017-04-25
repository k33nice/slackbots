package channel

import (
	"path/filepath"

	"github.com/nlopes/slack"
)

// Channel - object that aid to interact with slack channels.
type Channel struct {
	// Api - handler for salck client.
	API *slack.Client
}

// New - create new channel in slack.
func (ch *Channel) New(name string) (*slack.Channel, error) {
	return ch.API.CreateChannel(name)
}

// GetChannel - return prepared channel name.
func (ch *Channel) GetChannel(channel string, file string) string {
	slackChan := channel
	if slackChan == "" {
		_, logName := filepath.Split(file)
		ext := filepath.Ext(logName)
		slackChan = logName[0 : len(logName)-len(ext)]
	}

	// NOTE: Hack due to not fixed unification of log files.
	if slackChan == "app" {
		slackChan = "application-log"
	}

	return slackChan
}
