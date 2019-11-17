package main

import (
	"errors"
	"fmt"
	"github.com/tealeg/xlsx"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type pagespec struct {
	pathByLocale map[string]string
	locales      []string
	root         *specNode
}

type specNode struct {
	Key           string
	IsAttr        bool
	IsMacro       bool
	ValueByLocale map[string]string
	SheetRow      int
	Depth         int
	Children      []*specNode
}

func (sn *specNode) bestValueByLocale(preferredLocale string, localeOrder []string) string {
	val, found := sn.ValueByLocale[preferredLocale]
	if found {
		return val
	}
	for _, locale := range localeOrder {
		val, found = sn.ValueByLocale[locale]
		if found {
			return val
		}
	}
	return ""
}

func (sn *specNode) toHTMLNode(locale string, localeOrder []string) (*html.Node, error) {
	elementType := atom.Lookup([]byte(sn.Key))
	if elementType == 0 {
		return nil, fmt.Errorf("row %d: unknown HTML element type %q", sn.SheetRow, sn.Key)
	}
	n := &html.Node{
		Type:     html.ElementNode,
		DataAtom: elementType,
		Data:     sn.Key,
	}
	elemBody := sn.bestValueByLocale(locale, localeOrder)
	if elemBody != "" {
		sr := strings.NewReader(elemBody)
		elemBodyNodes, err := html.ParseFragment(sr, n)
		if err != nil {
			return nil, fmt.Errorf("row %d, locale %s: node body %q is invalid HTML: %s", sn.SheetRow, locale, elemBody, err)
		}
		for _, bodyNode := range elemBodyNodes {
			n.AppendChild(bodyNode)
		}
	}
	for _, csn := range sn.Children {
		if csn.IsAttr {
			n.Attr = append(n.Attr, html.Attribute{Key: csn.Key, Val: csn.bestValueByLocale(locale, localeOrder)})
			continue
		}
		cn, err := csn.toHTMLNode(locale, localeOrder)
		if err != nil {
			return nil, err
		}
		n.AppendChild(cn)
	}
	return n, nil
}

func (p *pagespec) path(locale string) string {
	if p.pathByLocale == nil {
		return ""
	}
	return p.pathByLocale[locale]
}

func (p *pagespec) loadFromSheet(sheet *xlsx.Sheet) (err error) {
	p.pathByLocale = make(map[string]string)
	if len(sheet.Rows) == 0 {
		return fmt.Errorf("sheet %q is empty", sheet.Name)
	}
	firstLocaleCol := 0
	attrCol := 0
	p.locales = []string{}
	beyondAttrCol := false
	for col, cell := range sheet.Rows[0].Cells {
		val := strings.TrimSpace(cell.String())
		if beyondAttrCol {
			if val == "" {
				break
			}
			p.locales = append(p.locales, val)
		} else if val == "attribute" {
			attrCol = col
			firstLocaleCol = col + 1
			beyondAttrCol = true
		}
	}
	if len(p.locales) == 0 {
		return fmt.Errorf("no locales in sheet %q", sheet.Name)
	}
	maxCols := firstLocaleCol + len(p.locales)
	if len(sheet.Rows) < 2 || len(sheet.Rows[1].Cells) < maxCols || sheet.Rows[1].Cells[0].String() != "path" {
		return fmt.Errorf("path row missing or incomplete for sheet %q", sheet.Name)
	}
	for i, locale := range p.locales {
		path := strings.TrimSpace(sheet.Rows[1].Cells[firstLocaleCol+i].String())
		p.pathByLocale[locale] = path
	}

	p.root = &specNode{Key: "html"}
	stack := []*specNode{p.root}
	for i := 2; i < len(sheet.Rows); i++ {
		row := sheet.Rows[i]
		node := &specNode{SheetRow: i + 1}
		keyCol := 0
		for col, cell := range row.Cells {
			if col >= firstLocaleCol {
				return nil
			}
			node.Key = strings.TrimSpace(cell.String())
			if node.Key != "" {
				keyCol = col
				node.Depth = col + 1
				break
			}
		}
		// if first letter of key is uppercase => macro
		for i, r := range node.Key {
			if i > 0 {
				break
			}
			node.IsMacro = unicode.IsUpper(r)
		}

		node.IsAttr = keyCol == attrCol
		node.ValueByLocale = map[string]string{}
		for i, locale := range p.locales {
			if (firstLocaleCol + i) >= len(row.Cells) {
				break
			}
			val := strings.TrimSpace(row.Cells[firstLocaleCol+i].String())
			if val == "" {
				continue
			}
			node.ValueByLocale[locale] = val
		}

		for len(stack) > 0 {
			n := stack[len(stack)-1]
			if n.Depth < node.Depth {
				n.Children = append(n.Children, node)
				stack = append(stack, node)
				break
			}
			stack = stack[:len(stack)-1]
		}
	}
	return nil
}

func (p *pagespec) toHTML(locale string) (doc *html.Node, err error) {
	if p.root == nil {
		return nil, errors.New("toHTML called on empty pagespec")
	}
	doc = &html.Node{Type: html.DocumentNode}
	doc.AppendChild(&html.Node{Type: html.DoctypeNode, Data: "html", Attr: []html.Attribute{{Key: "lang", Val: locale}}})
	htmlElem, err := p.root.toHTMLNode(locale, p.locales)
	doc.AppendChild(htmlElem)
	return
}

func (p *pagespec) emit(outputDir string, perm os.FileMode) error {
	for _, locale := range p.locales {
		if err := p.emitLocale(outputDir, locale, perm); err != nil {
			return err
		}
	}
	return nil
}

func (p *pagespec) emitLocale(outputDir, locale string, perm os.FileMode) error {
	// TODO system independent
	fn := filepath.Join(outputDir, p.pathByLocale[locale])
	baseDir := filepath.Dir(fn)
	if err := os.MkdirAll(baseDir, perm); err != nil {
		return err
	}
	file, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer file.Close()
	return p.writeLocale(file, locale)
}

func (p *pagespec) writeLocale(w io.Writer, locale string) error {
	doc, err := p.toHTML(locale)
	if err != nil {
		return err
	}
	return html.Render(w, doc)
}
