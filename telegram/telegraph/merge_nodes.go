package telegraph

import (
	"TLExtractor/telegram/telegraph/types"
	"strings"
)

func mergeNodes(nodes []types.RawNode) []types.Node {
	var rootNodes []types.Node
	nesting := nodes[0].Nesting
	for pos, node := range nodes {
		if node.Nesting == nesting {
			if node.Tag == "text" {
				if len(strings.TrimSpace(node.Attrs["value"])) == 0 {
					continue
				}
			}
			newNode := types.Node{
				Tag:   node.Tag,
				Attrs: node.Attrs,
			}
			if len(nodes) > pos+1 && nodes[pos+1:][0].Nesting > nesting {
				newNode.Children = mergeNodes(nodes[pos+1:])
			}
			rootNodes = append(rootNodes, newNode)
		} else if node.Nesting < nesting {
			break
		}
	}
	return rootNodes
}
