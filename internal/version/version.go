package version

var (
	Value     = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

func Current() Info {
	return Info{
		Version:   Value,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}
