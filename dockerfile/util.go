package dockerfile

import (
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"strings"
)

func SplitArg(node *parser.Node) (string, string) {
	if node.Next == nil {
		return "", ""
	}
	arg := node.Next.Value
	split := strings.SplitN(arg, "=", 2)

	// lalala you didn't see anything
	switch len(split) {
	case 0:
		return "", ""
	case 1:
		return split[0], ""
	default:
		return split[0], split[1]
	}
}

