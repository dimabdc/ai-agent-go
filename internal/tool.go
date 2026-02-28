package internal

import (
	"context"
	"fmt"
)

var UnknownToolsHandler = func(ctx context.Context, name, input string) (string, error) {
	return fmt.Sprintf(
		"unknown tool: %s; you made it up, try again with the correct tool name",
		name,
	), nil
}
