package executor

import (
	"article-assistant/internal/domain"
	"context"
)

// TaskCommand is the command interface for all query types
type TaskCommand interface {
	Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error)
}

// Executor with Registry
type Executor struct {
	commands map[string]TaskCommand
}

func NewExecutor() *Executor {
	return &Executor{commands: make(map[string]TaskCommand)}
}

func (e *Executor) Register(name string, cmd TaskCommand) {
	e.commands[name] = cmd
}

func (e *Executor) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	cmd, ok := e.commands[plan.Command]
	if !ok {
		return &domain.ChatResponse{
			Answer: "Command not supported: " + plan.Command,
			Task:   plan.Command,
		}, nil
	}
	return cmd.Execute(ctx, plan, query)
}
