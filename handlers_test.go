package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ckblck/gocryotic/network"
	"github.com/gofiber/fiber/v2"
)

type form struct {
	key   string
	value string
}

func TestBasicHandler(t *testing.T) {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("👋")
	})

	resp := newResponse(app, "GET", "/")
	got := resp.StatusCode
	want := 200

	if got != want {
		t.Errorf("got %v want %v", got, want)
	}

	gotResponse, _ := ioutil.ReadAll(resp.Body)
	wantResponse := "👋"

	if string(gotResponse) != wantResponse {
		t.Errorf("got %v want %v", gotResponse, want)
	}
}

func TestAddPlayerRequest(t *testing.T) {
	app := fiber.New()
	app.Post("/api/v1/player", network.AddPlayer)

	form := url.Values{}
	form.Add("?", "?")

	request := makeRequestWForm("POST", "/api/v1/player", form)
	resp := sendRequest(app, request)

	got := resp.StatusCode
	notWant := 201

	if got == notWant {
		t.Errorf("got %v and not want %v", got, notWant)
	}

}

func TestGetReplayByIDRequest(t *testing.T) {
	app := fiber.New()
	app.Get("/api/v1/replay/:id", network.GetReplay)

	resp := newResponse(app, "GET", "/api/v1/replay/????")

	got := resp.StatusCode
	want := 404

	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestPostReplayRequest(t *testing.T) {
	app := fiber.New()
	app.Post("/api/v1/replay", network.AddReplay)

	form := url.Values{}
	form.Add("file", "?")

	request := makeRequestWForm("POST", "/api/v1/replay", form)
	resp := sendRequest(app, request)

	got := resp.StatusCode
	want := 422

	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func makeRequestWForm(method, route string, form url.Values) (req *http.Request) {
	req = httptest.NewRequest(method, route, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return
}

func sendRequest(app *fiber.App, req *http.Request) *http.Response {
	resp, _ := app.Test(req)

	return resp
}

func newResponse(app *fiber.App, method, route string) *http.Response {
	req := httptest.NewRequest(method, route, nil)

	return sendRequest(app, req)
}
