package version

import "runtime/debug"

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func Go() string {
	return runtimeVersion()
}

func Module() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		return bi.Main.Version
	}
	return "unknown"
}

func runtimeVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		return bi.GoVersion
	}
	return "unknown"
}

func Info() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit":     Commit,
		"build_date": BuildDate,
		"go":         Go(),
		"module":     Module(),
	}
}

func String() string {
	return Version + " (" + Commit + ", " + BuildDate + ")"
}

func IsRelease() bool {
	return Version != "" && Version != "dev"
}
