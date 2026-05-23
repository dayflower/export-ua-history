package browser

import "fmt"

func FormatAccessError(path, action string, err error) error {
	return fmt.Errorf("%s: %s: %w\nhint: the invoking terminal or parent process may not have sufficient permission to access the browser profile directory", action, path, err)
}
