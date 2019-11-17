package main

import (
	"github.com/tealeg/xlsx"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func Test_pagespec_loadFromSheet(t *testing.T) {
	wb, err := xlsx.OpenFile("testdata/site1/spec/site1.xlsx")
	if err != nil {
		t.Fatal(err)
		return
	}
	type args struct {
		sheet *xlsx.Sheet
	}
	tests := []struct {
		sheet        *xlsx.Sheet
		expectedYaml string
		wantErr      bool
	}{
		{
			sheet:        wb.Sheets[0],
			expectedYaml: expectedIndexPageSpec,
			wantErr:      false,
		},
		{
			sheet:        wb.Sheets[1],
			expectedYaml: expectedImprintPageSpec,
			wantErr:      false,
		},
	}
	for i, tt := range tests {
		t.Run(tt.sheet.Name, func(t *testing.T) {
			p := &pagespec{}
			if err := p.loadFromSheet(tt.sheet); (err != nil) != tt.wantErr {
				t.Errorf("loadFromSheet() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			var sb strings.Builder
			enc := yaml.NewEncoder(&sb)
			if err := enc.Encode(p.root); err != nil {
				t.Fatal(err)
			}
			psYaml := sb.String()
			if psYaml != tt.expectedYaml {
				t.Errorf("case %d: pagespec not as expected, found:\n%s", i, psYaml)
			}
		})
	}
}

const expectedIndexPageSpec = `key: html
isattr: false
ismacro: false
valuebylocale: {}
sheetrow: 0
depth: 0
children:
  - key: head
    isattr: false
    ismacro: false
    valuebylocale: {}
    sheetrow: 3
    depth: 1
    children:
      - key: title
        isattr: false
        ismacro: false
        valuebylocale:
            de: Testseite
            en: Test Page
        sheetrow: 4
        depth: 2
        children: []
      - key: link
        isattr: false
        ismacro: false
        valuebylocale: {}
        sheetrow: 5
        depth: 2
        children:
          - key: rel
            isattr: true
            ismacro: false
            valuebylocale:
                en: stylesheet
            sheetrow: 6
            depth: 3
            children: []
          - key: href
            isattr: true
            ismacro: false
            valuebylocale:
                en: /style.css
            sheetrow: 7
            depth: 3
            children: []
  - key: body
    isattr: false
    ismacro: false
    valuebylocale: {}
    sheetrow: 8
    depth: 1
    children:
      - key: h1
        isattr: false
        ismacro: false
        valuebylocale:
            de: Willkommen
            en: Welcome
        sheetrow: 9
        depth: 2
        children: []
      - key: p
        isattr: false
        ismacro: false
        valuebylocale:
            de: Dies ist etwas Text.
            en: This is some text.
        sheetrow: 10
        depth: 2
        children: []
      - key: footer
        isattr: false
        ismacro: false
        valuebylocale:
            de: <a href="/imprint">Impressum</a>
            en: <a href="/imprint">Imprint</a>
        sheetrow: 11
        depth: 2
        children: []
`

const expectedImprintPageSpec = `key: html
isattr: false
ismacro: false
valuebylocale: {}
sheetrow: 0
depth: 0
children:
  - key: head
    isattr: false
    ismacro: false
    valuebylocale: {}
    sheetrow: 3
    depth: 1
    children:
      - key: title
        isattr: false
        ismacro: false
        valuebylocale: {}
        sheetrow: 4
        depth: 2
        children: []
      - key: link
        isattr: false
        ismacro: false
        valuebylocale: {}
        sheetrow: 5
        depth: 2
        children:
          - key: rel
            isattr: true
            ismacro: false
            valuebylocale:
                en: stylesheet
            sheetrow: 6
            depth: 3
            children: []
          - key: href
            isattr: true
            ismacro: false
            valuebylocale:
                en: /style.css
            sheetrow: 7
            depth: 3
            children: []
  - key: body
    isattr: false
    ismacro: false
    valuebylocale: {}
    sheetrow: 8
    depth: 1
    children:
      - key: h1
        isattr: false
        ismacro: false
        valuebylocale:
            de: Impressum
            en: Imprint
        sheetrow: 9
        depth: 2
        children: []
`

func Test_pagespec_writeLocale(t *testing.T) {
	wb, err := xlsx.OpenFile("testdata/site1/spec/site1.xlsx")
	if err != nil {
		t.Fatal(err)
		return
	}
	type args struct {
		sheet *xlsx.Sheet
	}
	tests := []struct {
		sheet        *xlsx.Sheet
		locale       string
		expectedHTML string
		wantErr      bool
	}{
		{
			sheet:        wb.Sheets[0],
			locale:       "en",
			expectedHTML: indexHTMLen,
			wantErr:      false,
		},
		{
			sheet:        wb.Sheets[0],
			locale:       "de",
			expectedHTML: indexHTMLde,
			wantErr:      false,
		},
		{
			sheet:        wb.Sheets[1],
			locale:       "en",
			expectedHTML: imprintHTMLen,
			wantErr:      false,
		},
		{
			sheet:        wb.Sheets[1],
			locale:       "de",
			expectedHTML: imprintHTMLde,
			wantErr:      false,
		},
	}
	for i, tt := range tests {
		t.Run(tt.sheet.Name, func(t *testing.T) {
			p := &pagespec{}
			if err := p.loadFromSheet(tt.sheet); err != nil {
				t.Errorf("loadFromSheet() error: %s", err)
				return
			}
			var sb strings.Builder
			if err := p.writeLocale(&sb, tt.locale); (err != nil) != tt.wantErr {
				t.Errorf("writeLocale() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			outHTML := sb.String()
			if outHTML != tt.expectedHTML {
				t.Errorf("case %d: expected HTML not as expected, found:\n%v", i, outHTML)
			}
		})
	}
}

const indexHTMLen = `<!DOCTYPE html><html lang="en"><head><title>Test Page</title><link rel="stylesheet" href="/style.css"/></head><body><h1>Welcome</h1><p>This is some text.</p><footer><a href="/imprint">Imprint</a></footer></body></html>`
const indexHTMLde = `<!DOCTYPE html><html lang="de"><head><title>Testseite</title><link rel="stylesheet" href="/style.css"/></head><body><h1>Willkommen</h1><p>Dies ist etwas Text.</p><footer><a href="/imprint">Impressum</a></footer></body></html>`
const imprintHTMLen = `<!DOCTYPE html><html lang="en"><head><title></title><link rel="stylesheet" href="/style.css"/></head><body><h1>Imprint</h1></body></html>`
const imprintHTMLde = `<!DOCTYPE html><html lang="de"><head><title></title><link rel="stylesheet" href="/style.css"/></head><body><h1>Impressum</h1></body></html>`