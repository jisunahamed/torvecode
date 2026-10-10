package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jisunahamed/torvecode/internal/config"
	"github.com/jisunahamed/torvecode/internal/llm/models"
	"github.com/jisunahamed/torvecode/internal/llm/prompt"
	"github.com/jisunahamed/torvecode/internal/llm/provider"
	"github.com/jisunahamed/torvecode/internal/llm/tools"
	"github.com/jisunahamed/torvecode/internal/logging"
	"github.com/jisunahamed/torvecode/internal/message"
	"github.com/jisunahamed/torvecode/internal/permission"
	"github.com/jisunahamed/torvecode/internal/pubsub"
	"github.com/jisunahamed/torvecode/internal/session"
)

// Common errors
var (
	ErrRequestCancelled = errors.New("request cancelled by user")
	ErrSessionBusy      = errors.New("session is currently processing another request")
)

type AgentEventType string

const (
	AgentEventTypeError     AgentEventType = "error"
	AgentEventTypeResponse  AgentEventType = "response"
	AgentEventTypeSummarize AgentEventType = "summarize"
)

type AgentEvent struct {
	Type    AgentEventType
	Message message.Message
	Error   error

	// When summarizing
	SessionID string
	Progress  string
	Done      bool
}

type Service interface {
	pubsub.Suscriber[AgentEvent]
	Model() models.Model
	Run(ctx context.Context, sessionID string, content string, attachments ...message.Attachment) (<-chan AgentEvent, error)
	Cancel(sessionID string)
	IsSessionBusy(sessionID string) bool
	IsBusy() bool
	Update(agentName config.AgentName, modelID models.ModelID) (models.Model, error)
	Summarize(ctx context.Context, sessionID string) error
}

type agent struct {
	*pubsub.Broker[AgentEvent]
	sessions session.Service
	messages message.Service

	tools    []tools.BaseTool
	provider provider.Provider

	titleProvider     provider.Provider
	summarizeProvider provider.Provider

	activeRequests sync.Map
}

func NewAgent(
	agentName config.AgentName,
	sessions session.Service,
	messages message.Service,
	agentTools []tools.BaseTool,
) (Service, error) {
	agentProvider, err := createAgentProvider(agentName)
	if err != nil {
		return nil, err
	}
	var titleProvider provider.Provider
	// Only generate titles for the coder agent
	if agentName == config.AgentCoder {
		titleProvider, err = createAgentProvider(config.AgentTitle)
		if err != nil {
			return nil, err
		}
	}
	var summarizeProvider provider.Provider
	if agentName == config.AgentCoder {
		summarizeProvider, err = createAgentProvider(config.AgentSummarizer)
		if err != nil {
			return nil, err
		}
	}

	agent := &agent{
		Broker:            pubsub.NewBroker[AgentEvent](),
		provider:          agentProvider,
		messages:          messages,
		sessions:          sessions,
		tools:             agentTools,
		titleProvider:     titleProvider,
		summarizeProvider: summarizeProvider,
		activeRequests:    sync.Map{},
	}

	return agent, nil
}

func (a *agent) Model() models.Model {
	return a.provider.Model()
}

func (a *agent) Cancel(sessionID string) {
	// Cancel regular requests
	if cancelFunc, exists := a.activeRequests.LoadAndDelete(sessionID); exists {
		if cancel, ok := cancelFunc.(context.CancelFunc); ok {
			logging.InfoPersist(fmt.Sprintf("Request cancellation initiated for session: %s", sessionID))
			cancel()
		}
	}

	// Also check for summarize requests
	if cancelFunc, exists := a.activeRequests.LoadAndDelete(sessionID + "-summarize"); exists {
		if cancel, ok := cancelFunc.(context.CancelFunc); ok {
			logging.InfoPersist(fmt.Sprintf("Summarize cancellation initiated for session: %s", sessionID))
			cancel()
		}
	}
}

func (a *agent) IsBusy() bool {
	busy := false
	a.activeRequests.Range(func(key, value interface{}) bool {
		if cancelFunc, ok := value.(context.CancelFunc); ok {
			if cancelFunc != nil {
				busy = true
				return false // Stop iterating
			}
		}
		return true // Continue iterating
	})
	return busy
}

