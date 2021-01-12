package main

import (
	"fmt"
	"strings"
)

type stringChoiceValue struct {
	choices []string

	value string
}

func (v *stringChoiceValue) String() string {
	return v.value
}

func (v *stringChoiceValue) Set(value string) error {
	for _, choice := range v.choices {
		if choice == value {
			v.value = value
			return nil
		}
	}

	return fmt.Errorf("Value must be one off: %s", strings.Join(v.choices, ", "))
}
