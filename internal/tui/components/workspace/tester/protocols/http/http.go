package http

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type HttpFocusEnum int

const (
	FocusInput HttpFocusEnum = iota
	FocusBody
	FocusHeaders
	FocusViewer
)

type HttpTesterModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	Input   *HttpInputModel
	Body    *HttpBodyModel
	Headers *HttpHeadersModel
	Viewer  *HttpViewerModel
	Client  *Client

	FocusRing  HttpFocusEnum
	IsFocused  bool
	ActiveTab  int // 0: Body, 1: Headers
	
	// Cancellation context for ghost request prevention
	cancelReq context.CancelFunc

	// In-memory request drafts cache
	drafts         map[string]RequestDraft
	activeDraftKey string
}

var _ types.Component = (*HttpTesterModel)(nil)

type ResponseMsg struct {
	Response *TestResponse
}

func NewHttpTesterModel(ctx types.TUIContext) *HttpTesterModel {
	client := NewClient()
	if ctx.Settings != nil {
		client = NewClientWithSettings(ctx.Settings.FollowRedirects, ctx.Settings.RequestTimeout)
	}

	return &HttpTesterModel{
		tuiCtx:    ctx,
		Input:     NewHttpInputModel(ctx),
		Body:      NewHttpBodyModel(ctx),
		Headers:   NewHttpHeadersModel(ctx),
		Viewer:    NewHttpViewerModel(ctx),
		Client:    client,
		FocusRing: FocusInput,
		ActiveTab: 0,
		drafts:    make(map[string]RequestDraft),
	}
}

func (m *HttpTesterModel) Init(ctx types.TUIContext) tea.Cmd {
	m.tuiCtx = ctx
	if ctx.Settings != nil && m.Client != nil {
		m.Client.ApplySettings(ctx.Settings.FollowRedirects, ctx.Settings.RequestTimeout)
	}
	return tea.Batch(
		m.Input.Init(ctx),
		m.Body.Init(ctx),
		m.Headers.Init(ctx),
		m.Viewer.Init(ctx),
	)
}

