package consumer

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/free5gc/openapi/models"
)

func TestNormalizeNaiveInvocationTimeStamp(t *testing.T) {
	body := []byte(`{"invocationTimeStamp":"2026-10-26T12:58:30.072063","invocationSequenceNumber":1}`)
	normalized := normalizeNaiveDateTimes(body)

	var response models.Chf_ConvCharging_ChargingDataResponse
	if err := json.Unmarshal(normalized, &response); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if response.InvocationTimeStamp == nil {
		t.Fatal("invocationTimeStamp is nil")
	}
	want := time.Date(2026, 10, 26, 12, 58, 30, 72063000, time.UTC)
	if !response.InvocationTimeStamp.Equal(want) {
		t.Fatalf("timestamp = %s", response.InvocationTimeStamp.Format(time.RFC3339Nano))
	}
}

func TestNormalizeKeepsTimezone(t *testing.T) {
	withZ := []byte(`{"invocationTimeStamp":"2026-10-26T12:58:30.072063Z"}`)
	if string(normalizeNaiveDateTimes(withZ)) != string(withZ) {
		t.Fatalf("Z timestamp changed: %s", normalizeNaiveDateTimes(withZ))
	}

	withOffset := []byte(`{"invocationTimeStamp":"2026-10-26T12:58:30.072063+03:00"}`)
	var response models.Chf_ConvCharging_ChargingDataResponse
	if err := json.Unmarshal(normalizeNaiveDateTimes(withOffset), &response); err != nil {
		t.Fatalf("unmarshal offset: %v", err)
	}
	want := time.Date(2026, 10, 26, 12, 58, 30, 72063000, time.FixedZone("MSK", 3*60*60))
	if !response.InvocationTimeStamp.Equal(want) {
		t.Fatalf("offset timestamp = %s", response.InvocationTimeStamp.Format(time.RFC3339Nano))
	}
}

func TestNaiveDateTimeRoundTripper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"invocationTimeStamp":"2026-10-26T12:58:30.072063"}`)
	}))
	defer server.Close()

	client := server.Client()
	client.Transport = naiveDateTimeRoundTripper{base: client.Transport}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var response models.Chf_ConvCharging_ChargingDataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("unmarshal rewritten body %s: %v", body, err)
	}
	if response.InvocationTimeStamp == nil || response.InvocationTimeStamp.Location() != time.UTC {
		t.Fatalf("timestamp = %v", response.InvocationTimeStamp)
	}
}
