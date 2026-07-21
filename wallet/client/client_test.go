package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Origami74/gonuts-tollgate/cashu"
	"github.com/Origami74/gonuts-tollgate/cashu/nuts/nut03"
)

func TestNormalizeMintURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://mint.example.com/", "https://mint.example.com"},
		{"https://mint.example.com", "https://mint.example.com"},
		{"https://mint.example.com///", "https://mint.example.com"},
		{"http://localhost:3338/", "http://localhost:3338"},
		{"http://localhost:3338", "http://localhost:3338"},
		{"https://mint.example.com/Bitcoin/", "https://mint.example.com/Bitcoin"},
		{"", ""},
	}
	for _, tt := range tests {
		got := normalizeMintURL(tt.input)
		if got != tt.want {
			t.Errorf("normalizeMintURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeMintURLPreventsDoubleSlash(t *testing.T) {
	mintWithSlash := "https://mint.example.com/"
	url := normalizeMintURL(mintWithSlash) + "/v1/keysets"
	if strings.Contains(url, "//v1") {
		t.Errorf("double slash in URL: %s", url)
	}
}

func TestGetHandles404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not_found","code":10000}`))
	}))
	defer server.Close()

	_, err := get(server.URL + "/v1/keysets")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "not_found") {
		t.Errorf("error should contain 'not_found', got: %v", err)
	}
}

func TestGetHandles400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"test-error","error":"something went wrong"}`))
	}))
	defer server.Close()

	_, err := get(server.URL + "/v1/swap")
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestGetHandles200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"keysets":[]}`))
	}))
	defer server.Close()

	resp, err := get(server.URL + "/v1/keysets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetAllKeysetsNormalizesURL(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"keysets":[]}`))
	}))
	defer server.Close()

	_, err := GetAllKeysets(server.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error with trailing slash: %v", err)
	}
	if receivedPath != "/v1/keysets" {
		t.Errorf("path = %s, want /v1/keysets", receivedPath)
	}
}

func TestGetAllKeysetsParsesValidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"keysets":[{"id":"00003ac3b4d68224","unit":"sat","active":true,"input_fee_ppk":0}]}`))
	}))
	defer server.Close()

	result, err := GetAllKeysets(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Keysets) != 1 {
		t.Fatalf("expected 1 keyset, got %d", len(result.Keysets))
	}
	if result.Keysets[0].Id != "00003ac3b4d68224" {
		t.Errorf("id = %s", result.Keysets[0].Id)
	}
	if !result.Keysets[0].Active {
		t.Error("expected active=true")
	}
}

func TestParseHandlesEmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("")),
	}
	_, err := parse(resp)
	if err == nil {
		t.Fatal("expected error for 500 with empty body")
	}
}

func TestParseHandlesLargeBody(t *testing.T) {
	large := strings.Repeat("x", 2*1024*1024)
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader(large)),
	}
	_, err := parse(resp)
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestGetMintInfoNormalizesURL(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test"}`))
	}))
	defer server.Close()

	_, err := GetMintInfo(server.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedPath != "/v1/info" {
		t.Errorf("path = %s, want /v1/info", receivedPath)
	}
}

func TestPostSwapNormalizesURL(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"signatures":[]}`))
	}))
	defer server.Close()

	swapReq := nut03.PostSwapRequest{
		Inputs:  cashu.Proofs{},
		Outputs: cashu.BlindedMessages{},
	}
	_, err := PostSwap(server.URL+"/", swapReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedPath != "/v1/swap" {
		t.Errorf("path = %s, want /v1/swap", receivedPath)
	}
}

func TestParseAppliesResponseLimit(t *testing.T) {
	large := strings.Repeat("x", 2*int(maxResponseBytes))
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader(large)),
	}
	_, err := parse(resp)
	if err == nil {
		t.Fatal("expected error for 500 with oversized body")
	}
	if !strings.Contains(err.Error(), "x") {
		t.Errorf("error should contain trimmed body content, got: %v", err)
	}
}
