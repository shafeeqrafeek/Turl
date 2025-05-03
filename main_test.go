package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShorten(t *testing.T) {
	UrlMapping = map[string]string{}
	testUrl := UrlData{Url: "http://google.com"}
	shortUrl := testUrl.Shorten()
	if len(shortUrl) != 6 {
		t.Errorf("shortened url length should be 6")
	}
	storedUrl := UrlMapping[shortUrl]
	if storedUrl != testUrl.Url {
		t.Errorf("shortened url should be %s not %s", testUrl.Url, storedUrl)
	}
}

func TestValidateUrl(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid url", "http://google.com", ""},
		{"invalid url", "invalid string", "invalid url"},
		{"missing url field", "", "missing required url field"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testUrl := UrlData{Url: testCase.input}
			err := testUrl.Validate()
			if testCase.expected != "" {
				assert.Equal(t, testCase.expected, err.Error())
			} else {
				require.NoError(t, err)
			}
		})

	}

}

func TestParseRequest(t *testing.T) {
	UrlMapping = map[string]string{}
	cases := []struct {
		name     string
		payload  string
		code     int
		response map[string]string
	}{
		{"invalid json field", `{"ur":"http://google.com"}`, http.StatusBadRequest, map[string]string{"error": "missing required url field"}},
		{"invalid json value", `{"url":"invalid string"}`, http.StatusBadRequest, map[string]string{"error": "invalid url"}},
		{"valid json value", `{"url":"http://google.com"}`, http.StatusOK, map[string]string{"short_url": ""}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/shorten", strings.NewReader(testCase.payload))
			req.Header.Add("Content-Type", "application/json")
			w := httptest.NewRecorder()
			UrlShortenHandler(w, req)
			resp := w.Result()
			testResp := map[string]string{}
			err := json.NewDecoder(resp.Body).Decode(&testResp)
			require.NoError(t, err)
			if testCase.code == http.StatusOK {
				shortCode, ok := testResp["short_url"]
				require.True(t, ok)
				assert.Len(t, shortCode, 6)
			} else {
				assert.Equal(t, testCase.code, resp.StatusCode)
				assert.Equal(t, testCase.response["error"], testResp["error"])
			}
		})
	}
}

func TestListUrls(t *testing.T) {
	UrlMapping = map[string]string{
		"abcd": "http://google.com",
		"cdef": "http://gmail.com",
	}
	defer func() { UrlMapping = map[string]string{} }()
	req := httptest.NewRequest(http.MethodGet, "/urls", nil)
	w := httptest.NewRecorder()
	GetUrlHandler(w, req)
	resp := w.Result()
	testResp := map[string]string{}
	err := json.NewDecoder(resp.Body).Decode(&testResp)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, UrlMapping, testResp)
}