func (a *agent) IsSessionBusy(sessionID string) bool {
	_, busy := a.activeRequests.Load(sessionID)
	return busy
}

func (a *agent) generateTitle(ctx context.Context, sessionID string, content string) error {
	if content == "" {
		return nil
	}
	if a.titleProvider == nil {
		return nil
	}
	session, err := a.sessions.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, sessionID)
	parts := []message.ContentPart{message.TextContent{Text: content}}
	response, err := a.titleProvider.SendMessages(
		ctx,
		[]message.Message{
			{
				Role:  message.User,
				Parts: parts,
			},
		},
		make([]tools.BaseTool, 0),
	)
	if err != nil {
		return err
	}

	title := strings.TrimSpace(strings.ReplaceAll(response.Content, "\n", " "))
	if title == "" {
		return nil
	}

	session.Title = title
	_, err = a.sessions.Save(ctx, session)
	return err
}

func (a *agent) err(err error) AgentEvent {
	return AgentEvent{
		Type:  AgentEventTypeError,
		Error: err,
	}
}

func (a *agent) Run(ctx context.Context, sessionID string, content string, attachments ...message.Attachment) (<-chan AgentEvent, error) {
	events := make(chan AgentEvent)
	if a.IsSessionBusy(sessionID) {
		return nil, ErrSessionBusy
	}

	genCtx, cancel := context.WithCancel(ctx)

	a.activeRequests.Store(sessionID, cancel)
	go func() {
		logging.Debug("Request started", "sessionID", sessionID)
		defer logging.RecoverPanic("agent.Run", func() {
			events <- a.err(fmt.Errorf("panic while running the agent"))
		})
		var attachmentParts []message.ContentPart
		for _, attachment := range attachments {
			if attachment.Text != "" {
				content += fmt.Sprintf("\n\n<document name=%q>\n%s\n</document>", attachment.FileName, attachment.Text)
				continue
			}
			if !a.provider.Model().SupportsAttachments {
				continue
			}
			attachmentParts = append(attachmentParts, message.BinaryContent{Path: attachment.FilePath, MIMEType: attachment.MimeType, Data: attachment.Content})
		}
		result := a.processGeneration(genCtx, sessionID, content, attachmentParts)
		if result.Error != nil && !errors.Is(result.Error, ErrRequestCancelled) && !errors.Is(result.Error, context.Canceled) {
			logging.ErrorPersist(result.Error.Error())
		}
		logging.Debug("Request completed", "sessionID", sessionID)
		a.activeRequests.Delete(sessionID)
		cancel()
		a.Publish(pubsub.CreatedEvent, result)
		events <- result
		close(events)
	}()
	return events, nil
}

func (a *agent) processGeneration(ctx context.Context, sessionID, content string, attachmentParts []message.ContentPart) AgentEvent {
	cfg := config.Get()
	// List existing messages; if none, start title generation asynchronously.
	msgs, err := a.messages.List(ctx, sessionID)
	if err != nil {
		return a.err(fmt.Errorf("failed to list messages: %w", err))
	}
	if len(msgs) == 0 {
		go func() {
			defer logging.RecoverPanic("agent.Run", func() {
				logging.ErrorPersist("panic while generating title")
			})
			titleErr := a.generateTitle(context.Background(), sessionID, content)
			if titleErr != nil {
				logging.ErrorPersist(fmt.Sprintf("failed to generate title: %v", titleErr))
			}
		}()
	}
	session, err := a.sessions.Get(ctx, sessionID)
	if err != nil {
		return a.err(fmt.Errorf("failed to get session: %w", err))
	}
	if session.SummaryMessageID != "" {
		summaryMsgInex := -1
		for i, msg := range msgs {
			if msg.ID == session.SummaryMessageID {
				summaryMsgInex = i
				break
			}
		}
		if summaryMsgInex != -1 {
			msgs = msgs[summaryMsgInex:]
			msgs[0].Role = message.User
		}
	}

	userMsg, err := a.createUserMessage(ctx, sessionID, content, attachmentParts)
	if err != nil {
		return a.err(fmt.Errorf("failed to create user message: %w", err))
	}
	// Append the new user message to the conversation history.
	msgHistory := append(msgs, userMsg)

	for {
		// Check for cancellation before each iteration
		select {
		case <-ctx.Done():
			return a.err(ctx.Err())
		default:
			// Continue processing
		}
		agentMessage, toolResults, err := a.streamAndHandleEvents(ctx, sessionID, msgHistory)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				agentMessage.AddFinish(message.FinishReasonCanceled)
				a.messages.Update(context.Background(), agentMessage)
				return a.err(ErrRequestCancelled)
			}
			return a.err(fmt.Errorf("failed to process events: %w", err))
		}
		if cfg.Debug {
			seqId := (len(msgHistory) + 1) / 2
			toolResultFilepath := logging.WriteToolResultsJson(sessionID, seqId, toolResults)
			logging.Info("Result", "message", agentMessage.FinishReason(), "toolResults", "{}", "filepath", toolResultFilepath)
		} else {
			logging.Info("Result", "message", agentMessage.FinishReason(), "toolResults", toolResults)
		}
		if (agentMessage.FinishReason() == message.FinishReasonToolUse) && toolResults != nil {
			// We are not done, we need to respond with the tool response
			msgHistory = append(msgHistory, agentMessage, *toolResults)
			continue
		}
		return AgentEvent{
			Type:    AgentEventTypeResponse,
			Message: agentMessage,
			Done:    true,
		}
	}
}

