package internal

import (
	"context"
	"testing"
)

func TestManifest(t *testing.T) {
	p := New()
	m := p.Manifest()
	if m.Name != "slack" {
		t.Errorf("expected manifest name %q, got %q", "slack", m.Name)
	}
	if m.Version == "" {
		t.Error("expected non-empty version")
	}
}

func TestModuleTypes(t *testing.T) {
	p := New()
	types := p.ModuleTypes()
	if len(types) == 0 {
		t.Fatal("expected at least one module type")
	}
	if types[0] != "slack.provider" {
		t.Errorf("expected %q, got %q", "slack.provider", types[0])
	}
}

func TestStepTypes(t *testing.T) {
	p := New()
	types := p.StepTypes()
	want := []string{
		"step.slack_send_message",
		"step.slack_send_blocks",
		"step.slack_edit_message",
		"step.slack_delete_message",
		"step.slack_add_reaction",
		"step.slack_upload_file",
		"step.slack_send_thread_reply",
		"step.slack_set_topic",
	}
	if len(types) != len(want) {
		t.Fatalf("expected %d step types, got %d", len(want), len(types))
	}
	for i, w := range want {
		if types[i] != w {
			t.Errorf("step[%d]: expected %q, got %q", i, w, types[i])
		}
	}
}

func TestTriggerTypes(t *testing.T) {
	p := New()
	types := p.TriggerTypes()
	if len(types) == 0 || types[0] != "trigger.slack" {
		t.Errorf("expected trigger.slack, got %v", types)
	}
}

func TestCreateModuleUnknownType(t *testing.T) {
	p := New()
	_, err := p.CreateModule("unknown.type", "test", nil)
	if err == nil {
		t.Error("expected error for unknown module type")
	}
}

func TestCreateStepMissingProvider(t *testing.T) {
	p := New()
	_, err := p.CreateStep("step.slack_send_message", "test", map[string]any{})
	if err == nil {
		t.Error("expected error when provider config field is missing")
	}
}

func TestCreateStepUnknownType(t *testing.T) {
	p := New()
	_, err := p.CreateStep("step.unknown", "test", map[string]any{"provider": "my-slack"})
	if err == nil {
		t.Error("expected error for unknown step type")
	}
}

func TestCreateTriggerMissingProvider(t *testing.T) {
	p := New()
	_, err := p.CreateTrigger("trigger.slack", map[string]any{}, func(action string, data map[string]any) error { return nil })
	if err == nil {
		t.Error("expected error when provider config field is missing")
	}
}

func TestCreateTriggerUnknownType(t *testing.T) {
	p := New()
	_, err := p.CreateTrigger("trigger.unknown", map[string]any{"provider": "my-slack"}, func(action string, data map[string]any) error { return nil })
	if err == nil {
		t.Error("expected error for unknown trigger type")
	}
}

func TestStepExecuteMissingProvider(t *testing.T) {
	step := &sendMessageStep{providerName: "nonexistent"}
	_, err := step.Execute(context.Background(), nil, nil, map[string]any{
		"channel_id": "C123",
		"content":    "hello",
	}, nil, nil)
	if err == nil {
		t.Error("expected error when provider is not registered")
	}
}

func withTestProvider(t *testing.T) {
	t.Helper()
	prov := &slackProvider{name: "test-slack"}
	providersMu.Lock()
	providers["test-slack"] = prov
	providersMu.Unlock()
	t.Cleanup(func() {
		providersMu.Lock()
		delete(providers, "test-slack")
		providersMu.Unlock()
	})
}

func TestStepsMissingRequiredFields(t *testing.T) {
	withTestProvider(t)

	steps := []struct {
		name string
		fn   func() error
	}{
		{"send_message", func() error {
			_, err := (&sendMessageStep{providerName: "test-slack"}).Execute(context.Background(), nil, nil, map[string]any{}, nil, nil)
			return err
		}},
		{"edit_message", func() error {
			_, err := (&editMessageStep{providerName: "test-slack"}).Execute(context.Background(), nil, nil, map[string]any{}, nil, nil)
			return err
		}},
		{"delete_message", func() error {
			_, err := (&deleteMessageStep{providerName: "test-slack"}).Execute(context.Background(), nil, nil, map[string]any{}, nil, nil)
			return err
		}},
		{"add_reaction", func() error {
			_, err := (&addReactionStep{providerName: "test-slack"}).Execute(context.Background(), nil, nil, map[string]any{}, nil, nil)
			return err
		}},
		{"upload_file", func() error {
			_, err := (&uploadFileStep{providerName: "test-slack"}).Execute(context.Background(), nil, nil, map[string]any{}, nil, nil)
			return err
		}},
		{"send_thread_reply", func() error {
			_, err := (&sendThreadReplyStep{providerName: "test-slack"}).Execute(context.Background(), nil, nil, map[string]any{}, nil, nil)
			return err
		}},
		{"set_topic", func() error {
			_, err := (&setTopicStep{providerName: "test-slack"}).Execute(context.Background(), nil, nil, map[string]any{}, nil, nil)
			return err
		}},
	}

	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err == nil {
				t.Errorf("%s: expected validation error for missing required fields", tt.name)
			}
		})
	}
}

func TestProviderInitMissingTokens(t *testing.T) {
	p := newSlackProvider("test", map[string]any{})
	if err := p.Init(); err == nil {
		t.Error("expected error when bot_token is missing")
	}
	p2 := newSlackProvider("test2", map[string]any{"bot_token": "xoxb-fake"})
	if err := p2.Init(); err == nil {
		t.Error("expected error when app_token is missing")
	}
}
