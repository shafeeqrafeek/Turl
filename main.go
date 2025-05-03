package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	urlpkg "net/url"
	"os"
	"time"
)

var UrlMapping = map[string]string{}

type UrlData struct {
	Url string `json:"url"`
}

func (u *UrlData) Shorten() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
	randomBytes := make([]byte, 6)
	for i := range randomBytes {
		randomBytes[i] = charset[rng.Intn(len(charset))]
	}
	shortId := string(randomBytes)
	UrlMapping[shortId] = u.Url
	return string(randomBytes)
}

func (u *UrlData) Validate() error {
	if u.Url == "" {
		return errors.New("missing required url field")
	}
	url, err := urlpkg.ParseRequestURI(u.Url)
	if err != nil || url.Scheme == "" || url.Host == "" {
		return errors.New("invalid url")
	}
	return nil
}

func ParseRequest(r *http.Request) (UrlData, error) {
	var data UrlData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return data, err
	}
	err = data.Validate()
	if err != nil {
		return data, err
	}
	return data, nil
}

func WriteError(w http.ResponseWriter, statusCode int, errMsg error) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(map[string]string{"error": errMsg.Error()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func UrlShortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	url, err := ParseRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	response := map[string]string{"short_url": url.Shorten()}
	respBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshalling response:", err)
	}
	_, err = w.Write(respBytes)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func GetUrlHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	responseBytes, err := json.Marshal(UrlMapping)
	if err != nil {
		fmt.Println("Error marshalling response:", err)
	}
	_, err = w.Write(responseBytes)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	shortCode := r.PathValue("shortCode")
	redirectUrl, ok := UrlMapping[shortCode]
	if !ok {
		WriteError(w, http.StatusNotFound, errors.New("short code not found"))
		return
	}
	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func main() {
	mappingFile, err := os.OpenFile("url_mapping.json", os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(mappingFile).Decode(&UrlMapping)
	if err != nil && err.Error() != "EOF" {
		log.Fatal("unable to unmarshall json mapping", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", UrlShortenHandler)
	mux.HandleFunc("GET /urls", GetUrlHandler)
	mux.HandleFunc("GET /t/{shortCode}", RedirectHandler)
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
