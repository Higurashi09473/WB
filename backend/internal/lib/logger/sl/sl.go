// Package sl provides small utility helpers for working with slog.Logger.
// Currently it contains a convenient way to log errors as attributes.
package sl

import "log/slog"

// Err returns a slog.Attr that represents an error value.
// Useful for structured logging: log.Error("something failed", sl.Err(err))
func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}

	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}