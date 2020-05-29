package fiber_errhandler

import (
	"encoding/json"
	"errors"
	"fmt"
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
	req.Header.Set("Accept", "*/*")
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
	req.Header.Set("Accept", fiber.MIMETextPlain)
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

func TestErrHandler_test_filter(t *testing.T) {
	app := fiber.New()
	app.Use(New(Config{
		Filter: func(c *fiber.Ctx) bool {
			return c.Path() == "/err"
		},
		Log: true,
	}))
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

	req := httptest.NewRequest("GET", "/err", nil)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "", string(b))
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
	req.Header.Set("Content-Type", "application/json")
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
	req.Header.Set("Content-Type", "application/json; utf-8")
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

func TestErrHandler_custom_handler(t *testing.T) {
	app := fiber.New()
	app.Use(New(Config{
		Handler: func(c *fiber.Ctx, err error, fn func(...interface{})) {
			if c.Path() == "/err" {
				c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
				return
			} else if c.Path() == "/400" {
				fn()
				return
			}
			fn("handled by custom handler")
		},
	}))

	app.Get("/err", func(c *fiber.Ctx) {
		c.Next(NewHttpError(0, "custom error handler", nil))
	})
	app.Get("/400", func(c *fiber.Ctx) {
		c.Next(NewHttpError(0, "custom error handler", nil))
	})
	app.Get("/500", func(c *fiber.Ctx) {
		c.Next(NewHttpError(0, "custom error handler", nil))
	})

	req := httptest.NewRequest("GET", "/err", nil)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, fmt.Sprintf(`{"message":"%s"}`,
				NewHttpError(0, "custom error handler", nil).Error()), string(b))
		}
	}

	req = httptest.NewRequest("GET", "/400", nil)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "Internal Server Error", string(b))
		}
	}

	req = httptest.NewRequest("GET", "/500", nil)
	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "handled by custom handler", string(b))
		}
	}
}

var _benchmark_string_original_fiber int
func Benchmark_string_original_fiber(b *testing.B) {
	//fmt.Println("Benchmark original fiber SendString")

	app := fiber.New()
	app.Get("/hello", func(c *fiber.Ctx) {
		c.SendString("Hello World")
	})

	var _statusCode int

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest("GET", "/hello", nil)
		if resp, err := app.Test(req); err != nil {
			assert.NoError(b, err)
		} else {
			assert.Equal(b, fiber.StatusOK, resp.StatusCode)
			_statusCode = resp.StatusCode
		}
	}
	_benchmark_string_original_fiber = _statusCode
}

var _benchmark_errhandler_string int
func Benchmark_errhandler_string(b *testing.B) {
	//fmt.Println("Benchmark errhandler non error SendString")

	app := fiber.New()
	app.Use(New())
	app.Get("/hello", func(c *fiber.Ctx) {
		c.SendString("Hello World")
	})

	var _statusCode int

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest("GET", "/hello", nil)
		if resp, err := app.Test(req); err != nil {
			assert.NoError(b, err)
		} else {
			assert.Equal(b, fiber.StatusOK, resp.StatusCode)
			_statusCode = resp.StatusCode
		}
	}
	_benchmark_errhandler_string = _statusCode
}

var _benchmark_errhandler_string_err int
func Benchmark_errhandler_string_err(b *testing.B) {
	//fmt.Println("Benchmark errhandler with error SendString")

	app := fiber.New()
	app.Use(New())
	app.Get("/hello", func(c *fiber.Ctx) {
		c.Next(NewHttpError(fiber.StatusBadRequest, "Bad request", nil))
	})

	var _statusCode int

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest("GET", "/hello", nil)
		if resp, err := app.Test(req); err != nil {
			assert.NoError(b, err)
		} else {
			assert.Equal(b, fiber.StatusBadRequest, resp.StatusCode)
			_statusCode = resp.StatusCode
		}
	}
	_benchmark_errhandler_string_err = _statusCode
}

var _benchmark_json_original_fiber int
func Benchmark_json_original_fiber(b *testing.B) {
	//fmt.Println("Benchmark original fiber JSON")

	app := fiber.New()
	app.Get("/hello", func(c *fiber.Ctx) {
		c.JSON(fiber.Map{
			"message": "Hello World",
		})
	})

	var _statusCode int

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest("GET", "/hello", nil)
		if resp, err := app.Test(req); err != nil {
			assert.NoError(b, err)
		} else {
			assert.Equal(b, fiber.StatusOK, resp.StatusCode)
			_statusCode = resp.StatusCode
		}
	}
	_benchmark_json_original_fiber = _statusCode
}

var _benchmark_errhandler_json int
func Benchmark_errhandler_json(b *testing.B) {
	//fmt.Println("Benchmark errhandler non error JSON")

	app := fiber.New()
	app.Get("/hello", func(c *fiber.Ctx) {
		c.JSON(fiber.Map{
			"message": "Hello World",
		})
	})

	var _statusCode int

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest("GET", "/hello", nil)
		if resp, err := app.Test(req); err != nil {
			assert.NoError(b, err)
		} else {
			assert.Equal(b, fiber.StatusOK, resp.StatusCode)
			_statusCode = resp.StatusCode
		}
	}
	_benchmark_errhandler_json = _statusCode
}

var _benchmark_errhandler_json_err int
func Benchmark_errhandler_json_err(b *testing.B) {
	//fmt.Println("Benchmark errhandler non error JSON")

	app := fiber.New()
	app.Get("/hello", func(c *fiber.Ctx) {
		c.Next(NewHttpError(fiber.StatusBadRequest, "Bad request", fiber.Map{
			"message": "Hello World",
		}))
	})

	var _statusCode int

	for n := 0; n < b.N; n++ {
		req := httptest.NewRequest("GET", "/hello", nil)
		if resp, err := app.Test(req); err != nil {
			assert.NoError(b, err)
		} else {
			assert.Equal(b, fiber.StatusOK, resp.StatusCode)
			_statusCode = resp.StatusCode
		}
	}
	_benchmark_errhandler_json_err = _statusCode
}