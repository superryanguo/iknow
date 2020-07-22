package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKnowHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(PostDataHandler))
	defer ts.Close()
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Errorf("Error occured while constructing request: %s", err)
	}

	w := httptest.NewRecorder()
	KnowHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Actual status: (%d); Expected status:(%d)", w.Code, http.StatusOK)
	}
}
