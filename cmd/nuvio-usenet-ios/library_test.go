package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const testToken = "0123456789abcdef0123456789abcdef"

func startForTest(t *testing.T) int {
	t.Helper()
	record, _ := json.Marshal(bootstrap{Token: testToken, MemoryMiB: 128, NZBCacheDir: t.TempDir()})
	out, err := start(record)
	if err != nil {
		t.Fatal(err)
	}
	var ready struct {
		Protocol int `json:"protocol"`
		Port     int `json:"port"`
	}
	if err := json.Unmarshal(out, &ready); err != nil || ready.Protocol != 1 || ready.Port == 0 {
		t.Fatalf("readiness record %s: %v", out, err)
	}
	t.Cleanup(stop)
	return ready.Port
}

func health(port int, token string) (int, string, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/health", port), nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), nil
}

func TestStartServesTheSidecarAPIUntilStopped(t *testing.T) {
	port := startForTest(t)
	if code, body, err := health(port, testToken); err != nil || code != http.StatusOK || !strings.Contains(body, `"ready":true`) {
		t.Fatalf("health = %d %q %v", code, body, err)
	}
	if code, _, err := health(port, ""); err != nil || code != http.StatusUnauthorized {
		t.Fatalf("health without token = %d %v", code, err)
	}
	stop()
	if _, _, err := health(port, testToken); err == nil {
		t.Fatal("the server still answers after stop")
	}
	stop() // a second stop is harmless
}

func TestStartReplacesTheRunningServer(t *testing.T) {
	first := startForTest(t)
	second := startForTest(t)
	if _, _, err := health(first, testToken); err == nil && first != second {
		t.Fatal("the first server still answers after a restart")
	}
	if code, _, err := health(second, testToken); err != nil || code != http.StatusOK {
		t.Fatalf("restarted health = %d %v", code, err)
	}
}

func TestStartRejectsAnInvalidBootstrapWithoutStopping(t *testing.T) {
	port := startForTest(t)
	for _, record := range []string{"", "{", `{"token":"short"}`, strings.Repeat(" ", 4<<20+1)} {
		if _, err := start([]byte(record)); err == nil {
			t.Fatalf("start(%.20q) succeeded", record)
		}
	}
	if code, _, err := health(port, testToken); err != nil || code != http.StatusOK {
		t.Fatalf("health after rejected bootstraps = %d %v", code, err)
	}
}
