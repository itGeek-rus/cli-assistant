package version

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func String() string {
	return Version + " (" + Commit + ", " + BuildDate + ")"
}

func IsRelease() bool {
	return Version != "" && Version != "dev"
}
