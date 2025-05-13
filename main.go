package main

import (
	"Turl/database"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	urlpkg "net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var UrlMapping = map[string]string{}

type UrlData struct {
	Url string `json:"url"`
}

type turlApp struct {
	DB *database.DB
}

func (u *UrlData) Shorten() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
	randomBytes := make([]byte, 6)
	for i := range randomBytes {
		randomBytes[i] = charset[rng.Intn(len(charset))]
	}
	shortId := string(randomBytes)
	return shortId
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

func (t *turlApp) UrlShortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	url, err := ParseRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}
	shortId := url.Shorten()
	UrlMapping[shortId] = url.Url
	err = t.DB.InsertUrl(url.Url, shortId)
	if err != nil {
		fmt.Println(err)
		WriteError(w, http.StatusInternalServerError, errors.New("unable to insert url into db"))
		return
	}

	response := map[string]string{"short_url": shortId}
	respBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshalling response:", err)
	}
	_, err = w.Write(respBytes)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func (t *turlApp) GetUrlHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	urlData, err := t.DB.GetUrls()
	if err != nil {
		fmt.Println(err)
		WriteError(w, http.StatusInternalServerError, errors.New("unable to get url data"))
		return
	}
	//responseBytes, err := json.Marshal(UrlMapping)
	responseBytes, err := json.Marshal(urlData)
	if err != nil {
		fmt.Println("Error marshalling response:", err)
	}
	_, err = w.Write(responseBytes)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func (t *turlApp) RedirectHandler(w http.ResponseWriter, r *http.Request) {
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
	DB, err := database.CreateDB()
	defer DB.Close()
	if err != nil {
		log.Fatal("unable to create db:", err)
	}
	rows, err := DB.Query("Select count(*) from urls")
	if err != nil {
		log.Fatal("unable to query urls table:", err)
	}
	var urlsCount *int
	rows.Next()
	err = rows.Scan(&urlsCount)
	rows.Close()
	if err != nil {
		log.Fatal("unable to scan urls count:", err)
	}
	fmt.Println("URLs count:", *urlsCount)

	var osSignal = make(chan os.Signal, 1)

	readFile, err := os.OpenFile("url_mapping.json", os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal("unable to open mapping file:", err)
	}
	defer func(readFile *os.File) {
		err := readFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(readFile)

	err = json.NewDecoder(readFile).Decode(&UrlMapping)
	if err != nil && err.Error() != "EOF" {
		log.Fatal("unable to unmarshall json mapping", err)
	}
	mux := http.NewServeMux()
	app := &turlApp{DB: DB}
	mux.HandleFunc("POST /shorten", app.UrlShortenHandler)
	mux.HandleFunc("GET /urls", app.GetUrlHandler)
	mux.HandleFunc("GET /t/{shortCode}", app.RedirectHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	signal.Notify(osSignal, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-osSignal
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	writeFile, err := os.OpenFile("url_mapping.json", os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatal("unable to open mapping file:", err)
	}
	defer func(writeFile *os.File) {
		err := writeFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(writeFile)
	err = json.NewEncoder(writeFile).Encode(UrlMapping)
	if err != nil {
		log.Fatal("unable to write json mapping:", err)
	}
}
