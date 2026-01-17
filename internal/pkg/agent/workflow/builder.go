// Package workflow provides workflow builder for ADK Flow.
package workflow

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
)

// Builder workflow builder with fluent API.
type Builder struct {
	name          string
	description   string
	mode          WorkflowMode
	subAgents     []adk.Agent
	maxIterations int
}

// NewBuilder creates a new workflow builder.
func NewBuilder(name string) *Builder {
	return &Builder{
		name:          name,
		mode:          ModeSequential,
		maxIterations: 10,
	}
}

// WithDescription sets the workflow description.
func (b *Builder) WithDescription(desc string) *Builder {
	b.description = desc
	return b
}

// Sequential sets the workflow to sequential mode.
func (b *Builder) Sequential() *Builder {
	b.mode = ModeSequential
	return b
}

// Loop sets the workflow to loop mode with max iterations.
func (b *Builder) Loop(maxIterations int) *Builder {
	b.mode = ModeLoop
	b.maxIterations = maxIterations
	return b
}

// Parallel sets the workflow to parallel mode.
func (b *Builder) Parallel() *Builder {
	b.mode = ModeParallel
	return b
}

// AddAgent adds a sub-agent to the workflow.
func (b *Builder) AddAgent(agent adk.Agent) *Builder {
	b.subAgents = append(b.subAgents, agent)
	return b
}

// AddAgents adds multiple sub-agents to the workflow.
func (b *Builder) AddAgents(agents ...adk.Agent) *Builder {
	b.subAgents = append(b.subAgents, agents...)
	return b
}

// Build creates the workflow.
func (b *Builder) Build(ctx context.Context) (*Workflow, error) {
	if b.name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if len(b.subAgents) == 0 {
		return nil, fmt.Errorf("at least one sub agent is required")
	}

	return New(ctx, &Config{
		Name:          b.name,
		Description:   b.description,
		Mode:          b.mode,
		SubAgents:     b.subAgents,
		MaxIterations: b.maxIterations,
	})
}

// SequentialWorkflow creates a sequential workflow directly.
func SequentialWorkflow(ctx context.Context, name string, agents ...adk.Agent) (*Workflow, error) {
	return New(ctx, &Config{
		Name:      name,
		Mode:      ModeSequential,
		SubAgents: agents,
	})
}

// LoopWorkflow creates a loop workflow directly.
func LoopWorkflow(ctx context.Context, name string, maxIterations int, agents ...adk.Agent) (*Workflow, error) {
	return New(ctx, &Config{
		Name:          name,
		Mode:          ModeLoop,
		SubAgents:     agents,
		MaxIterations: maxIterations,
	})
}

// ParallelWorkflow creates a parallel workflow directly.
func ParallelWorkflow(ctx context.Context, name string, agents ...adk.Agent) (*Workflow, error) {
	return New(ctx, &Config{
		Name:      name,
		Mode:      ModeParallel,
		SubAgents: agents,
	})
}
