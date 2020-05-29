package view_test

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber"
	"github.com/gofiber/template/html"
	errhandler "github.com/hendratommy/fiber-errhandler"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

var browserAccept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"

func newApp(useTemplate bool) *fiber.App {
	app := fiber.New()
	app.Settings.Templates = html.New("./views", ".html")

	app.Use(errhandler.New(errhandler.Config{
		UseTemplate: useTemplate,
	}))

	app.Get("/user", func(c *fiber.Ctx) {
		c.Render("user", fiber.Map{
			"FirstName": "John",
			"LastName":  "Doe",
		})
	})

	app.Get("/err", func(c *fiber.Ctx) {
		c.Next(errors.New("bad thing happens"))
	})
	app.Get("/400", func(c *fiber.Ctx) {
		c.Next(errhandler.NewHttpError(fiber.StatusBadRequest, "Bad request", fiber.Map{
			"Field": "Not empty",
		}))
	})
	app.Get("/panic", func(c *fiber.Ctx) {
		panic("i'm panic")
	})

	return app
}

func TestErrHandler_view_normal_condition(t *testing.T) {
	app := newApp(false)

	req := httptest.NewRequest("GET", "/user", nil)
	req.Header.Set("Accept", browserAccept)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "John<br />Doe", string(b))
		}
	}
}

func TestErrHandler_view_notemplate(t *testing.T) {
	app := newApp(false)

	req := httptest.NewRequest("GET", "/err", nil)
	req.Header.Set("Accept", browserAccept)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		b := make(map[string]interface{})
		if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, map[string]interface{}{
				"message": "bad thing happens",
				"error":    "bad thing happens",
			}, b)
		}
	}

	req = httptest.NewRequest("GET", "/400", nil)
	req.Header.Set("Accept", browserAccept)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		b := make(map[string]interface{})
		if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, map[string]interface{}{
				"message": "Bad request",
				"error": map[string]interface{}{
					"Field": "Not empty",
				},
			}, b)
		}
	}

	req = httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set("Accept", browserAccept)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		b := make(map[string]interface{})
		if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, map[string]interface{}{
				"message": "i'm panic",
				"error":    "i'm panic",
			}, b)
		}
	}
}

func TestErrHandler_view_template(t *testing.T) {
	app := newApp(true)

	req := httptest.NewRequest("GET", "/err", nil)
	req.Header.Set("Accept", browserAccept)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "bad thing happens<br />bad thing happens", string(b))
		}
	}

	req = httptest.NewRequest("GET", "/400", nil)
	req.Header.Set("Accept", browserAccept)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "Bad request<br />Not empty", string(b))
		}
	}

	req = httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set("Accept", browserAccept)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "i&#39;m panic<br />i&#39;m panic", string(b))
		}
	}
}