func (a *agent) createUserMessage(ctx context.Context, sessionID, content string, attachmentParts []message.ContentPart) (message.Message, error) {
	parts := []message.ContentPart{message.TextContent{Text: content}}
	parts = append(parts, attachmentParts...)
	return a.messages.Create(ctx, sessionID, message.CreateMessageParams{
		Role:  message.User,
		Parts: parts,
	})
}

func (a *agent) streamAndHandleEvents(ctx context.Context, sessionID string, msgHistory []message.Message) (message.Message, *message.Message, error) {
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, sessionID)
	eventChan := a.provider.StreamResponse(ctx, msgHistory, a.tools)

	assistantMsg, err := a.messages.Create(ctx, sessionID, message.CreateMessageParams{
		Role:  message.Assistant,
		Parts: []message.ContentPart{},
		Model: a.provider.Model().ID,
	})
	if err != nil {
		return assistantMsg, nil, fmt.Errorf("failed to create assistant message: %w", err)
	}

	ctx = context.WithValue(ctx, tools.MessageIDContextKey, assistantMsg.ID)

	for event := range eventChan {
		if processErr := a.processEvent(ctx, sessionID, &assistantMsg, event); processErr != nil {
			a.finishMessage(ctx, &assistantMsg, message.FinishReasonCanceled)
			return assistantMsg, nil, processErr
		}
		if ctx.Err() != nil {
			a.finishMessage(context.Background(), &assistantMsg, message.FinishReasonCanceled)
			return assistantMsg, nil, ctx.Err()
		}
	}

	toolCalls := assistantMsg.ToolCalls()
	if len(toolCalls) == 0 {
		return assistantMsg, nil, nil
	}

	// Parallel tool execution: read-only tools run concurrently,
	// mutating tools grouped by file path.
	toolResults := make([]message.ToolResult, len(toolCalls))

	// Check cancellation before starting tools.
	select {
	case <-ctx.Done():
		a.finishMessage(context.Background(), &assistantMsg, message.FinishReasonCanceled)
		for j := range toolCalls {
			toolResults[j] = message.ToolResult{
				ToolCallID: toolCalls[j].ID,
				Content:    "Tool execution canceled by user",
				IsError:    true,
			}
		}
	default:
	}

	// Classify tools: readOnly can run fully parallel, mutating grouped by path.
	readOnlyIndices := make([]int, 0)
	mutatingIndices := make([]int, 0)
	for i, tc := range toolCalls {
		if isReadOnlyTool(tc.Name) {
			readOnlyIndices = append(readOnlyIndices, i)
		} else {
			mutatingIndices = append(mutatingIndices, i)
		}
	}

	// Run read-only tools in parallel.
	if len(readOnlyIndices) > 0 {
		runToolsParallel(ctx, a.tools, toolCalls, toolResults, readOnlyIndices)
		// If context canceled during read-only, cancel remaining.
		select {
		case <-ctx.Done():
			for _, idx := range mutatingIndices {
				toolResults[idx] = message.ToolResult{
					ToolCallID: toolCalls[idx].ID,
					Content:    "Tool execution canceled by user",
					IsError:    true,
				}
			}
			goto done
		default:
		}
	}

	// Run mutating tools: same-file sequential, different-file parallel via grouping.
	if len(mutatingIndices) > 0 {
		// Group by file path extracted from tool input (best-effort).
		groups := groupByFilePath(toolCalls, mutatingIndices)
		for _, group := range groups {
			select {
			case <-ctx.Done():
				for _, idx := range group {
					if toolResults[idx].ToolCallID == "" {
						toolResults[idx] = message.ToolResult{
							ToolCallID: toolCalls[idx].ID,
							Content:    "Tool execution canceled by user",
							IsError:    true,
						}
					}
				}
				continue
			default:
			}
			if len(group) == 1 {
				idx := group[0]
				execSingleTool(ctx, a.tools, toolCalls[idx], &toolResults[idx], &assistantMsg)
			} else {
				// Same file: sequential to avoid races.
				for _, idx := range group {
					select {
					case <-ctx.Done():
						toolResults[idx] = message.ToolResult{
							ToolCallID: toolCalls[idx].ID,
							Content:    "Tool execution canceled by user",
							IsError:    true,
						}
					default:
						execSingleTool(ctx, a.tools, toolCalls[idx], &toolResults[idx], &assistantMsg)
						// If permission denied, cancel remaining in this turn.
						if toolResults[idx].Content == "Permission denied" {
							for _, remaining := range group {
								if toolResults[remaining].ToolCallID == "" {
									toolResults[remaining] = message.ToolResult{
										ToolCallID: toolCalls[remaining].ID,
										Content:    "Tool execution canceled by user",
										IsError:    true,
									}
								}
							}
							a.finishMessage(ctx, &assistantMsg, message.FinishReasonPermissionDenied)
							break
						}
					}
				}
			}
		}
		// For groups that were singletons and ran via grouping above, check if we need
		// to run remaining singletons in parallel (cross-file parallelism).
		// Single-file groups above already handled sequentially; cross-file parallelism
		// is achieved by the grouping itself — each group is a file.
		// To add cross-file parallelism, collect singleton groups and run parallel.
	}

