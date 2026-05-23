package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	CommandExport       = "export"
	CommandHelp         = "help"
	CommandListProfiles = "list-profiles"
	CommandVersion      = "version"
)

var ErrHelpRequested = errors.New("flag: help requested")

var supportedCommands = []string{CommandExport, CommandHelp, CommandListProfiles, CommandVersion}

var unsupportedOptions = map[string]struct{}{
	"--all-browsers": {},
	"--db-path":      {},
	"--utc":          {},
	"--tz":           {},
}

type ParseResult struct {
	Command string
	Options Options
}

type Options struct {
	Browser     string
	ProfileName string
	ProfilePath string
	OutputPath  string
	Range       ResolvedRange
}

type ResolvedRange struct {
	StartLocal               time.Time
	EndLocalExclusive        time.Time
	DisplayEndLocalInclusive time.Time
}

func Parse(args []string, now time.Time) (ParseResult, error) {
	command, commandArgs, err := resolveCommand(args)
	if err != nil {
		return ParseResult{}, err
	}

	switch command {
	case CommandExport:
		opts, err := parseExportOptions(commandArgs, now)
		if err != nil {
			return ParseResult{}, err
		}
		return ParseResult{Command: command, Options: opts}, nil
	case CommandHelp:
		if len(commandArgs) > 1 {
			return ParseResult{}, errors.New("help accepts at most one command name")
		}
		if len(commandArgs) == 1 && !slices.Contains(supportedCommands, commandArgs[0]) {
			return ParseResult{}, fmt.Errorf("unknown help topic %q", commandArgs[0])
		}
		return ParseResult{}, ErrHelpRequested
	case CommandListProfiles, CommandVersion:
		if len(commandArgs) == 1 && (commandArgs[0] == "--help" || commandArgs[0] == "-h") {
			return ParseResult{}, ErrHelpRequested
		}
		if len(commandArgs) > 0 {
			return ParseResult{}, fmt.Errorf("%s does not accept arguments", command)
		}
		return ParseResult{Command: command}, nil
	default:
		return ParseResult{}, fmt.Errorf("unsupported command: %q", command)
	}
}

func HelpText(args []string) string {
	command, _, err := resolveCommand(args)
	if err != nil {
		return rootHelp()
	}
	if command == CommandHelp {
		if len(args) >= 2 {
			return HelpText(args[1:])
		}
		return rootHelp()
	}
	switch command {
	case CommandExport:
		return exportHelp()
	case CommandListProfiles:
		return "Usage:\n  export-ua-history list-profiles\n\nList Chrome profiles.\n"
	case CommandVersion:
		return "Usage:\n  export-ua-history version\n\nPrint version information.\n"
	default:
		return rootHelp()
	}
}

func resolveCommand(args []string) (string, []string, error) {
	if len(args) == 0 {
		return CommandExport, nil, nil
	}
	first := args[0]
	if slices.Contains(supportedCommands, first) {
		return first, args[1:], nil
	}
	if strings.HasPrefix(first, "-") {
		return CommandExport, args, nil
	}
	return "", nil, fmt.Errorf("unknown command %q", first)
}

