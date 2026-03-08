package process

import (
	"fmt"
	"strings"
)

func splitCommandLine(command string) ([]string, error) {
	var args []string
	var current strings.Builder

	inSingle := false
	inDouble := false
	escape := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		args = append(args, current.String())
		current.Reset()
	}

	for _, r := range command {
		switch {
		case escape:
			current.WriteRune(r)
			escape = false
		case r == '\\' && !inSingle:
			escape = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case (r == ' ' || r == '\t') && !inSingle && !inDouble:
			flush()
		default:
			current.WriteRune(r)
		}
	}

	if escape || inSingle || inDouble {
		return nil, fmt.Errorf("invalid command line: %q", command)
	}
	flush()
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return args, nil
}
