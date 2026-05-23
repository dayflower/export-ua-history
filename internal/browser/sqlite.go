package browser

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

func SQLiteReadOnlyURI(path string) string {
	uri := &url.URL{Scheme: "file", Path: path}
	query := uri.Query()
	query.Set("mode", "ro")
	uri.RawQuery = query.Encode()
	return uri.String()
}

func EscapeSQLiteString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func CopyFileIfExists(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return FormatAccessError(src, "failed to copy Chrome history snapshot source", err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return FormatAccessError(src, "failed to stat Chrome history snapshot source", err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("failed to create snapshot file %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy snapshot file %s: %w", src, err)
	}

	return nil
}
