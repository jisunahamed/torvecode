package provider

import (
	"net/http"
	"strings"
	"testing"

	"github.com/openai/openai-go"
)

func TestTorveGatewayErrorExplainsPlanLimits(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"rpm_limit_exceeded", "RPM"},
		{"tpm_limit_exceeded", "TPM"},
		{"insufficient_balance", "balance"},
		{"monthly_limit_exceeded", "monthly"},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			apierr := &openai.Error{Code: test.code, Response: &http.Response{Header: http.Header{"Retry-After": []string{"12"}}}}
			err := torveGatewayError(apierr)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error=%v, want %q", err, test.want)
			}
		})
	}
}

func TestPlanLimitIsNotRetriedAsTransientFailure(t *testing.T) {
	apierr := &openai.Error{
		Code:       "tpm_limit_exceeded",
		StatusCode: http.StatusTooManyRequests,
		Response:   &http.Response{Header: http.Header{"Retry-After": []string{"12"}}},
	}
	retry, _, err := (&openaiClient{}).shouldRetry(1, apierr)
	if retry || err == nil || !strings.Contains(err.Error(), "TPM") {
		t.Fatalf("retry=%t error=%v", retry, err)
	}
}

func TestTorveGatewayErrorLeavesUnknownErrorsUntouched(t *testing.T) {
	if err := torveGatewayError(&openai.Error{Code: "upstream_unavailable"}); err != nil {
		t.Fatalf("unexpected mapping: %v", err)
	}
}