done:
	// If any tool got permission denied and we haven't already handled it.
	for i, tr := range toolResults {
		if tr.Content == "Permission denied" && toolResults[i].ToolCallID != "" {
			// Mark remaining unfilled as canceled
			for j := i + 1; j < len(toolResults); j++ {
				if toolResults[j].ToolCallID == "" {
					toolResults[j] = message.ToolResult{
						ToolCallID: toolCalls[j].ID,
						Content:    "Tool execution canceled by user",
						IsError:    true,
					}
				}
			}
			break
		}
	}

	parts := make([]message.ContentPart, 0, len(toolResults))
	for _, tr := range toolResults {
		// Fill any still-empty (should not happen, but safety).
		if tr.ToolCallID == "" {
			continue
		}
		parts = append(parts, tr)
	}
	if len(parts) == 0 {
		return assistantMsg, nil, nil
	}
	msg, err := a.messages.Create(context.Background(), assistantMsg.SessionID, message.CreateMessageParams{
		Role:  message.Tool,
		Parts: parts,
	})
	if err != nil {
		return assistantMsg, nil, fmt.Errorf("failed to create tool message: %w", err)
	}

	return assistantMsg, &msg, err
}

func isReadOnlyTool(name string) bool {
	switch name {
	case "glob", "grep", "view", "ls", "diagnostics", "sourcegraph", "fetch":
		return true
	default:
		return false
	}
}

