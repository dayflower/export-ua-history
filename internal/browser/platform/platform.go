package platform

import "os"

type Directories struct {
	HomeDir      string
	ConfigHome   string
	LocalAppData string
}

type EnvLookupFunc func(string) string
type HomeDirFunc func() (string, error)

func DefaultDirectories() (Directories, error) {
	return currentDirectories(os.Getenv, os.UserHomeDir)
}
