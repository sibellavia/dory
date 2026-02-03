package doryfile

import (
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// MarshalYAML ensures body uses literal block style for readability.
func (e Entry) MarshalYAML() (interface{}, error) {
	// Build node manually to control styles.
	node := &yaml.Node{Kind: yaml.MappingNode}

	addField := func(key, value string) {
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key},
			&yaml.Node{Kind: yaml.ScalarNode, Value: value},
		)
	}

	addField("id", e.ID)
	addField("type", e.Type)
	if e.Topic != "" {
		addField("topic", e.Topic)
	}
	if e.Domain != "" {
		addField("domain", e.Domain)
	}
	if e.Severity != "" {
		addField("severity", e.Severity)
	}
	addField("oneliner", e.Oneliner)
	addField("created", e.Created.Format(time.RFC3339Nano))

	if len(e.Refs) > 0 {
		refsNode := &yaml.Node{Kind: yaml.SequenceNode}
		for _, ref := range e.Refs {
			refsNode.Content = append(refsNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: ref})
		}
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "refs"},
			refsNode,
		)
	}

	if e.Body != "" {
		// Strip trailing spaces from lines (YAML literal blocks can't preserve them).
		body := e.Body
		lines := strings.Split(body, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimRight(line, " \t")
		}
		body = strings.Join(lines, "\n")

		bodyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: body}
		if strings.Contains(body, "\n") {
			bodyNode.Style = yaml.LiteralStyle
		}
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "body"},
			bodyNode,
		)
	}

	return node, nil
}
