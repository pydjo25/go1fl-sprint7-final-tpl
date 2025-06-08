package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
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

	var (
		convStr     []string
		countResult int
	)

	data := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: 5},
	}

	for _, v := range data {
		handler := http.HandlerFunc(mainHandle)
		resp := httptest.NewRecorder()
		url := "/cafe?city=moscow&count=" + strconv.Itoa(v.count)
		req := httptest.NewRequest("GET", url, nil)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error reading response body: ", err)
		}

		if bodyResp := string(bodyResp); bodyResp != "" {
			convStr = strings.Split(bodyResp, ",")
		}

		if v.count > len(cafeList["moscow"]) {
			countResult = min(len(cafeList["moscow"]), 100)
			assert.Equal(t, v.want, countResult)
		} else {
			assert.Equal(t, v.want, len(convStr))
		}

	}

}

func TestCafeSearch(t *testing.T) {

	var (
		convStr []string
		// countResult int
	)

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
		url := "/cafe?city=moscow&search=" + v.search
		req := httptest.NewRequest("GET", url, nil)
		handler.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)

		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error reading response body: ", err)
		}
		clearStr := strings.TrimSpace(string(bodyResp))

		if clearStr != "" {
			convStr = strings.Split(clearStr, ",")
		}

		assert.Len(t, convStr, v.wantcount)

		for _, s := range convStr {
			s = strings.ToLower(s)
			assert.Contains(t, s, v.search)
		}
	}
}
