package http

import "strings"

type RequestDraft struct {
	Method  string
	Path    string
	Body    string
	Headers string
}

// SaveCurrentDraft caches user inputs for the active endpoint.
func (m *HttpTesterModel) SaveCurrentDraft() {
	if m.activeDraftKey != "" {
		m.drafts[m.activeDraftKey] = RequestDraft{
			Method:  m.Input.Method,
			Path:    m.Input.PathInput.Value(),
			Body:    m.Body.Input.Value(),
			Headers: m.Headers.Input.Value(),
		}
	}
}

// LoadDraft restores a cached draft or populates default values for the target endpoint.
func (m *HttpTesterModel) LoadDraft(projectName, endpointPath string) {
	m.SaveCurrentDraft()

	targetKey := projectName + ":" + endpointPath
	m.activeDraftKey = targetKey

	if draft, exists := m.drafts[targetKey]; exists {
		m.Input.Method = draft.Method
		m.Input.PathInput.SetValue(draft.Path)
		m.Body.Input.SetValue(draft.Body)
		m.Headers.Input.SetValue(draft.Headers)
		return
	}

	path := endpointPath
	parts := strings.SplitN(path, " ", 2)
	if len(parts) == 2 {
		m.Input.Method = strings.ToUpper(parts[0])
		path = strings.TrimSpace(parts[1])
	} else {
		m.Input.Method = "GET"
	}

	fullPath := "/" + projectName
	if !strings.HasPrefix(path, "/") {
		fullPath += "/" + path
	} else {
		fullPath += path
	}

	m.Input.PathInput.SetValue(fullPath)
	m.Body.Input.SetValue("")
	m.Headers.Input.SetValue("")
}
