package fiber_errhandler

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func newApp() *fiber.App {
	app := fiber.New()

	app.Use(New())

	app.Get("/user", func(c *fiber.Ctx) {
		c.JSON(fiber.Map{
			"FirstName": "John",
			"LastName":  "Doe",
		})
	})

	app.Get("/err", func(c *fiber.Ctx) {
		c.Next(errors.New("bad thing happens"))
	})
	app.Get("/400", func(c *fiber.Ctx) {
		c.Next(NewHttpError(fiber.StatusBadRequest, "Bad request", fiber.Map{
			"Field": "Not empty",
		}))
	})
	app.Get("/panic", func(c *fiber.Ctx) {
		panic("i'm panic")
	})

	return app
}

func TestErrHandler_normal_condition(t *testing.T) {
	app := newApp()

	req := httptest.NewRequest("GET", "/user", nil)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		b := make(map[string]interface{})
		if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, map[string]interface{}{
				"FirstName": "John",
				"LastName":    "Doe",
			}, b)
		}
	}
}

func TestErrHandler_plaintext(t *testing.T) {
	app := newApp()

	req := httptest.NewRequest("GET", "/err", nil)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "bad thing happens", string(b))
		}
	}

	req = httptest.NewRequest("GET", "/400", nil)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "Bad request", string(b))
		}
	}

	req = httptest.NewRequest("GET", "/panic", nil)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "i'm panic", string(b))
		}
	}
}

func TestErrHandler_json(t *testing.T) {
	app := newApp()

	req := httptest.NewRequest("GET", "/err", nil)
	req.Header.Set("Accept", "application/json")
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
	req.Header.Set("Accept", "application/json")
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
	req.Header.Set("Accept", "application/json")
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
