package svc

import "github.com/goccy/go-yaml"

func ParseConfigYAML(file []byte) (*GiTrimConfig, error) {
	result := &GiTrimConfig{}

	if err := yaml.Unmarshal(file, result); err != nil {
		return nil, err
	}

	return result, nil
}
