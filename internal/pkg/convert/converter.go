package convert

import (
	"github.com/bytedance/sonic"
	"sigs.k8s.io/yaml"
)

// JSONToYAML → YAML
func JSONToYAML(jsonData []byte) ([]byte, error) {
	y, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return nil, err
	}
	return y, nil
}

// YAMLToJSON → JSON
func YAMLToJSON(yamlData []byte) ([]byte, error) {
	j, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// PrettyJSONToYAML JSON → YAML (pretty json as input)
func PrettyJSONToYAML(v any) ([]byte, error) {
	b, err := sonic.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return JSONToYAML(b)
}
