package consumer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/free5gc/openapi"
)

// naiveDateTime matches an OpenAPI date-time that has no timezone.
// encoding/json rejects it: time.Parse RFC3339 requires Z or a numeric offset.
var naiveDateTime = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?$`)

// normalizeNaiveDateTimes appends Z to date-time strings that omit the timezone,
// so they decode as UTC. Strings that already carry Z or an offset are unchanged.
func normalizeNaiveDateTimes(body []byte) []byte {
	if len(body) == 0 || !json.Valid(body) {
		return body
	}
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return body
	}
	if !rewriteNaiveDateTimes(value) {
		return body
	}
	out, err := json.Marshal(value)
	if err != nil {
		return body
	}
	return out
}

func rewriteNaiveDateTimes(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		changed := false
		for key, item := range typed {
			if text, ok := item.(string); ok && naiveDateTime.MatchString(text) {
				typed[key] = text + "Z"
				changed = true
				continue
			}
			if rewriteNaiveDateTimes(item) {
				changed = true
			}
		}
		return changed
	case []any:
		changed := false
		for i, item := range typed {
			if text, ok := item.(string); ok && naiveDateTime.MatchString(text) {
				typed[i] = text + "Z"
				changed = true
				continue
			}
			if rewriteNaiveDateTimes(item) {
				changed = true
			}
		}
		return changed
	default:
		return false
	}
}

// naiveDateTimeRoundTripper rewrites CHF JSON responses before the generated
// client decodes *time.Time fields with RFC3339.
type naiveDateTimeRoundTripper struct {
	base http.RoundTripper
}

func (t naiveDateTimeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	resp, err := base.RoundTrip(req)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}

	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || strings.Contains(contentType, "json") {
		body = normalizeNaiveDateTimes(body)
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	resp.ContentLength = int64(len(body))
	resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
	return resp, nil
}

func (s *nchfService) convergedChargingHTTPClient(uri string) *http.Client {
	https := strings.HasPrefix(strings.ToLower(uri), "https://")
	client := openapi.NewHttp2Client(
		https,
		openapi.ClientTimeout,
		openapi.TransportReadIdleTimeout,
		openapi.TransportPingTimeout,
	)
	client.Transport = naiveDateTimeRoundTripper{base: client.Transport}
	return client
}
