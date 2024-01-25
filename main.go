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
	"strconv"
)

const (
	resultsPerPage = 20  // number of results to return per query
	resultsWidth   = 250 // number of chars to return on each side of matching idx
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
	indexCache    map[string][][]int
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		searchQuery, ok := query["q"]
		if !ok || len(searchQuery[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		var page = 1
		pageQuery, ok := query["p"]
		if ok {
			var err error
			page, err = strconv.Atoi(pageQuery[0])
			if err != nil || page < 1 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("malformed page in URL params"))
				return
			}
		}

		results, err := searcher.Search(searchQuery[0], page)
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
	s.indexCache = make(map[string][][]int)
	return nil
}

func (s *Searcher) Search(query string, page int) ([]string, error) {
	indexes, ok := s.indexCache[query]

	if !ok {
		expr := fmt.Sprintf("(?i)%v", regexp.QuoteMeta(query))
		reg, err := regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("Unable to compile regex: %v\n, %w", expr, err)
		}

		idxs := s.SuffixArray.FindAllIndex(reg, -1)
		s.indexCache[query] = idxs
		indexes = idxs
	}

	baseIdx := (page - 1) * resultsPerPage
	resultsLen := clamp(len(indexes)-baseIdx, 0, resultsPerPage)

	results := make([]string, 0, resultsLen)
	for i := 0; i < resultsLen; i++ {
		idx := baseIdx + indexes[i][0]
		startIdx := max(idx-resultsWidth, 0)
		endIdx := min(idx+resultsWidth, len(s.CompleteWorks)-1)
		results = append(results, s.CompleteWorks[startIdx:endIdx])
	}
	return results, nil
}

// TODO: upgrade to Go 1.21 and use builtins
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TODO: upgrade to Go 1.21 and use builtins
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clamp(val, minval, maxval int) int {
	return max(minval, min(maxval, val))
}
