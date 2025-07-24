package validator

import (
	"net/mail"
	"strings"
)

type Validator struct {
	disposableDomains map[string]struct{}
	// disposable map[string]struct{}
	zeroBounce *zeroBounceClient
}

func New(disposableDomains []string, zeroBounceClient *zeroBounceClient) *Validator {
	m := map[string]struct{}{}
	for _, domain := range disposableDomains {
		m[strings.ToLower(domain)] = struct{}{}
	}
	return &Validator{m, zeroBounceClient}
}

func (v *Validator) IsValid(address string) bool {
	if _, err := mail.ParseAddress(address); err != nil {
		return false
	}

	p := strings.Split(address, "@")
	if len(p) != 2 {
		return false
	}
	if _, bad := v.disposableDomains[strings.ToLower(p[1])]; bad {
		return !bad
	}

	if res := v.zeroBounce.Validate(address); res.Err != nil {
		return false
	}
	return true
}
