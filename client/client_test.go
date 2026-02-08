package client

import "testing"

func TestFetchIssues(t *testing.T) {
	issues, err := FetchIssues()
	if err != nil {
		t.Fatalf("expected no error, fot %v", err)
	}

	if len(issues) == 0 {
		t.Fatalf("expected at least one issue")
	}
}
