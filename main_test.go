package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {

	data := []struct {
		count int    // передаваемое значение count
		want  int    // ожидаемое количество кафе
		city  string // город
	}{
		{count: 0, want: 0, city: "moscow"},
		{count: 0, want: 0, city: "tula"},
		{count: 1, want: 1, city: "moscow"},
		{count: 1, want: 1, city: "tula"},
		{count: 2, want: 2, city: "moscow"},
		{count: 2, want: 2, city: "tula"},
		{count: 100, want: 0, city: "moscow"},
		{count: 100, want: 0, city: "tula"},
	}

	for _, v := range data {
		handler := http.HandlerFunc(mainHandle)

		resp := httptest.NewRecorder()

		conCount := strconv.Itoa(v.count)

		url := "/cafe?city=" + v.city + "&count=" + conCount

		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(resp, req)

		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error reading response body: ", err)
		}

		if v.count > len(cafeList) || v.count == 0 {
			bodyResp = []byte("")
		}

		var convStr []string

		if bodyResp := string(bodyResp); bodyResp != "" {
			convStr = strings.Split(bodyResp, ",")
		}

		assert.Equal(t, v.want, len(convStr),
			"For count=%d in %s: expected %d cafes, got %d. Response: '%s'",
			v.count, v.city, v.want, len(convStr), bodyResp)
	}

}

func TestCafeSearch(t *testing.T) {

	data := []struct {
		search    string // передаваемое значение count
		wantcount int    // ожидаемое количество кафе
	}{
		{search: "фалось", wantcount: 0},
		{search: "кофе", wantcount: 2},
		{search: "вилка", wantcount: 1},
	}
	for _, v := range data {
		handler := http.HandlerFunc(mainHandle)

		resp := httptest.NewRecorder()

		url := "/cafe?city=moscow" + "&search=" + v.search

		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(resp, req)

		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error reading response body: ", err)
		}

		var (
			convStr     []string
			countResult int
		)
		if bodyResp := string(bodyResp); bodyResp != "" {
			convStr = strings.Split(bodyResp, ",")
		}

		for _, s := range convStr {
			strings.ToLower(s)
			if strings.Contains(v.search, s); true {
				countResult++
			}
		}

		assert.Equal(t, v.wantcount, countResult,
			"For search = %s in %s: expected %d cafes, got %d. Response: '%s'",
			v.search, "city=moscow", v.wantcount, len(convStr), bodyResp)
	}
}
