package internal

import (
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// slackPlugin implements PluginProvider, ModuleProvider, StepProvider, TriggerProvider, and SchemaProvider.
type slackPlugin struct{}

// New returns a new slackPlugin ready to serve.
func New() *slackPlugin { return &slackPlugin{} }

// Manifest returns plugin metadata.
func (p *slackPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "slack",
		Version:     "0.1.0",
		Author:      "GoCodeAlone",
		Description: "Slack messaging, file uploads, and real-time event triggers via Socket Mode",
	}
}

// ModuleTypes returns the module type names this plugin provides.
func (p *slackPlugin) ModuleTypes() []string {
	return []string{"slack.provider"}
}

// CreateModule creates a module instance of the given type.
func (p *slackPlugin) CreateModule(typeName, name string, config map[string]any) (sdk.ModuleInstance, error) {
	switch typeName {
	case "slack.provider":
		return newSlackProvider(name, config), nil
	default:
		return nil, fmt.Errorf("slack plugin: unknown module type %q", typeName)
	}
}

// StepTypes returns the step type names this plugin provides.
func (p *slackPlugin) StepTypes() []string {
	return []string{
		"step.slack_send_message",
		"step.slack_send_blocks",
		"step.slack_edit_message",
		"step.slack_delete_message",
		"step.slack_add_reaction",
		"step.slack_upload_file",
		"step.slack_send_thread_reply",
		"step.slack_set_topic",
	}
}

// CreateStep creates a step instance of the given type.
func (p *slackPlugin) CreateStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	providerName, _ := config["provider"].(string)
	if providerName == "" {
		return nil, fmt.Errorf("slack plugin: step %q requires a 'provider' config field", typeName)
	}

	switch typeName {
	case "step.slack_send_message":
		return &sendMessageStep{providerName: providerName}, nil
	case "step.slack_send_blocks":
		return &sendBlocksStep{providerName: providerName}, nil
	case "step.slack_edit_message":
		return &editMessageStep{providerName: providerName}, nil
	case "step.slack_delete_message":
		return &deleteMessageStep{providerName: providerName}, nil
	case "step.slack_add_reaction":
		return &addReactionStep{providerName: providerName}, nil
	case "step.slack_upload_file":
		return &uploadFileStep{providerName: providerName}, nil
	case "step.slack_send_thread_reply":
		return &sendThreadReplyStep{providerName: providerName}, nil
	case "step.slack_set_topic":
		return &setTopicStep{providerName: providerName}, nil
	default:
		return nil, fmt.Errorf("slack plugin: unknown step type %q", typeName)
	}
}

// TriggerTypes returns the trigger type names this plugin provides.
func (p *slackPlugin) TriggerTypes() []string {
	return []string{"trigger.slack"}
}

// CreateTrigger creates a trigger instance of the given type.
func (p *slackPlugin) CreateTrigger(typeName string, config map[string]any, cb sdk.TriggerCallback) (sdk.TriggerInstance, error) {
	switch typeName {
	case "trigger.slack":
		return newSlackTrigger(config, cb)
	default:
		return nil, fmt.Errorf("slack plugin: unknown trigger type %q", typeName)
	}
}

// ModuleSchemas returns UI schema metadata for module configuration.
func (p *slackPlugin) ModuleSchemas() []sdk.ModuleSchemaData {
	return []sdk.ModuleSchemaData{
		{
			Type:        "slack.provider",
			Label:       "Slack Provider",
			Category:    "Messaging",
			Description: "Slack bot client using bot token and Socket Mode app token",
			ConfigFields: []sdk.ConfigField{
				{Name: "bot_token", Type: "string", Description: "Slack bot token (xoxb-...)", Required: true},
				{Name: "app_token", Type: "string", Description: "Slack app-level token for Socket Mode (xapp-...)", Required: true},
			},
		},
	}
}
