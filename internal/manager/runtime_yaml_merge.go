package manager

import (
	"bytes"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

// mergeBaseRuntimeSections 让 base 中声明的运行配置覆盖订阅中的同字段。
func mergeBaseRuntimeSections(baseContent, subscriptionContent string, keys ...string) (string, []string, error) {
	baseRoot, err := parseYAMLMapping(baseContent)
	if err != nil {
		return "", nil, fmt.Errorf("解析基础配置失败：%w", err)
	}

	subscriptionRoot, err := parseYAMLMapping(subscriptionContent)
	if err != nil {
		return "", nil, fmt.Errorf("解析订阅配置失败：%w", err)
	}

	merged := yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	overrides := make([]string, 0, len(keys))

	for _, key := range keys {
		baseKey, baseValue, ok := mappingValue(baseRoot, key)
		if !ok {
			continue
		}

		value := cloneYAMLNode(baseValue)
		if _, subscriptionValue, found := mappingValue(subscriptionRoot, key); found {
			value = mergeYAMLNode(subscriptionValue, baseValue)
		}

		merged.Content = append(merged.Content, cloneYAMLNode(baseKey), value)
		overrides = append(overrides, key)
	}

	if len(merged.Content) == 0 {
		return "", nil, nil
	}

	content, err := encodeYAMLMapping(&merged)
	if err != nil {
		return "", nil, fmt.Errorf("生成基础覆盖配置失败：%w", err)
	}

	return strings.TrimSpace(content), overrides, nil
}

func parseYAMLMapping(content string) (*yaml.Node, error) {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		return nil, err
	}

	if document.Kind != yaml.DocumentNode || len(document.Content) == 0 {
		return nil, fmt.Errorf("顶层配置不是 YAML 文档")
	}

	if document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("顶层配置不是映射")
	}

	return document.Content[0], nil
}

func mappingValue(mapping *yaml.Node, key string) (*yaml.Node, *yaml.Node, bool) {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil, nil, false
	}

	for index := 0; index+1 < len(mapping.Content); index += 2 {
		keyNode := mapping.Content[index]
		if keyNode.Value == key {
			return keyNode, mapping.Content[index+1], true
		}
	}

	return nil, nil, false
}

func mergeYAMLNode(current, override *yaml.Node) *yaml.Node {
	if current == nil {
		return cloneYAMLNode(override)
	}

	if override == nil {
		return cloneYAMLNode(current)
	}

	if current.Kind != yaml.MappingNode || override.Kind != yaml.MappingNode {
		return cloneYAMLNode(override)
	}

	merged := cloneYAMLNode(current)
	for index := 0; index+1 < len(override.Content); index += 2 {
		overrideKey := override.Content[index]
		overrideValue := override.Content[index+1]

		_, mergedValue, found := mappingValue(merged, overrideKey.Value)
		if !found {
			merged.Content = append(merged.Content, cloneYAMLNode(overrideKey), cloneYAMLNode(overrideValue))
			continue
		}

		for mergedIndex := 0; mergedIndex+1 < len(merged.Content); mergedIndex += 2 {
			if merged.Content[mergedIndex].Value != overrideKey.Value {
				continue
			}

			merged.Content[mergedIndex] = cloneYAMLNode(overrideKey)
			merged.Content[mergedIndex+1] = mergeYAMLNode(mergedValue, overrideValue)
			break
		}
	}

	return merged
}

func cloneYAMLNode(source *yaml.Node) *yaml.Node {
	if source == nil {
		return nil
	}

	clone := *source
	if len(source.Content) == 0 {
		return &clone
	}

	clone.Content = make([]*yaml.Node, 0, len(source.Content))
	for _, child := range source.Content {
		clone.Content = append(clone.Content, cloneYAMLNode(child))
	}

	return &clone
}

func encodeYAMLMapping(mapping *yaml.Node) (string, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)

	if err := encoder.Encode(mapping); err != nil {
		return "", err
	}

	if err := encoder.Close(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
