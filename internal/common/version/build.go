package version

var (
	commitHash string
	version    string
	date       string
)

// Info holds build information
type Info struct {
	BuildDate string `json:"build_date"`
	BuildHash string `json:"build_hash"`
	Version   string `json:"version"`
}

// Get creates and initialized Info object
func Get() Info {
	return Info{
		BuildDate: date,
		BuildHash: commitHash,
		Version:   version,
	}
}
