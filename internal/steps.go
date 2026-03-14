package internal

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"github.com/slack-go/slack"
)

// --- step.slack_send_message ---

type sendMessageStep struct{ providerName string }

func (s *sendMessageStep) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current, _, _ map[string]any) (*sdk.StepResult, error) {
	p, err := getProvider(s.providerName)
	if err != nil {
		return nil, err
	}
	channelID := stringFrom(current, "channel_id")
	content := stringFrom(current, "content")
	if channelID == "" || content == "" {
		return nil, fmt.Errorf("slack_send_message: channel_id and content are required")
	}
	var ch, ts string
	if err := withRateLimit(func() error {
		ch, ts, _, err = p.client.SendMessage(channelID, slack.MsgOptionText(content, false))
		return err
	}); err != nil {
		return nil, fmt.Errorf("slack_send_message: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"channel": ch, "timestamp": ts}}, nil
}

// --- step.slack_send_blocks ---

type sendBlocksStep struct{ providerName string }

func (s *sendBlocksStep) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current, _, _ map[string]any) (*sdk.StepResult, error) {
	p, err := getProvider(s.providerName)
	if err != nil {
		return nil, err
	}
	channelID := stringFrom(current, "channel_id")
	blocksRaw, _ := current["blocks"]
	if channelID == "" || blocksRaw == nil {
		return nil, fmt.Errorf("slack_send_blocks: channel_id and blocks are required")
	}

	// blocks can be a JSON string or already a slice.
	var blocks []slack.Block
	switch v := blocksRaw.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &blocks); err != nil {
			return nil, fmt.Errorf("slack_send_blocks: invalid blocks JSON: %w", err)
		}
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("slack_send_blocks: marshal blocks: %w", err)
		}
		if err := json.Unmarshal(raw, &blocks); err != nil {
			return nil, fmt.Errorf("slack_send_blocks: unmarshal blocks: %w", err)
		}
	}

	var ch, ts string
	if err := withRateLimit(func() error {
		ch, ts, _, err = p.client.SendMessage(channelID, slack.MsgOptionBlocks(blocks...))
		return err
	}); err != nil {
		return nil, fmt.Errorf("slack_send_blocks: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"channel": ch, "timestamp": ts}}, nil
}

// --- step.slack_edit_message ---

type editMessageStep struct{ providerName string }

func (s *editMessageStep) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current, _, _ map[string]any) (*sdk.StepResult, error) {
	p, err := getProvider(s.providerName)
	if err != nil {
		return nil, err
	}
	channelID := stringFrom(current, "channel_id")
	timestamp := stringFrom(current, "timestamp")
	content := stringFrom(current, "content")
	if channelID == "" || timestamp == "" || content == "" {
		return nil, fmt.Errorf("slack_edit_message: channel_id, timestamp, and content are required")
	}
	var ch, ts string
	if err := withRateLimit(func() error {
		ch, ts, _, err = p.client.UpdateMessage(channelID, timestamp, slack.MsgOptionText(content, false))
		return err
	}); err != nil {
		return nil, fmt.Errorf("slack_edit_message: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"channel": ch, "timestamp": ts}}, nil
}

// --- step.slack_delete_message ---

type deleteMessageStep struct{ providerName string }

func (s *deleteMessageStep) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current, _, _ map[string]any) (*sdk.StepResult, error) {
	p, err := getProvider(s.providerName)
	if err != nil {
		return nil, err
	}
	channelID := stringFrom(current, "channel_id")
	timestamp := stringFrom(current, "timestamp")
	if channelID == "" || timestamp == "" {
		return nil, fmt.Errorf("slack_delete_message: channel_id and timestamp are required")
	}
	if err := withRateLimit(func() error {
		_, _, err = p.client.DeleteMessage(channelID, timestamp)
		return err
	}); err != nil {
		return nil, fmt.Errorf("slack_delete_message: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"deleted": true}}, nil
}

// --- step.slack_add_reaction ---

type addReactionStep struct{ providerName string }

func (s *addReactionStep) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current, _, _ map[string]any) (*sdk.StepResult, error) {
	p, err := getProvider(s.providerName)
	if err != nil {
		return nil, err
	}
	channelID := stringFrom(current, "channel_id")
	timestamp := stringFrom(current, "timestamp")
	emoji := stringFrom(current, "emoji")
	if channelID == "" || timestamp == "" || emoji == "" {
		return nil, fmt.Errorf("slack_add_reaction: channel_id, timestamp, and emoji are required")
	}
	if err := withRateLimit(func() error {
		return p.client.AddReaction(emoji, slack.ItemRef{Channel: channelID, Timestamp: timestamp})
	}); err != nil {
		return nil, fmt.Errorf("slack_add_reaction: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"ok": true}}, nil
}

// --- step.slack_upload_file ---

type uploadFileStep struct{ providerName string }

func (s *uploadFileStep) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current, _, _ map[string]any) (*sdk.StepResult, error) {
	p, err := getProvider(s.providerName)
	if err != nil {
		return nil, err
	}
	channelID := stringFrom(current, "channel_id")
	filename := stringFrom(current, "filename")
	content := stringFrom(current, "content")
	if channelID == "" || filename == "" || content == "" {
		return nil, fmt.Errorf("slack_upload_file: channel_id, filename, and content are required")
	}
	var fileID string
	if err := withRateLimit(func() error {
		f, err := p.client.UploadFile(slack.UploadFileParameters{
			Channel:  channelID,
			Filename: filename,
			Content:  content,
		})
		if err != nil {
			return err
		}
		fileID = f.ID
		return nil
	}); err != nil {
		return nil, fmt.Errorf("slack_upload_file: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"file_id": fileID}}, nil
}

// --- step.slack_send_thread_reply ---

type sendThreadReplyStep struct{ providerName string }

func (s *sendThreadReplyStep) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current, _, _ map[string]any) (*sdk.StepResult, error) {
	p, err := getProvider(s.providerName)
	if err != nil {
		return nil, err
	}
	channelID := stringFrom(current, "channel_id")
	threadTS := stringFrom(current, "thread_ts")
	content := stringFrom(current, "content")
	if channelID == "" || threadTS == "" || content == "" {
		return nil, fmt.Errorf("slack_send_thread_reply: channel_id, thread_ts, and content are required")
	}
	var ch, ts string
	if err := withRateLimit(func() error {
		ch, ts, _, err = p.client.SendMessage(channelID,
			slack.MsgOptionText(content, false),
			slack.MsgOptionTS(threadTS),
		)
		return err
	}); err != nil {
		return nil, fmt.Errorf("slack_send_thread_reply: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"channel": ch, "timestamp": ts}}, nil
}

// --- step.slack_set_topic ---

type setTopicStep struct{ providerName string }

func (s *setTopicStep) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current, _, _ map[string]any) (*sdk.StepResult, error) {
	p, err := getProvider(s.providerName)
	if err != nil {
		return nil, err
	}
	channelID := stringFrom(current, "channel_id")
	topic := stringFrom(current, "topic")
	if channelID == "" || topic == "" {
		return nil, fmt.Errorf("slack_set_topic: channel_id and topic are required")
	}
	if err := withRateLimit(func() error {
		_, err = p.client.SetTopicOfConversation(channelID, topic)
		return err
	}); err != nil {
		return nil, fmt.Errorf("slack_set_topic: %w", err)
	}
	return &sdk.StepResult{Output: map[string]any{"channel": channelID, "topic": topic}}, nil
}
