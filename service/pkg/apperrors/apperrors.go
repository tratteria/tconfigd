package apperrors

import "errors"

var ErrVerificationRuleNotFound = errors.New("verification rule not found for the service")