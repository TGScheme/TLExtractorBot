package ios

import "github.com/Laky-64/goswift/demangling"

func deepTextMatch(node *demangling.Node, text string, kindOrder ...demangling.NodeKind) bool {
	for _, child := range node.Children {
		if len(kindOrder) > 0 && child.Kind == kindOrder[0] {
			kindOrder = kindOrder[1:]
		}
		if len(kindOrder) == 0 && child.Text == text {
			return true
		}
		if deepTextMatch(child, text, kindOrder...) {
			return true
		}
	}
	return false
}
