package main

import "testing"

func Test_sitespec_loadFromDir(t *testing.T) {
	var s sitespec
	s.dir = "testdata/site1"
	loadErr := s.loadFromDir()
	if loadErr != nil {
		t.Fatalf("load error: %s", loadErr)
		return
	}
	if len(s.pages) != 3 {
		t.Errorf("expected 2 pages but found %d", len(s.pages))
	}
}
