package validator

import (
	"fmt"

	"github.com/ilivestrong/email_warmup_service/internal/config"
	"github.com/zerobounce/zerobouncego"
)

type (
	zeroBounceClient          struct{}
	zeroBoundValidationResult struct {
		Success bool
		Err     error
	}
)

func NewZeroBounceClient(zbCfg config.ZeroBounceConfig) *zeroBounceClient {
	zerobouncego.SetApiKey(zbCfg.ApiKey)
	return &zeroBounceClient{}
}

func (zbv *zeroBounceClient) Validate(email string) zeroBoundValidationResult {
	resp, err := zerobouncego.Validate(email, "")
	if err != nil {
		return zeroBoundValidationResult{false, fmt.Errorf("failed to validate email: %v", err)}
	}
	if resp.Status == zerobouncego.S_INVALID {
		return zeroBoundValidationResult{false, fmt.Errorf("invalid email: %s", email)}
	}
	return zeroBoundValidationResult{true, nil}
}
