package agent

import (
	"context"

	"github.com/jisunahamed/torvecode/internal/history"
	"github.com/jisunahamed/torvecode/internal/llm/tools"
	"github.com/jisunahamed/torvecode/internal/lsp"
	"github.com/jisunahamed/torvecode/internal/message"
	"github.com/jisunahamed/torvecode/internal/permission"
	"github.com/jisunahamed/torvecode/internal/session"
)

func CoderAgentTools(
	permissions permission.Service,
	sessions session.Service,
	messages message.Service,
	history history.Service,
	lspClients map[string]*lsp.Client,
) []tools.BaseTool {
	ctx := context.Background()
	otherTools := GetMcpTools(ctx, permissions)
	if len(lspClients) > 0 {
		otherTools = append(otherTools, tools.NewDiagnosticsTool(lspClients))
	}
	return append(
		[]tools.BaseTool{
			tools.NewBashTool(permissions),
			tools.NewEditTool(lspClients, permissions, history),
			tools.NewFetchTool(permissions),
			tools.NewGlobTool(),
			tools.NewGrepTool(),
			tools.NewLsTool(),
			tools.NewSourcegraphTool(),
			tools.NewViewTool(lspClients),
			tools.NewPatchTool(lspClients, permissions, history),
			tools.NewWriteTool(lspClients, permissions, history),
			tools.NewSkillTool(),
			NewAgentTool(sessions, messages, lspClients),
		}, otherTools...,
	)
}

func TaskAgentTools(lspClients map[string]*lsp.Client) []tools.BaseTool {
	base := []tools.BaseTool{
		tools.NewGlobTool(),
		tools.NewGrepTool(),
		tools.NewLsTool(),
		tools.NewSourcegraphTool(),
		tools.NewViewTool(lspClients),
		tools.NewSkillTool(),
	}
	if len(lspClients) > 0 {
		base = append(base, tools.NewDiagnosticsTool(lspClients))
	}
	return base
}
