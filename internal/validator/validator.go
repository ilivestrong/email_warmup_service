package validator

import "strings"

type Validator struct{ m map[string]struct{} }

func New(list []string) *Validator {
	m := map[string]struct{}{}
	for _, d := range list {
		m[strings.ToLower(d)] = struct{}{}
	}
	return &Validator{m}
}

func (v *Validator) IsValid(e string) bool {
	p := strings.Split(e, "@")
	if len(p) != 2 {
		return false
	}
	_, bad := v.m[strings.ToLower(p[1])]
	return !bad
}
