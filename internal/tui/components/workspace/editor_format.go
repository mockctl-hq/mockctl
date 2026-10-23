package workspace

import (
	"bytes"
	"encoding/json"
	
	"gopkg.in/yaml.v3"
)

// FormatPayload takes a string of JSON or YAML and returns it formatted.
func FormatPayload(content string, isYaml bool, tabSize int) (string, error) {
	if content == "" {
		return "", nil
	}
	
	indent := ""
	for i := 0; i < tabSize; i++ {
		indent += " "
	}
	
	if isYaml {
		var node yaml.Node
		err := yaml.Unmarshal([]byte(content), &node)
		if err != nil {
			return content, err
		}
		
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(tabSize)
		err = encoder.Encode(&node)
		_ = encoder.Close()
		if err != nil {
			return content, err
		}
		
		return buf.String(), nil
	}
	
	// JSON
	var obj interface{}
	err := json.Unmarshal([]byte(content), &obj)
	if err != nil {
		return content, err
	}
	
	formatted, err := json.MarshalIndent(obj, "", indent)
	if err != nil {
		return content, err
	}
	
	return string(formatted), nil
}