func parseExportOptions(args []string, now time.Time) (Options, error) {
	if hasHelp(args) {
		return Options{}, ErrHelpRequested
	}
	if err := rejectUnsupportedOptions(args); err != nil {
		return Options{}, err
	}

	fs := flag.NewFlagSet(CommandExport, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts Options
	var (
		date      string
		startDate string
		endDate   string
		startTime string
		endTime   string
		hour      int
		timeSet   bool
	)

	fs.StringVar(&opts.Browser, "browser", "chrome", "")
	fs.StringVar(&opts.ProfileName, "profile", "", "")
	fs.StringVar(&opts.ProfilePath, "profile-path", "", "")
	fs.StringVar(&date, "date", "", "")
	fs.StringVar(&startDate, "start-date", "", "")
	fs.StringVar(&endDate, "end-date", "", "")
	fs.StringVar(&startTime, "start-time", "", "")
	fs.StringVar(&endTime, "end-time", "", "")
	fs.Func("time", "", func(value string) error {
		timeSet = true
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid --time value %q: must be an integer from 0 to 23", value)
		}
		hour = parsed
		return nil
	})
	fs.StringVar(&opts.OutputPath, "output", "", "")
	fs.StringVar(&opts.OutputPath, "o", "", "")

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	if len(fs.Args()) > 0 {
		return Options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	if opts.Browser != "chrome" {
		return Options{}, fmt.Errorf("unsupported browser %q: only \"chrome\" is supported", opts.Browser)
	}
	if opts.ProfileName != "" && opts.ProfilePath != "" {
		return Options{}, errors.New("--profile and --profile-path are mutually exclusive")
	}

	rng, err := resolveRange(now, date, startDate, endDate, startTime, endTime, hour, timeSet)
	if err != nil {
		return Options{}, err
	}
	opts.Range = rng
	return opts, nil
}

func resolveRange(now time.Time, date, startDate, endDate, startTime, endTime string, hour int, timeSet bool) (ResolvedRange, error) {
	location := now.Location()
	if date != "" && (startDate != "" || endDate != "") {
		return ResolvedRange{}, errors.New("--date is mutually exclusive with --start-date and --end-date")
	}
	if (startDate == "") != (endDate == "") {
		return ResolvedRange{}, errors.New("--start-date and --end-date must be supplied together")
	}
	if timeSet && (startTime != "" || endTime != "") {
		return ResolvedRange{}, errors.New("--time is mutually exclusive with --start-time and --end-time")
	}

	if date == "" && startDate == "" && endDate == "" {
		date = now.In(location).Format("2006-01-02")
	}

	if startDate != "" {
		if timeSet || startTime != "" || endTime != "" {
			return ResolvedRange{}, errors.New("--time, --start-time, and --end-time may only be used with --date")
		}
		startDay, err := parseDate(startDate, location)
		if err != nil {
			return ResolvedRange{}, fmt.Errorf("invalid --start-date: %w", err)
		}
		endDay, err := parseDate(endDate, location)
		if err != nil {
			return ResolvedRange{}, fmt.Errorf("invalid --end-date: %w", err)
		}
		if endDay.Before(startDay) {
			return ResolvedRange{}, errors.New("--end-date must be on or after --start-date")
		}
		start := startDay
		endExclusive := endDay.AddDate(0, 0, 1)
		return ResolvedRange{
			StartLocal:               start,
			EndLocalExclusive:        endExclusive,
			DisplayEndLocalInclusive: endExclusive.Add(-time.Second),
		}, nil
	}

	day, err := parseDate(date, location)
	if err != nil {
		return ResolvedRange{}, fmt.Errorf("invalid --date: %w", err)
	}

	start := day
	endExclusive := day.AddDate(0, 0, 1)

	if timeSet {
		if hour < 0 || hour > 23 {
			return ResolvedRange{}, fmt.Errorf("invalid --time value %d: must be an integer from 0 to 23", hour)
		}
		start = day.Add(time.Duration(hour) * time.Hour)
		endExclusive = start.Add(time.Hour)
	} else if startTime != "" || endTime != "" {
		if startTime == "" || endTime == "" {
			return ResolvedRange{}, errors.New("--start-time and --end-time must be supplied together")
		}
		startOffset, err := parseClock(startTime)
		if err != nil {
			return ResolvedRange{}, fmt.Errorf("invalid --start-time: %w", err)
		}
		endOffset, err := parseClock(endTime)
		if err != nil {
			return ResolvedRange{}, fmt.Errorf("invalid --end-time: %w", err)
		}
		if startOffset >= endOffset {
			return ResolvedRange{}, errors.New("--start-time must be strictly earlier than --end-time")
		}
		start = day.Add(startOffset)
		endExclusive = day.Add(endOffset)
	}

	return ResolvedRange{
		StartLocal:               start,
		EndLocalExclusive:        endExclusive,
		DisplayEndLocalInclusive: endExclusive.Add(-time.Second),
	}, nil
}

func parseDate(value string, location *time.Location) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", value, location)
	if err != nil {
		return time.Time{}, errors.New("must use YYYY-MM-DD")
	}
	return parsed, nil
}

func parseClock(value string) (time.Duration, error) {
	if len(value) != len("15:04") {
		return 0, errors.New("must use HH:MM in 24-hour format")
	}
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, errors.New("must use HH:MM in 24-hour format")
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, errors.New("must use HH:MM in 24-hour format")
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, errors.New("must use HH:MM in 24-hour format")
	}
	return time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute, nil
}

func rejectUnsupportedOptions(args []string) error {
	for _, arg := range args {
		name := arg
		if key, _, ok := strings.Cut(arg, "="); ok {
			name = key
		}
		if _, ok := unsupportedOptions[name]; ok {
			return fmt.Errorf("unsupported option %q", name)
		}
	}
	return nil
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func rootHelp() string {
	return `Usage:
  export-ua-history export [options]
  export-ua-history [options]
  export-ua-history list-profiles
  export-ua-history version
  export-ua-history help [command]

Commands:
  export         Export Chrome history as JSON.
  help           Show help for the root command or a subcommand.
  list-profiles  List Chrome profiles.
  version        Print version information.

The root command without a subcommand behaves the same as "export".
`
}

func exportHelp() string {
	return `Usage:
  export-ua-history export [options]
  export-ua-history [options]

Related commands:
  export-ua-history list-profiles
  export-ua-history version
  export-ua-history help [command]

Options:
  --browser <browser>
  --profile <profile name>
  --profile-path <profile path>
  --date <YYYY-MM-DD>
  --start-date <YYYY-MM-DD>
  --end-date <YYYY-MM-DD>
  --start-time <HH:MM>
  --end-time <HH:MM>
  --time <0-23>
  -o, --output <file>
  --help
`
}
