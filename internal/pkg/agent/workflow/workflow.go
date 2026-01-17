// Package workflow provides ADK Flow workflow agent implementations.
package workflow

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// WorkflowMode defines the workflow execution mode.
type WorkflowMode string

const (
	// ModeSequential executes sub-agents sequentially.
	ModeSequential WorkflowMode = "sequential"
	// ModeLoop executes sub-agents in a loop until max iterations.
	ModeLoop WorkflowMode = "loop"
	// ModeParallel executes sub-agents in parallel.
	ModeParallel WorkflowMode = "parallel"
)

// Config workflow configuration.
type Config struct {
	// Name workflow name
	Name string

	// Description workflow description
	Description string

	// Mode workflow execution mode
	Mode WorkflowMode

	// SubAgents list of sub-agents
	SubAgents []adk.Agent

	// MaxIterations max iterations for loop mode (default: 10)
	MaxIterations int
}

// Workflow ADK Flow workflow agent wrapper.
type Workflow struct {
	name  string
	mode  WorkflowMode
	agent adk.ResumableAgent
}

// New creates a new workflow agent.
func New(ctx context.Context, cfg *Config) (*Workflow, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if len(cfg.SubAgents) == 0 {
		return nil, fmt.Errorf("at least one sub agent is required")
	}

	var agent adk.ResumableAgent
	var err error

	switch cfg.Mode {
	case ModeSequential, "":
		agent, err = adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
			Name:        cfg.Name,
			Description: cfg.Description,
			SubAgents:   cfg.SubAgents,
		})

	case ModeLoop:
		maxIter := cfg.MaxIterations
		if maxIter <= 0 {
			maxIter = 10
		}
		agent, err = adk.NewLoopAgent(ctx, &adk.LoopAgentConfig{
			Name:          cfg.Name,
			Description:   cfg.Description,
			SubAgents:     cfg.SubAgents,
			MaxIterations: maxIter,
		})

	case ModeParallel:
		agent, err = adk.NewParallelAgent(ctx, &adk.ParallelAgentConfig{
			Name:        cfg.Name,
			Description: cfg.Description,
			SubAgents:   cfg.SubAgents,
		})

	default:
		return nil, fmt.Errorf("unsupported workflow mode: %s", cfg.Mode)
	}

	if err != nil {
		return nil, fmt.Errorf("create workflow agent: %w", err)
	}

	return &Workflow{
		name:  cfg.Name,
		mode:  cfg.Mode,
		agent: agent,
	}, nil
}

// Run executes the workflow.
func (w *Workflow) Run(ctx context.Context, input string) (*schema.Message, error) {
	messages := []adk.Message{
		schema.UserMessage(input),
	}

	iter := w.agent.Run(ctx, &adk.AgentInput{
		Messages: messages,
	})

	var lastMessage *schema.Message
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if msg, _, err := adk.GetMessage(event); err == nil {
			lastMessage = msg
		}
	}

	return lastMessage, nil
}

// Stream executes the workflow with streaming output.
func (w *Workflow) Stream(ctx context.Context, input string) *adk.AsyncIterator[*adk.AgentEvent] {
	messages := []adk.Message{
		schema.UserMessage(input),
	}

	return w.agent.Run(ctx, &adk.AgentInput{
		Messages:        messages,
		EnableStreaming: true,
	})
}

// Resume resumes an interrupted workflow.
func (w *Workflow) Resume(ctx context.Context, resumeInfo *adk.ResumeInfo) *adk.AsyncIterator[*adk.AgentEvent] {
	return w.agent.Resume(ctx, resumeInfo)
}

// Name returns the workflow name.
func (w *Workflow) Name() string {
	return w.name
}

// Mode returns the workflow mode.
func (w *Workflow) Mode() WorkflowMode {
	return w.mode
}

// GetAgent returns the underlying ResumableAgent.
func (w *Workflow) GetAgent() adk.ResumableAgent {
	return w.agent
}
