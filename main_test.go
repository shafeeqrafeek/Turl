package main

import (
	"Turl/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupApp(t *testing.T) *turlApp {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		fmt.Println("DB setup for tests failed:", err)
		panic(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (id INTEGER PRIMARY KEY, url TEXT, short_code TEXT)`)
	require.NoError(t, err)
	testDB := database.DB{db}
	testApp := &turlApp{
		&testDB,
	}

	t.Cleanup(
		func() {
			err := db.Close()
			if err != nil {
				panic(err)
			}
		})

	return testApp

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
			testApp := setupApp(t)
			testApp.UrlShortenHandler(w, req)
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
	testApp := setupApp(t)
	err := testApp.DB.InsertUrl("http://google.com", "foo123")
	require.NoError(t, err)
	err = testApp.DB.InsertUrl("http://linkedin.com", "bar456")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/urls", nil)
	w := httptest.NewRecorder()
	testApp.GetUrlHandler(w, req)
	resp := w.Result()
	actual := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&actual)
	require.NoError(t, err)

	expected := map[string]string{
		"foo123": "http://google.com",
		"bar456": "http://linkedin.com",
	}
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, expected, actual)
}
