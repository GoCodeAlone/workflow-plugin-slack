package internal

import (
	"context"
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// slackTrigger fires workflow callbacks from Socket Mode events.
type slackTrigger struct {
	providerName string
	callback     sdk.TriggerCallback
	cancel       context.CancelFunc
}

func newSlackTrigger(config map[string]any, cb sdk.TriggerCallback) (*slackTrigger, error) {
	providerName := stringFrom(config, "provider")
	if providerName == "" {
		return nil, fmt.Errorf("trigger.slack: 'provider' config field is required")
	}
	return &slackTrigger{providerName: providerName, callback: cb}, nil
}

// Start begins consuming Socket Mode events from the registered provider.
func (t *slackTrigger) Start(ctx context.Context) error {
	p, err := getProvider(t.providerName)
	if err != nil {
		return fmt.Errorf("trigger.slack: %w", err)
	}
	ctx, t.cancel = context.WithCancel(ctx)
	go t.run(ctx, p.socketClient)
	return nil
}

// Stop cancels the event loop.
func (t *slackTrigger) Stop(_ context.Context) error {
	if t.cancel != nil {
		t.cancel()
	}
	return nil
}

func (t *slackTrigger) run(ctx context.Context, sc *socketmode.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-sc.Events:
			if !ok {
				return
			}
			t.handleEvent(evt)
		}
	}
}

func (t *slackTrigger) handleEvent(evt socketmode.Event) {
	switch evt.Type {
	case socketmode.EventTypeEventsAPI:
		inner, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return
		}
		_ = t.callback("message", map[string]any{
			"type":       string(inner.Type),
			"team_id":    inner.TeamID,
			"api_app_id": inner.APIAppID,
			"event":      inner.InnerEvent.Data,
		})

	case socketmode.EventTypeSlashCommand:
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			return
		}
		_ = t.callback("slash_command", map[string]any{
			"command":    cmd.Command,
			"text":       cmd.Text,
			"user_id":    cmd.UserID,
			"channel_id": cmd.ChannelID,
			"team_id":    cmd.TeamID,
		})

	case socketmode.EventTypeInteractive:
		cb, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			return
		}
		_ = t.callback("interaction", map[string]any{
			"type":        string(cb.Type),
			"callback_id": cb.CallbackID,
			"action_id":   cb.ActionID,
			"user_id":     cb.User.ID,
			"channel_id":  cb.Channel.ID,
			"team_id":     cb.Team.ID,
		})
	}
}
