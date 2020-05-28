package main

import (
	"errors"
	"github.com/gofiber/fiber"
	"github.com/gofiber/template/html"
	errhandler "github.com/hendratommy/fiber-errhandler"
)

var data = fiber.Map{
	"firstName": "John",
	"lastName":  "Doe",
}

func main() {
	app := fiber.New()
	app.Settings.Templates = html.New("./views", ".html")

	app.Use(errhandler.New(errhandler.Config{
		UseTemplate: true,
		//Log: true,
		Handler: func(c *fiber.Ctx, err error, fallback func(...interface{})) {
			if he, ok := err.(errhandler.HTTPError); ok {
				switch he.StatusCode() {
				case fiber.StatusUnauthorized:
					c.Status(he.StatusCode()).Render("custom-err", fiber.Map{
						"StatusCode": he.StatusCode(),
						"Message":    he.Message(),
						"Reason":     "Please login first",
						"SomeData":   he.Data(),
					})
					return

				default:
					break
				}
			}
			fallback(err)
		},
	}))

	apiv1 := app.Group("/api")
	apiv1.Get("/", func(c *fiber.Ctx) {
		c.JSON(data)
	})
	apiv1.Get("/panic", func(c *fiber.Ctx) {
		panic(errors.New("winter is coming to the api"))
	})
	apiv1.Get("/err-403", func(c *fiber.Ctx) {
		c.Next(errhandler.NewHttpError(fiber.StatusForbidden, "Cannot access this", nil))
	})

	app.Get("/", func(c *fiber.Ctx) {
		c.Render("index", data)
	})
	app.Get("/panic", func(c *fiber.Ctx) {
		panic(errors.New("winter is coming to the web"))
	})
	app.Get("/err-403", func(c *fiber.Ctx) {
		c.Next(errhandler.NewHttpError(fiber.StatusForbidden, "Cannot access this", nil))
	})
	app.Get("/custom", func(c *fiber.Ctx) {
		c.Next(errhandler.NewHttpError(fiber.StatusUnauthorized, "unauthorized", struct {
			FirstName string
			LastName  string
		}{
			FirstName: "John",
			LastName:  "Wick",
		}))
	})

	app.Listen(3000)
}
