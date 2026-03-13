package main

import "strings"

// toolNameList supports repeated flags and comma-separated values.
type toolNameList struct {
	values []string
	set    bool
}

func (l *toolNameList) String() string {
	return strings.Join(l.values, ",")
}

func (l *toolNameList) Set(v string) error {
	l.set = true
	for _, item := range strings.Split(v, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		l.values = append(l.values, item)
	}
	return nil
}

func (l *toolNameList) Values() []string {
	if len(l.values) == 0 {
		return nil
	}
	out := make([]string, len(l.values))
	copy(out, l.values)
	return out
}

func (l *toolNameList) IsSet() bool {
	return l.set
}
