package github

import "fmt"

type ReleasesNotFoundErrors struct {
	URL string
}

func NewReleasesNotFoundErrors(url string) *ReleasesNotFoundErrors {
	return &ReleasesNotFoundErrors{
		URL: url,
	}
}

func (e ReleasesNotFoundErrors) Error() string {
	return fmt.Sprintf("no releases found for %q", e.URL)
}