func (m *HttpTesterModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		leftWidth := msg.Width / 2
		rightWidth := msg.Width - leftWidth
		m.Input.Update(tea.WindowSizeMsg{Width: leftWidth, Height: 1})
		m.Body.Update(tea.WindowSizeMsg{Width: leftWidth, Height: msg.Height - 2})
		m.Headers.Update(tea.WindowSizeMsg{Width: leftWidth, Height: msg.Height - 2})
		m.Viewer.Update(tea.WindowSizeMsg{Width: rightWidth, Height: msg.Height})

	case types.EndpointSelectedMsg:
		m.LoadDraft(msg.ProjectName, msg.EndpointPath)
		return m, func() tea.Msg {
			return types.TesterDraftChangedMsg{
				Method:  m.Input.Method,
				Path:    m.Input.PathInput.Value(),
				Body:    m.Body.Input.Value(),
				Headers: m.Headers.Input.Value(),
			}
		}

	case ResponseMsg:
		m.Viewer.SetResponse(msg.Response)
		m.cancelReq = nil // clear
		return m, nil

	case spinner.TickMsg:
		if m.Viewer.IsLoading {
			_, cmd := m.Viewer.Update(msg)
			return m, cmd
		}
		return m, nil

	case types.SettingsChangedMsg:
		if msg.Config != nil {
			m.Client.ApplySettings(msg.Config.FollowRedirects, msg.Config.RequestTimeout)
		}
		return m, nil

	case tea.KeyMsg:
		if m.IsFocused {
			// Handle cancellation (Esc)
			if key.Matches(msg, types.GlobalKeys.Escape) {
				if m.cancelReq != nil {
					m.cancelReq()
					m.cancelReq = nil
					m.Viewer.SetResponse(&TestResponse{Error: fmt.Errorf("request cancelled by user")})
					return m, nil
				}
				// Otherwise, return focus to the URL input bar
				if m.FocusRing != FocusInput {
					m.FocusRing = FocusInput
					m.syncFocus()
				}
				return m, nil
			}

			// Handle execute
			isEnterInInput := m.FocusRing == FocusInput && msg.String() == "enter"
			if key.Matches(msg, types.GlobalKeys.ExecuteTest) || isEnterInInput {
				m.Input.PushHistory()
				m.Body.PushHistory()
				
				// PKS-008: context cancellation 
				ctx, cancel := context.WithCancel(context.Background())
				m.cancelReq = cancel
				
				inputVal := strings.TrimSpace(m.Input.PathInput.Value())
				
				var url string
				if strings.HasPrefix(inputVal, "http://") || strings.HasPrefix(inputVal, "https://") {
					url = inputVal
				} else {
					if !strings.HasPrefix(inputVal, "/") {
						inputVal = "/" + inputVal
					}
					// Clean multiple slashes
					for strings.Contains(inputVal, "//") {
						inputVal = strings.ReplaceAll(inputVal, "//", "/")
					}
					url = fmt.Sprintf("http://localhost:%d%s", m.tuiCtx.ActivePort, inputVal)
				}
				
				method := m.Input.Method
				body := m.Body.Input.Value()
				headers := m.Headers.Input.Value()
				
				spinCmd := m.Viewer.SetLoading(true)
				
				return m, tea.Batch(
					spinCmd,
					func() tea.Msg {
						resp := m.Client.ExecuteRequest(ctx, method, url, headers, body)
						return ResponseMsg{Response: resp}
					},
				)
			}

			// Local focus routing is now handled by RootModel via AdvanceFocus/ReverseFocus
		}
	}

	// Route down
	if m.IsFocused {
		if _, isWindowSize := msg.(tea.WindowSizeMsg); !isWindowSize {
			// Always route MouseMsg to Viewer for easy scrolling
			if _, isMouse := msg.(tea.MouseMsg); isMouse {
				_, cmd := m.Viewer.Update(msg)
				cmds = append(cmds, cmd)
			}
			
			switch m.FocusRing {
			case FocusInput:
				_, cmd := m.Input.Update(msg)
				cmds = append(cmds, cmd)
			case FocusBody:
				oldBody := m.Body.Input.Value()
				_, cmd := m.Body.Update(msg)
				cmds = append(cmds, cmd)
				if m.Body.Input.Value() != oldBody {
					cmds = append(cmds, func() tea.Msg {
						return types.TesterDraftChangedMsg{
							Method:  m.Input.Method,
							Path:    m.Input.PathInput.Value(),
							Body:    m.Body.Input.Value(),
							Headers: m.Headers.Input.Value(),
						}
					})
				}
			case FocusHeaders:
				oldHeaders := m.Headers.Input.Value()
				_, cmd := m.Headers.Update(msg)
				cmds = append(cmds, cmd)
				if m.Headers.Input.Value() != oldHeaders {
					cmds = append(cmds, func() tea.Msg {
						return types.TesterDraftChangedMsg{
							Method:  m.Input.Method,
							Path:    m.Input.PathInput.Value(),
							Body:    m.Body.Input.Value(),
							Headers: m.Headers.Input.Value(),
						}
					})
				}
			case FocusViewer:
				// If it's a mouse msg, we already routed it above
				if _, isMouse := msg.(tea.MouseMsg); !isMouse {
					_, cmd := m.Viewer.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *HttpTesterModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if !focus {
		m.SaveCurrentDraft()
		m.Input.SetFocus(false)
		m.Body.SetFocus(false)
		m.Headers.SetFocus(false)
		m.Viewer.SetFocus(false)
	} else {
		m.syncFocus()
	}
}

func (m *HttpTesterModel) syncFocus() {
	m.Input.SetFocus(m.FocusRing == FocusInput)
	m.Body.SetFocus(m.FocusRing == FocusBody)
	m.Headers.SetFocus(m.FocusRing == FocusHeaders)
	m.Viewer.SetFocus(m.FocusRing == FocusViewer)
}

func (m *HttpTesterModel) AdvanceFocus() bool {
	if m.FocusRing == FocusViewer {
		m.FocusRing = FocusInput // Reset for next time we enter
		m.syncFocus()
		return false // Tell parent to move on
	}
	m.FocusRing++
	m.syncFocus()
	return true // Stay here
}

func (m *HttpTesterModel) ReverseFocus() bool {
	if m.FocusRing == FocusInput {
		return false // Tell parent to move on
	}
	m.FocusRing--
	m.syncFocus()
	return true // Stay here
}
