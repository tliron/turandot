package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/mitchellh/go-wordwrap"
)

var whitespaceRe = regexp.MustCompile(`[ \t][ \t]+`)

const textWidth = 80

func (self *Generator) Println(args ...interface{}) {
	for _, arg := range args {
		self.writer.WriteString(fmt.Sprintf("%s", arg))
	}
	self.writer.WriteString("\n")
}

func (self *Generator) Printf(format string, args ...interface{}) {
	self.writer.WriteString(fmt.Sprintf(format, args...))
}

func (self *Generator) Printlnf(format string, args ...interface{}) {
	self.writer.WriteString(fmt.Sprintf(format, args...))
	self.writer.WriteString("\n")
}

func (self *Generator) PrintText(indent int, key string, text string, forceBlock bool) {
	text = strings.TrimSpace(text)
	text = whitespaceRe.ReplaceAllString(text, " ")
	if text == "" {
		return
	}

	quoted := quote(text)

	block := forceBlock
	if !block {
		if len(quoted) > textWidth {
			block = true
		} else if strings.Contains(text, "\n") {
			block = true
		}
	}

	indent_ := strings.Repeat(" ", indent)
	if block {
		text = wordwrap.WrapString(text, textWidth)
		paragraphs := strings.Split(text, "\n")
		if len(paragraphs) == 1 {
			// A single line will not be be a block
			self.Printlnf("%s%s: %s", indent_, key, quoted)
		} else {
			self.Printlnf("%s%s: >-", indent_, key)
			indent_ = strings.Repeat(" ", indent+2)
			for _, paragraph := range paragraphs {
				paragraph = strings.TrimRightFunc(paragraph, unicode.IsSpace)
				if paragraph != "" {
					self.Printlnf("%s%s", indent_, paragraph)
				} else {
					self.Println()
				}
			}
		}
	} else {
		self.Printlnf("%s%s: %s", indent_, key, quoted)
	}
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
