package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/mitchellh/go-wordwrap"
)

const textWidth = 80

var whitespaceRe = regexp.MustCompile(`[ \t][ \t]+`)

func (self *Generator) Writeln(args ...interface{}) {
	for _, arg := range args {
		self.Writer.WriteString(fmt.Sprintf("%s", arg))
	}
	self.Writer.WriteString("\n")
}

func (self *Generator) Writef(format string, args ...interface{}) {
	self.Writer.WriteString(fmt.Sprintf(format, args...))
}

func (self *Generator) Writelnf(format string, args ...interface{}) {
	self.Writer.WriteString(fmt.Sprintf(format, args...))
	self.Writer.WriteString("\n")
}

func (self *Generator) WriteKey(indent int, key string) {
	self.Writelnf("%s%s:", indentation(indent), key)
}

func (self *Generator) WriteKeyValue(indent int, key string, text string, forceBlock bool) {
	text = strings.TrimSpace(text)
	text = whitespaceRe.ReplaceAllString(text, " ")
	if text == "" {
		return
	}

	quoted := quote(text)

	var paragraphs []string
	if forceBlock || (len(quoted) > textWidth) || strings.Contains(text, "\n") {
		text = wordwrap.WrapString(text, textWidth)
		paragraphs = strings.Split(text, "\n")
		if !forceBlock && (len(paragraphs) == 1) {
			// A single line will not be be a block
			paragraphs = nil
		}
	}

	if len(paragraphs) > 0 {
		self.Writelnf("%s%s: >-", indentation(indent), key)
		indent_ := indentation(indent + 1)
		for _, paragraph := range paragraphs {
			paragraph = strings.TrimRightFunc(paragraph, unicode.IsSpace)
			if paragraph != "" {
				self.Writelnf("%s%s", indent_, paragraph)
			} else {
				self.Writeln()
			}
		}
	} else {
		self.Writelnf("%s%s: %s", indentation(indent), key, quoted)
	}
}

// Utils

func indentation(indent int) string {
	return strings.Repeat("  ", indent)
}

func quote(s string) string {
	toQuote := false

	if _, err := strconv.ParseFloat(s, 64); err == nil {
		// YAML will consider this a float
		toQuote = true
	}

	// Note: there are many more reasons to quote, but these
	// checks are good enough for our data
	if len(s) > 0 {
		switch s[0] {
		case '`', '!':
			toQuote = true
		}
	}
	if strings.Contains(s, ": ") {
		toQuote = true
	}

	if toQuote {
		s = strings.ReplaceAll(s, "'", "''")
		return "'" + s + "'"
	} else {
		return s
	}
}
