package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchHamlet(t *testing.T) {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		t.Fatal(err)
	}

	query := "Hamlet"
	req, err := http.NewRequest("GET", "/search?q="+query, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleSearch(searcher))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var results []string
	err = json.Unmarshal(rr.Body.Bytes(), &results)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, result := range results {
		if strings.Contains(strings.ToLower(result), strings.ToLower(query)) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected result not found for query: %s", query)
	}
}

func TestSearchCaseSensitive(t *testing.T) {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		t.Fatal(err)
	}

	query := "hAmLeT"
	req, err := http.NewRequest("GET", "/search?q="+query, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleSearch(searcher))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var results []string
	err = json.Unmarshal(rr.Body.Bytes(), &results)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, result := range results {
		if strings.Contains(strings.ToLower(result), strings.ToLower(query)) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected result not found for query: %s", query)
	}
}

func TestSearchDrunk(t *testing.T) {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		t.Fatal(err)
	}

	query := "drunk"
	req, err := http.NewRequest("GET", "/search?q="+query, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleSearch(searcher))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var results []string
	err = json.Unmarshal(rr.Body.Bytes(), &results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 20 {
		t.Errorf("expected 20 results for query: %s, got %d", query, len(results))
	}
}

// ebook appears at beginning and end of the file, so we can test out-of-bounds handling and pages
func TestEbook(t *testing.T) {
	query := "ebook"
	firstPage := searchQuery(t, query)
	secondPage := searchQueryPage(t, query, 2)
	thirdPage := searchQueryPage(t, query, 3)

	if len(firstPage) != 20 {
		t.Errorf("expected 20 results for first page: %s, got %d", query, len(firstPage))
	}

	if len(secondPage) != 5 {
		t.Errorf("expected 5 results for second page: %s, got %d", query, len(secondPage))
	}

	if len(thirdPage) != 0 {
		t.Errorf("expected 1 results for third page: %s, got %d", query, len(thirdPage))
	}
}

func makeRequest(t *testing.T, path string) []string {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleSearch(searcher))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var results []string
	err = json.Unmarshal(rr.Body.Bytes(), &results)
	if err != nil {
		t.Fatal(err)
	}

	return results
}

func searchQuery(t *testing.T, query string) []string {
	return makeRequest(t, fmt.Sprintf("/search?q=%s", query))
}

func searchQueryPage(t *testing.T, query string, page int) []string {
	return makeRequest(t, fmt.Sprintf("/search?q=%s&p=%d", query, page))
}
