package jdextract

import "strings"

func Parse(s string) {
}

type JobDescriptionNode struct {
	Content  string
	NodeType string
}

func buildProtoAST(s string) []JobDescriptionNode {
	lines := strings.Split(s, "\n")
	nodes := make([]JobDescriptionNode, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		node := JobDescriptionNode{
			Content:  line,
			NodeType: "unknown",
		}
		nodes = append(nodes, node)
	}

	return nodes
}