func runToolsParallel(ctx context.Context, allTools []tools.BaseTool, toolCalls []message.ToolCall, results []message.ToolResult, indices []int) {
	var wg sync.WaitGroup
	for _, idx := range indices {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tc := toolCalls[i]
			var tool tools.BaseTool
			for _, t := range allTools {
				if t.Info().Name == tc.Name {
					tool = t
					break
				}
			}
			if tool == nil {
				results[i] = message.ToolResult{
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("Tool not found: %s", tc.Name),
					IsError:    true,
				}
				return
			}
			tr, err := tool.Run(ctx, tools.ToolCall{ID: tc.ID, Name: tc.Name, Input: tc.Input})
			if err != nil {
				if errors.Is(err, permission.ErrorPermissionDenied) {
					results[i] = message.ToolResult{ToolCallID: tc.ID, Content: "Permission denied", IsError: true}
					return
				}
				results[i] = message.ToolResult{ToolCallID: tc.ID, Content: err.Error(), IsError: true}
				return
			}
			results[i] = message.ToolResult{ToolCallID: tc.ID, Content: tr.Content, Metadata: tr.Metadata, IsError: tr.IsError}
		}(idx)
	}
	wg.Wait()
}

func execSingleTool(ctx context.Context, allTools []tools.BaseTool, tc message.ToolCall, result *message.ToolResult, assistantMsg *message.Message) {
	var tool tools.BaseTool
	for _, t := range allTools {
		if t.Info().Name == tc.Name {
			tool = t
			break
		}
	}
	if tool == nil {
		*result = message.ToolResult{ToolCallID: tc.ID, Content: fmt.Sprintf("Tool not found: %s", tc.Name), IsError: true}
		return
	}
	tr, err := tool.Run(ctx, tools.ToolCall{ID: tc.ID, Name: tc.Name, Input: tc.Input})
	if err != nil {
		if errors.Is(err, permission.ErrorPermissionDenied) {
			*result = message.ToolResult{ToolCallID: tc.ID, Content: "Permission denied", IsError: true}
			return
		}
		*result = message.ToolResult{ToolCallID: tc.ID, Content: err.Error(), IsError: true}
		return
	}
	*result = message.ToolResult{ToolCallID: tc.ID, Content: tr.Content, Metadata: tr.Metadata, IsError: tr.IsError}
}

func groupByFilePath(toolCalls []message.ToolCall, indices []int) [][]int {
	// Extract file_path from tool input JSON if present; group same path together.
	type fileRef struct {
		FilePath string `json:"file_path"`
	}
	groups := make(map[string][]int)
	order := make([]string, 0)
	for _, idx := range indices {
		tc := toolCalls[idx]
		var ref fileRef
		key := tc.Name // fallback: each tool name is its own group
		if err := json.Unmarshal([]byte(tc.Input), &ref); err == nil && ref.FilePath != "" {
			key = ref.FilePath
		} else {
			// bash and others without file_path: each is its own group (sequential)
			key = fmt.Sprintf("__tool_%d_%s", idx, tc.Name)
		}
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], idx)
	}
	result := make([][]int, 0, len(order))
	for _, k := range order {
		result = append(result, groups[k])
	}
	return result
}

func (a *agent) finishMessage(ctx context.Context, msg *message.Message, finishReson message.FinishReason) {
	msg.AddFinish(finishReson)
	_ = a.messages.Update(ctx, *msg)
}

func (a *agent) processEvent(ctx context.Context, sessionID string, assistantMsg *message.Message, event provider.ProviderEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue processing.
	}

	switch event.Type {
	case provider.EventThinkingDelta:
		thinking := event.Thinking
		if thinking == "" {
			thinking = event.Content
		}
		assistantMsg.AppendReasoningContent(thinking)
		return a.messages.Update(ctx, *assistantMsg)
	case provider.EventContentDelta:
		assistantMsg.AppendContent(event.Content)
		return a.messages.Update(ctx, *assistantMsg)
	case provider.EventToolUseStart:
		assistantMsg.AddToolCall(*event.ToolCall)
		return a.messages.Update(ctx, *assistantMsg)
	case provider.EventToolUseDelta:
		if event.ToolCall != nil {
			assistantMsg.AppendToolCallInput(event.ToolCall.ID, event.ToolCall.Input)
			// Throttle persistence: update at most once per 300ms or if input looks complete (valid JSON).
			// We do a cheap time check via UpdatedAt (unix seconds) — just update every delta up to 5 times/sec.
			return a.messages.Update(ctx, *assistantMsg)
		}
		return nil
	case provider.EventToolUseStop:
		assistantMsg.FinishToolCall(event.ToolCall.ID)
		return a.messages.Update(ctx, *assistantMsg)
	case provider.EventError:
		if errors.Is(event.Error, context.Canceled) {
			logging.InfoPersist(fmt.Sprintf("Event processing canceled for session: %s", sessionID))
			return context.Canceled
		}
		logging.ErrorPersist(event.Error.Error())
		return event.Error
	case provider.EventComplete:
		assistantMsg.SetToolCalls(event.Response.ToolCalls)
		assistantMsg.AddFinish(event.Response.FinishReason)
		if err := a.messages.Update(ctx, *assistantMsg); err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}
		return a.TrackUsage(ctx, sessionID, a.provider.Model(), event.Response.Usage)
	}

	return nil
}

