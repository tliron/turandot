package main

import (
	"strconv"
	"strings"
	"unicode"
)

func isVersionNewer(newer string, older string) bool {
	if newer_ := NewVersion(newer); newer_ != nil {
		if older_ := NewVersion(older); older_ != nil {
			return newer_.IsNewer(older_)
		}
	}
	return false
}

//
// Version
//

type Version struct {
	Major int
	Minor int
	Patch int
}

// e.g. v1alpha2
func NewVersion(text string) *Version {
	if strings.HasPrefix(text, "v") {
		phase := 1
		var major, modifier, patch string
		for _, rune_ := range text[1:] {
			if phase == 1 {
				if unicode.IsDigit(rune_) {
					major += string(rune_)
				} else {
					phase = 2
				}
			}

			if phase == 2 {
				if unicode.IsDigit(rune_) {
					phase = 3
				} else {
					modifier += string(rune_)
				}
			}

			if phase == 3 {
				patch += string(rune_)
			}
		}

		major_, _ := strconv.Atoi(major)
		patch_, _ := strconv.Atoi(patch)

		var minor_ int
		switch modifier {
		case "alpha":
			minor_ = -2
		case "beta":
			minor_ = -1
		}

		return &Version{
			Major: major_,
			Minor: minor_,
			Patch: patch_,
		}
	} else {
		return nil
	}
}

func (self *Version) IsNewer(older *Version) bool {
	if self.Major > older.Major {
		return true
	} else if self.Major == older.Major {
		if self.Minor > older.Minor {
			return true
		} else if self.Minor == older.Minor {
			return self.Patch > older.Patch
		} else {
			return false
		}
	} else {
		return false
	}
}
