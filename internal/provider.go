package internal

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	messaging "github.com/GoCodeAlone/workflow-plugin-messaging-core"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Compile-time interface satisfaction check.
var _ messaging.Provider = (*slackProvider)(nil)

// slackProvider is the slack.provider module. It holds the Slack bot client and
// Socket Mode client, exposing them to steps and the trigger via the provider registry.
type slackProvider struct {
	name         string
	botToken     string
	appToken     string
	client       *slack.Client
	socketClient *socketmode.Client
}

// provider registry maps provider names to live instances so steps can look them up.
var (
	providersMu sync.RWMutex
	providers   = map[string]*slackProvider{}
)

func newSlackProvider(name string, config map[string]any) *slackProvider {
	return &slackProvider{
		name:     name,
		botToken: stringFrom(config, "bot_token"),
		appToken: stringFrom(config, "app_token"),
	}
}

// Init creates the Slack clients.
func (m *slackProvider) Init() error {
	if m.botToken == "" {
		return fmt.Errorf("slack.provider %q: bot_token is required", m.name)
	}
	if m.appToken == "" {
		return fmt.Errorf("slack.provider %q: app_token is required (Socket Mode requires an xapp- token)", m.name)
	}

	m.client = slack.New(
		m.botToken,
		slack.OptionAppLevelToken(m.appToken),
	)
	m.socketClient = socketmode.New(m.client)

	providersMu.Lock()
	providers[m.name] = m
	providersMu.Unlock()

	return nil
}

// Start opens the Socket Mode WebSocket connection in the background.
func (m *slackProvider) Start(ctx context.Context) error {
	go func() {
		if err := m.socketClient.RunContext(ctx); err != nil && ctx.Err() == nil {
			// Context cancellation is normal shutdown — only unexpected errors matter here.
			_ = err
		}
	}()
	return nil
}

// Stop removes the provider from the registry.
func (m *slackProvider) Stop(_ context.Context) error {
	providersMu.Lock()
	delete(providers, m.name)
	providersMu.Unlock()
	return nil
}

// getProvider retrieves a registered provider by name.
func getProvider(name string) (*slackProvider, error) {
	providersMu.RLock()
	p, ok := providers[name]
	providersMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("slack provider %q not found (is it started?)", name)
	}
	return p, nil
}

// stringFrom safely extracts a string value from a map.
func stringFrom(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

// --- messaging.Provider implementation ---

// Name returns the platform identifier.
func (m *slackProvider) Name() string { return "slack" }

// SendMessage sends a plain text message and returns the message timestamp.
func (m *slackProvider) SendMessage(_ context.Context, channelID, content string, _ *messaging.MessageOpts) (string, error) {
	_, ts, _, err := m.client.SendMessage(channelID, slack.MsgOptionText(content, false))
	if err != nil {
		return "", fmt.Errorf("slack send: %w", err)
	}
	return ts, nil
}

// EditMessage updates an existing message identified by its timestamp.
func (m *slackProvider) EditMessage(_ context.Context, channelID, messageID, content string) error {
	_, _, _, err := m.client.UpdateMessage(channelID, messageID, slack.MsgOptionText(content, false))
	return err
}

// DeleteMessage removes a message.
func (m *slackProvider) DeleteMessage(_ context.Context, channelID, messageID string) error {
	_, _, err := m.client.DeleteMessage(channelID, messageID)
	return err
}

// SendReply sends a threaded reply and returns the reply's timestamp.
func (m *slackProvider) SendReply(_ context.Context, channelID, parentID, content string, _ *messaging.MessageOpts) (string, error) {
	_, ts, _, err := m.client.SendMessage(channelID,
		slack.MsgOptionText(content, false),
		slack.MsgOptionTS(parentID),
	)
	if err != nil {
		return "", fmt.Errorf("slack reply: %w", err)
	}
	return ts, nil
}

// React adds an emoji reaction to a message.
func (m *slackProvider) React(_ context.Context, channelID, messageID, emoji string) error {
	return m.client.AddReaction(emoji, slack.ItemRef{Channel: channelID, Timestamp: messageID})
}

// UploadFile sends a file to a channel and returns the file ID.
func (m *slackProvider) UploadFile(_ context.Context, channelID string, file io.Reader, filename string) (string, error) {
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, file); err != nil {
		return "", fmt.Errorf("slack upload read: %w", err)
	}
	f, err := m.client.UploadFile(slack.UploadFileParameters{
		Channel:  channelID,
		Filename: filename,
		Content:  buf.String(),
	})
	if err != nil {
		return "", fmt.Errorf("slack upload: %w", err)
	}
	return f.ID, nil
}