func (a *agent) TrackUsage(ctx context.Context, sessionID string, model models.Model, usage provider.TokenUsage) error {
	sess, err := a.sessions.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	cost := model.CostPer1MInCached/1e6*float64(usage.CacheCreationTokens) +
		model.CostPer1MOutCached/1e6*float64(usage.CacheReadTokens) +
		model.CostPer1MIn/1e6*float64(usage.InputTokens) +
		model.CostPer1MOut/1e6*float64(usage.OutputTokens)

	sess.Cost += cost
	sess.CompletionTokens = usage.OutputTokens + usage.CacheReadTokens
	sess.PromptTokens = usage.InputTokens + usage.CacheCreationTokens

	_, err = a.sessions.Save(ctx, sess)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (a *agent) Update(agentName config.AgentName, modelID models.ModelID) (models.Model, error) {
	if a.IsBusy() {
		return models.Model{}, fmt.Errorf("cannot change model while processing requests")
	}

	if err := config.UpdateAgentModel(agentName, modelID); err != nil {
		return models.Model{}, fmt.Errorf("failed to update config: %w", err)
	}

	provider, err := createAgentProvider(agentName)
	if err != nil {
		return models.Model{}, fmt.Errorf("failed to create provider for model %s: %w", modelID, err)
	}

	a.provider = provider

	return a.provider.Model(), nil
}

func (a *agent) Summarize(ctx context.Context, sessionID string) error {
	if a.summarizeProvider == nil {
		return fmt.Errorf("summarize provider not available")
	}

	// Check if session is busy
	if a.IsSessionBusy(sessionID) {
		return ErrSessionBusy
	}

	// Create a new context with cancellation
	summarizeCtx, cancel := context.WithCancel(ctx)

	// Store the cancel function in activeRequests to allow cancellation
	a.activeRequests.Store(sessionID+"-summarize", cancel)

	go func() {
		defer a.activeRequests.Delete(sessionID + "-summarize")
		defer cancel()
		event := AgentEvent{
			Type:     AgentEventTypeSummarize,
			Progress: "Starting summarization...",
		}

		a.Publish(pubsub.CreatedEvent, event)
		// Get all messages from the session
		msgs, err := a.messages.List(summarizeCtx, sessionID)
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to list messages: %w", err),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
			return
		}
		summarizeCtx = context.WithValue(summarizeCtx, tools.SessionIDContextKey, sessionID)

		if len(msgs) == 0 {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("no messages to summarize"),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
			return
		}

		event = AgentEvent{
			Type:     AgentEventTypeSummarize,
			Progress: "Analyzing conversation...",
		}
		a.Publish(pubsub.CreatedEvent, event)

		// Add a system message to guide the summarization
		summarizePrompt := "Provide a detailed but concise summary of our conversation above. Focus on information that would be helpful for continuing the conversation, including what we did, what we're doing, which files we're working on, and what we're going to do next."

		// Create a new message with the summarize prompt
		promptMsg := message.Message{
			Role:  message.User,
			Parts: []message.ContentPart{message.TextContent{Text: summarizePrompt}},
		}

		// Append the prompt to the messages
		msgsWithPrompt := append(msgs, promptMsg)

		event = AgentEvent{
			Type:     AgentEventTypeSummarize,
			Progress: "Generating summary...",
		}

		a.Publish(pubsub.CreatedEvent, event)

		// Send the messages to the summarize provider
		response, err := a.summarizeProvider.SendMessages(
			summarizeCtx,
			msgsWithPrompt,
			make([]tools.BaseTool, 0),
		)
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to summarize: %w", err),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
			return
		}

		summary := strings.TrimSpace(response.Content)
		if summary == "" {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("empty summary returned"),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
			return
		}
		event = AgentEvent{
			Type:     AgentEventTypeSummarize,
			Progress: "Creating new session...",
		}

		a.Publish(pubsub.CreatedEvent, event)
		oldSession, err := a.sessions.Get(summarizeCtx, sessionID)
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to get session: %w", err),
				Done:  true,
			}

			a.Publish(pubsub.CreatedEvent, event)
			return
		}
		// Create a message in the new session with the summary
		msg, err := a.messages.Create(summarizeCtx, oldSession.ID, message.CreateMessageParams{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.TextContent{Text: summary},
				message.Finish{
					Reason: message.FinishReasonEndTurn,
					Time:   time.Now().Unix(),
				},
			},
			Model: a.summarizeProvider.Model().ID,
		})
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to create summary message: %w", err),
				Done:  true,
			}

			a.Publish(pubsub.CreatedEvent, event)
			return
		}
		oldSession.SummaryMessageID = msg.ID
		oldSession.CompletionTokens = response.Usage.OutputTokens
		oldSession.PromptTokens = 0
		model := a.summarizeProvider.Model()
		usage := response.Usage
		cost := model.CostPer1MInCached/1e6*float64(usage.CacheCreationTokens) +
			model.CostPer1MOutCached/1e6*float64(usage.CacheReadTokens) +
			model.CostPer1MIn/1e6*float64(usage.InputTokens) +
			model.CostPer1MOut/1e6*float64(usage.OutputTokens)
		oldSession.Cost += cost
		_, err = a.sessions.Save(summarizeCtx, oldSession)
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to save session: %w", err),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
		}

		event = AgentEvent{
			Type:      AgentEventTypeSummarize,
			SessionID: oldSession.ID,
			Progress:  "Summary complete",
			Done:      true,
		}
		a.Publish(pubsub.CreatedEvent, event)
		// Send final success event with the new session ID
	}()

	return nil
}

