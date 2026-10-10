package permission

import (
	"errors"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/uuid"
	"github.com/jisunahamed/torvecode/internal/config"
	"github.com/jisunahamed/torvecode/internal/pubsub"
)

var ErrorPermissionDenied = errors.New("permission denied")

type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
	DecisionAsk   Decision = "ask"
)

type CreatePermissionRequest struct {
	SessionID   string `json:"session_id"`
	ToolName    string `json:"tool_name"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Params      any    `json:"params"`
	Path        string `json:"path"`
}

type PermissionRequest struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	ToolName    string `json:"tool_name"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Params      any    `json:"params"`
	Path        string `json:"path"`
}

type Service interface {
	pubsub.Suscriber[PermissionRequest]
	GrantPersistant(permission PermissionRequest)
	Grant(permission PermissionRequest)
	Deny(permission PermissionRequest)
	Request(opts CreatePermissionRequest) bool
	RequestBatch(opts []CreatePermissionRequest) []bool
	Check(opts CreatePermissionRequest) Decision
	AutoApproveSession(sessionID string)
	DenySession(sessionID string)
	Mode() config.PermissionMode
	SetMode(mode config.PermissionMode)
}

type permissionService struct {
	*pubsub.Broker[PermissionRequest]

	mu                  sync.RWMutex
	sessionPermissions  []PermissionRequest
	pendingRequests     sync.Map
	autoApproveSessions []string
	deniedSessions      []string
	mode                config.PermissionMode
}

func (s *permissionService) Mode() config.PermissionMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.mode != "" {
		return s.mode
	}
	return config.GetPermissionMode()
}

func (s *permissionService) SetMode(mode config.PermissionMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
}

func (s *permissionService) GrantPersistant(permission PermissionRequest) {
	respCh, ok := s.pendingRequests.Load(permission.ID)
	if ok {
		select {
		case respCh.(chan bool) <- true:
		default:
		}
	}
	s.mu.Lock()
	s.sessionPermissions = append(s.sessionPermissions, permission)
	s.mu.Unlock()
}

func (s *permissionService) Grant(permission PermissionRequest) {
	respCh, ok := s.pendingRequests.Load(permission.ID)
	if ok {
		select {
		case respCh.(chan bool) <- true:
		default:
		}
	}
}

func (s *permissionService) Deny(permission PermissionRequest) {
	respCh, ok := s.pendingRequests.Load(permission.ID)
	if ok {
		select {
		case respCh.(chan bool) <- false:
		default:
		}
	}
}

func (s *permissionService) Check(opts CreatePermissionRequest) Decision {
	s.mu.RLock()
	autoApprove := slices.Contains(s.autoApproveSessions, opts.SessionID)
	denied := slices.Contains(s.deniedSessions, opts.SessionID)
	mode := s.mode
	s.mu.RUnlock()

	if autoApprove {
		return DecisionAllow
	}
	if denied {
		return DecisionDeny
	}

	// Permission mode overrides
	effectiveMode := mode
	if effectiveMode == "" {
		effectiveMode = config.GetPermissionMode()
	}
	switch effectiveMode {
	case config.PermissionModeBypassPermissions:
		return DecisionAllow
	case config.PermissionModePlan:
		if config.IsReadOnlyTool(opts.ToolName) {
			return DecisionAllow
		}
		return DecisionDeny
	case config.PermissionModeAcceptEdits:
		if opts.ToolName == "edit" || opts.ToolName == "write" || opts.ToolName == "patch" {
			return DecisionAllow
		}
	}

	// Allow/deny lists (glob on Path and tool name)
	cfg := config.Get()
	if cfg != nil {
		for _, pat := range cfg.Permissions.DenyList {
			if matchRule(pat, opts.ToolName, opts.Path) {
				return DecisionDeny
			}
		}
		for _, pat := range cfg.Permissions.AllowList {
			if matchRule(pat, opts.ToolName, opts.Path) {
				return DecisionAllow
			}
		}
		for _, pat := range cfg.Permissions.AskList {
			if matchRule(pat, opts.ToolName, opts.Path) {
				return DecisionAsk
			}
		}
	}

	// Check persisted session permissions
	s.mu.RLock()
	defer s.mu.RUnlock()
	dir := filepath.Dir(opts.Path)
	if dir == "." {
		dir = config.WorkingDirectory()
	}
	for _, p := range s.sessionPermissions {
		if p.ToolName == opts.ToolName && p.Action == opts.Action && p.SessionID == opts.SessionID && p.Path == dir {
			return DecisionAllow
		}
	}
	return DecisionAsk
}

func matchRule(pattern, toolName, path string) bool {
	if pattern == toolName {
		return true
	}
	if ok, _ := doublestar.Match(pattern, toolName); ok {
		return true
	}
	if path != "" {
		if ok, _ := doublestar.Match(pattern, path); ok {
			return true
		}
	}
	return false
}

func (s *permissionService) Request(opts CreatePermissionRequest) bool {
	decision := s.Check(opts)
	if decision == DecisionAllow {
		return true
	}
	if decision == DecisionDeny {
		return false
	}

	dir := filepath.Dir(opts.Path)
	if dir == "." {
		dir = config.WorkingDirectory()
	}
	permission := PermissionRequest{
		ID:          uuid.New().String(),
		Path:        dir,
		SessionID:   opts.SessionID,
		ToolName:    opts.ToolName,
		Description: opts.Description,
		Action:      opts.Action,
		Params:      opts.Params,
	}

	respCh := make(chan bool, 1)
	s.pendingRequests.Store(permission.ID, respCh)
	defer s.pendingRequests.Delete(permission.ID)

	s.Publish(pubsub.CreatedEvent, permission)

	select {
	case resp := <-respCh:
		return resp
	case <-time.After(60 * time.Second):
		return false
	}
}

func (s *permissionService) RequestBatch(opts []CreatePermissionRequest) []bool {
	results := make([]bool, len(opts))
	for i, o := range opts {
		decision := s.Check(o)
		if decision == DecisionAllow {
			results[i] = true
			continue
		}
		if decision == DecisionDeny {
			results[i] = false
			continue
		}
		// For Ask, we still need to prompt — do sequentially for now
		// to avoid overlapping permission dialogs.
		results[i] = s.Request(o)
	}
	return results
}

func (s *permissionService) AutoApproveSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !slices.Contains(s.autoApproveSessions, sessionID) {
		s.autoApproveSessions = append(s.autoApproveSessions, sessionID)
	}
}

func (s *permissionService) DenySession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !slices.Contains(s.deniedSessions, sessionID) {
		s.deniedSessions = append(s.deniedSessions, sessionID)
	}
}

func NewPermissionService() Service {
	return &permissionService{
		Broker:             pubsub.NewBroker[PermissionRequest](),
		sessionPermissions: make([]PermissionRequest, 0),
	}
}
