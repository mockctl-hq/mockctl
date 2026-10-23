package inspector

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type Node struct {
	Key      string
	Value    string
	IsObject bool
	IsArray  bool
	Expanded bool
	Children []*Node
	Depth    int
}

type StateInspectorModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	RootNodes   []*Node
	FlatVisible []*Node
	Cursor      int
	
	Error string
}

var _ types.Component = (*StateInspectorModel)(nil)

func NewStateInspectorModel(ctx types.TUIContext) *StateInspectorModel {
	return &StateInspectorModel{
		tuiCtx: ctx,
	}
}

func (m *StateInspectorModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *StateInspectorModel) RefreshState(projectName string) {
	if projectName == "" {
		m.Error = "No project selected."
		return
	}
	
	store := m.tuiCtx.Core.StateStore()
	if store == nil {
		m.Error = "StateStore not found for project."
		return
	}
	
	memStore, ok := store.(interface {
		GetSnapshot() map[string]map[string]map[string]any
	})
	if !ok {
		m.Error = "StateStore does not support GetSnapshot"
		return
	}
	
	snapshot := memStore.GetSnapshot()
	
	m.RootNodes = []*Node{}
	
	// Sort routes for stable rendering
	var routes []string
	for route := range snapshot {
		routes = append(routes, route)
	}
	sort.Strings(routes)
	
	for _, route := range routes {
		data := snapshot[route]
		b, _ := json.Marshal(data) // simple marshal
		
		var parsed interface{}
		json.Unmarshal(b, &parsed)
		
		node := buildNode(route, parsed, 0)
		m.RootNodes = append(m.RootNodes, node)
	}
	
	m.Error = ""
	m.rebuildFlatList()
}

func buildNode(key string, val interface{}, depth int) *Node {
	// PKS-008: MaxDepth: 5 limit
	if depth >= 5 {
		return &Node{Key: key, Value: "{...}", Depth: depth}
	}
	
	n := &Node{Key: key, Depth: depth}
	
	switch v := val.(type) {
	case map[string]interface{}:
		n.IsObject = true
		n.Expanded = depth < 1 // auto-expand root
		var keys []string
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			n.Children = append(n.Children, buildNode(k, v[k], depth+1))
		}
	case []interface{}:
		n.IsArray = true
		n.Expanded = depth < 1
		for i, item := range v {
			n.Children = append(n.Children, buildNode(fmt.Sprintf("[%d]", i), item, depth+1))
		}
	default:
		n.Value = fmt.Sprintf("%v", v)
		// string formatting for chroma later
		if _, ok := v.(string); ok {
			n.Value = fmt.Sprintf(`"%v"`, strings.ReplaceAll(n.Value, "\n", "\\n"))
		}
	}
	return n
}

func (m *StateInspectorModel) rebuildFlatList() {
	m.FlatVisible = []*Node{}
	for _, root := range m.RootNodes {
		m.flatten(root)
	}
	
	if m.Cursor >= len(m.FlatVisible) {
		m.Cursor = len(m.FlatVisible) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
}

func (m *StateInspectorModel) flatten(n *Node) {
	m.FlatVisible = append(m.FlatVisible, n)
	if n.Expanded && (n.IsObject || n.IsArray) {
		for _, child := range n.Children {
			m.flatten(child)
		}
	}
}

func (m *StateInspectorModel) SetFocus(focus bool) {}