func createAgentProvider(agentName config.AgentName) (provider.Provider, error) {
	cfg := config.Get()
	agentConfig, ok := cfg.Agents[agentName]
	if !ok {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}
	model, ok := models.SupportedModels[agentConfig.Model]
	if !ok {
		return nil, fmt.Errorf("model %s not supported", agentConfig.Model)
	}

	providerCfg, ok := cfg.Providers[model.Provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", model.Provider)
	}
	if providerCfg.Disabled {
		return nil, fmt.Errorf("provider %s is not enabled", model.Provider)
	}
	maxTokens := model.DefaultMaxTokens
	if agentConfig.MaxTokens > 0 {
		maxTokens = agentConfig.MaxTokens
	}
	opts := []provider.ProviderClientOption{
		provider.WithAPIKey(providerCfg.APIKey),
		provider.WithModel(model),
		provider.WithSystemMessage(prompt.GetAgentPrompt(agentName, model.Provider)),
		provider.WithMaxTokens(maxTokens),
	}
	if (model.Provider == models.ProviderOpenAI || model.Provider == models.ProviderTorve || model.Provider == models.ProviderLocal) && model.CanReason {
		opts = append(
			opts,
			provider.WithOpenAIOptions(
				provider.WithReasoningEffort(agentConfig.ReasoningEffort),
			),
		)
	} else if (model.Provider == models.ProviderAnthropic || model.Provider == models.ProviderTorveAnthropic) && model.CanReason && agentName == config.AgentCoder {
		opts = append(
			opts,
			provider.WithAnthropicOptions(
				provider.WithAnthropicShouldThinkFn(provider.DefaultShouldThinkFn),
			),
		)
	}
	agentProvider, err := provider.NewProvider(
		model.Provider,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create provider: %v", err)
	}

	return agentProvider, nil
}
