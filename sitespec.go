package main

import (
	"fmt"
	"github.com/otiai10/copy"
	"github.com/tealeg/xlsx"
	"os"
	"path/filepath"
	"strings"
)

type sitespec struct {
	dir           string
	locales       map[string]struct{}
	defaultLocale string
	pages         []*pagespec
}

func (s *sitespec) clear() {
	s.locales = make(map[string]struct{})
	s.defaultLocale = ""
	s.pages = make([]*pagespec, 0, 10)
}

func (s *sitespec) loadFromDir() (err error) {
	s.clear()
	if fi, statErr := os.Stat(s.dir); statErr != nil {
		return statErr
	} else if !fi.IsDir() {
		return fmt.Errorf("%q is not a directory", s.dir)
	}
	// TODO sub-directories
	matches, err := filepath.Glob(filepath.Join(s.dir, "spec", "*.xlsx"))
	if err != nil {
		return
	}
	for _, fileName := range matches {
		if strings.HasPrefix(filepath.Base(fileName), "~") {
			continue // ignore auto save files
		}
		file, openErr := xlsx.OpenFile(fileName)
		if openErr != nil {
			return fmt.Errorf("%s: %s", openErr, fileName)
		}
		loadErr := s.loadFromExcelFile(file)
		if loadErr != nil {
			return fmt.Errorf("%s: %s", loadErr, fileName)
		}
	}
	return nil
}

func (s *sitespec) loadFromExcelFile(file *xlsx.File) (err error) {
	for _, sheet := range file.Sheets {
		var p pagespec
		if err = p.loadFromSheet(sheet); err != nil {
			return
		}
		s.pages = append(s.pages, &p)
	}
	return nil
}

func (s *sitespec) execute() error {
	fi, err := os.Stat(s.dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%q is not a directory", s.dir)
	}
	outputDir := filepath.Join(s.dir, "www")
	if err := os.RemoveAll(outputDir); err != nil {
		return err
	}
	staticDir := filepath.Join(s.dir, "spec", "static")
	if hasStaticDir, err := dirExists(staticDir); err != nil {
		return err
	} else if hasStaticDir {
		if err := copy.Copy(staticDir, outputDir); err != nil {
			return err
		}
	} else {
		if err := os.MkdirAll(outputDir, fi.Mode()); err != nil {
			return err
		}
	}

	for _, ps := range s.pages {
		if err := ps.emit(outputDir, fi.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func dirExists(dir string) (found bool, err error) {
	fi, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	found = fi.IsDir()
	return
}
