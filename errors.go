package supervisor

import "fmt"

// Error represents an API error from Supervisor.
type Error struct {
	StatusCode int
	Message    string
	Details    *string
}

func (e *Error) Error() string {
	s := fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
	if e.Details != nil {
		s += ": " + *e.Details
	}
	return s
}

// IsAuthError returns true if this is a 401 authentication error.
func (e *Error) IsAuthError() bool {
	return e.StatusCode == 401
}

// IsRateLimit returns true if this is a 429 rate limit error.
func (e *Error) IsRateLimit() bool {
	return e.StatusCode == 429
}

// IsValidationError returns true if this is a 400/422 validation error.
func (e *Error) IsValidationError() bool {
	return e.StatusCode == 400 || e.StatusCode == 422
}
