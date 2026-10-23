package body

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// TreeNode represents a Project or Endpoint in the Explorer Tree.
type TreeNode struct {
	ID          string // Project name or Endpoint Path
	IsProject   bool
	FileName    string // The spec filename
	IsExpanded  bool
	HasChaos    bool
	HasOverride bool
	Children    []*TreeNode
}

// ProjectTreeModel handles the custom file explorer tree.
type ProjectTreeModel struct {
	tuiCtx        types.TUIContext
	Nodes         []*TreeNode
	ActiveIndex   int
	VisibleNodes  []*TreeNode // Flattened list for rendering and navigation
	IsFocused     bool
	Size          tea.WindowSizeMsg
	ScrollYOffset int
	SelectedID    string
}

func NewProjectTreeModel(ctx types.TUIContext) *ProjectTreeModel {
	m := &ProjectTreeModel{
		tuiCtx: ctx,
		Nodes:  make([]*TreeNode, 0),
	}
	m.RefreshFromCore()
	return m
}

func (m *ProjectTreeModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

// RefreshFromCore reads the projects from the backend core.
func (m *ProjectTreeModel) RefreshFromCore() {
	if m.tuiCtx.Core == nil {
		return
	}

	tree, err := m.tuiCtx.Core.GetProjectTree(context.Background())
	if err != nil {
		return
	}

	// Preserve expansion state
	expandedState := make(map[string]bool)
	for _, n := range m.Nodes {
		expandedState[n.ID] = n.IsExpanded
	}

	var newNodes []*TreeNode
	for _, proj := range tree {
		projNode := &TreeNode{
			ID:         proj.Name,
			IsProject:  true,
			FileName:   proj.SpecFilename,
			IsExpanded: expandedState[proj.Name],
			Children:   make([]*TreeNode, 0, len(proj.Endpoints)),
		}

		for _, ep := range proj.Endpoints {
			projNode.Children = append(projNode.Children, &TreeNode{
				ID:          ep.Path,
				IsProject:   false,
				HasOverride: ep.HasOverride,
				HasChaos:    ep.HasChaos,
			})
		}
		newNodes = append(newNodes, projNode)
	}

	m.Nodes = newNodes
	m.rebuildVisibleNodes()
}

// rebuildVisibleNodes flattens the tree into a 1D slice based on expansion state.
func (m *ProjectTreeModel) rebuildVisibleNodes() {
	var visible []*TreeNode
	for _, proj := range m.Nodes {
		visible = append(visible, proj)
		if proj.IsExpanded {
			visible = append(visible, proj.Children...)
		}
	}
	m.VisibleNodes = visible
	
	// Restore ActiveIndex based on SelectedID if possible
	found := false
	if m.SelectedID != "" {
		for i, node := range m.VisibleNodes {
			if node.ID == m.SelectedID {
				m.ActiveIndex = i
				found = true
				break
			}
		}
	}
	
	if !found {
		// Clamp active index
		if m.ActiveIndex >= len(m.VisibleNodes) {
			m.ActiveIndex = len(m.VisibleNodes) - 1
		}
		if m.ActiveIndex < 0 {
			m.ActiveIndex = 0
		}
	}
}

func (m *ProjectTreeModel) SetFocus(focus bool) {
	m.IsFocused = focus
}
