package http

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/delfimarime/sauron/pkg/telemetry"
)

type LoggerRoundTripper struct {
	logger *zap.Logger
	next   http.RoundTripper
}

func NewLoggerRoundTripper(
	logger *zap.Logger,
	next http.RoundTripper,
) *LoggerRoundTripper {
	return &LoggerRoundTripper{
		logger: logger, next: next,
	}
}

func (rt *LoggerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	path := "/"
	host := ""
	if req.URL != nil {
		path = req.URL.Path
		host = req.URL.Host
	}
	mimeType := ""
	if req.Body != nil && req.Header.Get(telemetry.HeaderContentType) != "" {
		mimeType = req.Header.Get(telemetry.HeaderContentType)
	}
	ctx := req.Context()
	rt.logger.Debug("Submitting http request to remote server",
		zap.Any(telemetry.KeyHTTP, map[string]any{
			telemetry.KeyRequest: map[string]any{
				telemetry.KeyMethod:   req.Method,
				telemetry.KeyURL:      path,
				telemetry.KeyHost:     host,
				telemetry.KeyMimeType: mimeType,
				telemetry.KeyVersion:  req.Proto,
			},
		}),
	)

	resp, err := rt.next.RoundTrip(req.WithContext(ctx))

	if err != nil {
		rt.logger.Debug("An error was thrown attempting to submit http request to remote server",
			zap.Error(err),
		)
	} else {
		rt.logger.Debug("The http request was submitted to remote server",
			zap.Any(telemetry.KeyHTTP, map[string]any{
				telemetry.KeyRequest: map[string]any{
					telemetry.KeyMethod:   req.Method,
					telemetry.KeyURL:      path,
					telemetry.KeyHost:     host,
					telemetry.KeyMimeType: mimeType,
					telemetry.KeyVersion:  req.Proto,
				},
				telemetry.KeyResponse: map[string]any{
					telemetry.KeyStatusCode: resp.StatusCode,
					telemetry.KeyMimeType:   resp.Header.Get(telemetry.HeaderContentType),
				},
			}),
		)
	}
	return resp, err
}

func WithLoggerRoundTripper(logger *zap.Logger) func(*http.Client) error {
	return func(c *http.Client) error {
		c.Transport = NewLoggerRoundTripper(logger, c.Transport)
		return nil
	}
}
