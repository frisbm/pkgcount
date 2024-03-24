package models

import (
	"fmt"
	"regexp"
)

type Args struct {
	Help       bool
	Unrendered bool
	Exclude    string
	Lte        int
	Gte        int
	Out        string
	Dir        string

	// internal fields - not set by user
	ExcludeRegExp *regexp.Regexp
	ModuleName    string
}

func (a *Args) Validate() error {
	if a.Lte < 0 {
		return fmt.Errorf("lte must be greater than 0")
	}
	if a.Gte < 0 {
		return fmt.Errorf("gte must be greater than or equal to 0")
	}

	if a.Lte > 0 && a.Gte > 0 && a.Lte < a.Gte {
		return fmt.Errorf("lte must be greater than or equal to gte")
	}

	if a.Exclude != "" {
		excludeRegExp, err := regexp.Compile(a.Exclude)
		if err != nil {
			return fmt.Errorf("bad regular expression ['%s'] error: %w", a.Exclude, err)
		}
		a.ExcludeRegExp = excludeRegExp
	}

	return nil
}
