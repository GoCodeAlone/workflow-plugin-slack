package internal_test

import (
	"testing"

	"github.com/GoCodeAlone/workflow/wftest"
)

func TestIntegration_SendMessage(t *testing.T) {
	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  notify:
    steps:
      - name: send
        type: step.slack_send_message
        config:
          channel_id: "C0123456789"
          content: "hello from workflow"
      - name: confirm
        type: step.set
        config:
          values:
            sent: true
`),
		wftest.MockStep("step.slack_send_message", wftest.Returns(map[string]any{
			"ok": true, "channel": "C0123456789", "ts": "1234567890.123456", "timestamp": "1234567890.123456",
		})),
	)
	result := h.ExecutePipeline("notify", nil)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.Output["sent"] != true {
		t.Error("expected sent=true")
	}
}

func TestIntegration_SendThreadReply(t *testing.T) {
	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  reply:
    steps:
      - name: thread_reply
        type: step.slack_send_thread_reply
        config:
          channel_id: "C0123456789"
          thread_ts: "1234567890.000001"
          content: "thread reply"
      - name: record
        type: step.set
        config:
          values:
            replied: true
`),
		wftest.MockStep("step.slack_send_thread_reply", wftest.Returns(map[string]any{
			"ok": true, "channel": "C0123456789", "ts": "1234567890.123457",
			"timestamp": "1234567890.123457", "thread_ts": "1234567890.000001",
		})),
	)
	result := h.ExecutePipeline("reply", nil)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.Output["replied"] != true {
		t.Error("expected replied=true")
	}
}

func TestIntegration_AddReaction(t *testing.T) {
	rec := wftest.RecordStep("step.slack_add_reaction")
	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  react:
    steps:
      - name: reaction
        type: step.slack_add_reaction
        config:
          channel_id: "C0123456789"
          timestamp: "1234567890.123456"
          emoji: "thumbsup"
      - name: done
        type: step.set
        config:
          values:
            reacted: true
`),
		rec,
	)
	result := h.ExecutePipeline("react", nil)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.Output["reacted"] != true {
		t.Error("expected reacted=true")
	}
	if rec.CallCount() != 1 {
		t.Errorf("expected step.slack_add_reaction called once, got %d", rec.CallCount())
	}
}
