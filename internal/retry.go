package internal

import (
	"time"

	"github.com/slack-go/slack"
)

// withRateLimit retries fn once if Slack responds with a rate-limit error,
// sleeping for the RetryAfter duration specified in the error response.
func withRateLimit(fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}
	if rl, ok := err.(*slack.RateLimitedError); ok {
		time.Sleep(rl.RetryAfter)
		return fn()
	}
	return err
}
