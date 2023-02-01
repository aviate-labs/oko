package github

// Example: https://api.github.com/repos/internet-computer/testing.mo/releases
type Release struct {
	TagName string `json:"tag_name"`
}
