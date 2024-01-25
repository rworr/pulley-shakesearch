package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

const (
	resultsLim   = 20  // number of results to return per query
	resultsWidth = 250 // number of chars to return on each side of matching idx
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("shakesearch available at http://localhost:%s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		results, err := searcher.Search(query[0])
		if err != nil {
			log.Print(err) // dump failure info to logs before returning generic error to customer
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("search failure"))
			return
		}

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err = enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	return nil
}

func (s *Searcher) Search(query string) ([]string, error) {
	expr := fmt.Sprintf("(?i)%v", regexp.QuoteMeta(query)) // case-insensitive, escape regex chars
	reg, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("Unable to compile regex: %v\n, %w", expr, err)
	}

	results := make([]string, 0, resultsLim)            // update to use make with fixed capacity
	idxs := s.SuffixArray.FindAllIndex(reg, resultsLim) // update to use case-insensitive regex with fixed return limit
	for _, idx := range idxs {
		results = append(results, s.CompleteWorks[idx[0]-resultsWidth:idx[0]+resultsWidth])
	}
	return results, nil
}
