package app

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/dayflower/export-ua-history/internal/browser"
	"github.com/dayflower/export-ua-history/internal/browser/chrome"
	"github.com/dayflower/export-ua-history/internal/buildinfo"
	"github.com/dayflower/export-ua-history/internal/cli"
	"github.com/dayflower/export-ua-history/internal/output"
)

type Runner struct {
	LoadProfiles    func() ([]browser.Profile, error)
	ResolveProfile  func([]browser.Profile, string, string, string) (browser.Profile, error)
	HistoryDBPath   func(string) (string, error)
	ExportHistory   func(string, browser.ExportRange) ([]browser.HistoryEntry, error)
	WriteProfiles   func(io.Writer, []browser.Profile) error
	WriteReport     func(io.Writer, output.Report) error
	WriteReportFile func(string, output.Report) error
	FormatVersion   func(string, string) string
	Version         string
}

func NewRunner() Runner {
	return Runner{
		LoadProfiles:    chrome.LoadProfilesFromDefaultLocation,
		ResolveProfile:  browser.ResolveProfile,
		HistoryDBPath:   chrome.HistoryDBPath,
		ExportHistory:   chrome.ExportHistory,
		WriteProfiles:   output.WriteProfiles,
		WriteReport:     output.WriteReport,
		WriteReportFile: output.WriteReportFile,
		FormatVersion:   output.FormatVersion,
		Version:         buildinfo.Version,
	}
}

func (r Runner) Run(args []string, stdout, stderr io.Writer, now time.Time) error {
	result, err := cli.Parse(args, now)
	if err != nil {
		if errors.Is(err, cli.ErrHelpRequested) {
			_, writeErr := fmt.Fprint(stdout, cli.HelpText(args))
			return writeErr
		}
		return err
	}

	switch result.Command {
	case cli.CommandVersion:
		_, err := fmt.Fprintln(stdout, r.FormatVersion("export-ua-history", r.Version))
		return err
	case cli.CommandListProfiles:
		return r.runListProfiles(stdout)
	case cli.CommandExport:
		return r.runExport(result.Options, stdout)
	default:
		return fmt.Errorf("unsupported command: %q", result.Command)
	}
}

func (r Runner) runListProfiles(stdout io.Writer) error {
	profiles, err := r.LoadProfiles()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return errors.New("no Chrome profiles found")
	}
	return r.WriteProfiles(stdout, profiles)
}

func (r Runner) runExport(opts cli.Options, stdout io.Writer) error {
	profiles, err := r.LoadProfiles()
	if err != nil {
		return err
	}

	profile, err := r.ResolveProfile(profiles, opts.ProfileName, opts.ProfilePath, "Default")
	if err != nil {
		return err
	}

	historyPath, err := r.HistoryDBPath(profile.Path)
	if err != nil {
		return err
	}

	entries, err := r.ExportHistory(historyPath, opts.Range)
	if err != nil {
		return err
	}

	report := output.NewReport("chrome", opts.Range, entries)
	if opts.OutputPath == "" {
		return r.WriteReport(stdout, report)
	}
	return r.WriteReportFile(opts.OutputPath, report)
}
