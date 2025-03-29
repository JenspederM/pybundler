package types

type Project struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

type PyProject struct {
	Project Project `toml:"project"`
}
