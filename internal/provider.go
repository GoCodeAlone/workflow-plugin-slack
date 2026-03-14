package internal

import (
	"context"
	"fmt"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

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
