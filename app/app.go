package app

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/abiiranathan/gowrap/orm"
	"github.com/abiiranathan/gowrap/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

var json = jsoniter.ConfigFastest

type Handler struct {
	ORM       orm.ORM
	App       *fiber.App
	validator validation.Validator
}

type HandlerOption func(*Handler)

func ValidationTagName(tagName string) HandlerOption {
	return func(h *Handler) {
		if h.validator != nil {
			h.validator.SetTagName(tagName)
		}
	}
}

func NewApplication(db *gorm.DB, options ...HandlerOption) *Handler {
	app := fiber.New(fiber.Config{
		ServerHeader:  "ECLINIC HMS",
		Prefork:       true,
		StrictRouting: false,
		CaseSensitive: true,
		AppName:       "eclinicgo",
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
	})

	app.Use(recover.New())
	app.Use(etag.New(etag.Config{
		Weak: true,
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ALLOWED_ORIGINS"),
		AllowMethods:     "GET, POST, PATCH, PUT, DELETE, HEAD, OPTIONS",
		AllowHeaders:     "Authorization, Content-Type, Content-Length, Accept, Origin, Accept-Encoding, Cache-Control, X-Requested-With",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
		MaxAge:           int((7 * 24 * time.Hour).Seconds()),
	}))

	app.Use(logger.New(logger.Config{
		Output: os.Stdout,
	}))

	// Struct validator
	validate := validation.NewValidator("binding")

	h := &Handler{
		ORM:       orm.New(db),
		App:       app,
		validator: validate,
	}

	for _, option := range options {
		option(h)
	}
	return h
}

func (h *Handler) AbortWithError(c *fiber.Ctx, status int, err string) {
	c.Status(status).JSON(fiber.Map{"error": err})
}

// Responds with status and custom error string err
func (h *Handler) AbortWithMessage(c *fiber.Ctx, status int, message string) {
	c.Status(status).JSON(fiber.Map{"error": message})
}

// If err is gorm.ErrRecordNotFound, it sends a 404
// otherwise sends a 500 status code
func (h *Handler) RespondWithError(c *fiber.Ctx, err error) {
	data := fiber.Map{"error": err.Error()}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(http.StatusNotFound).JSON(data)
		return
	}
	c.Status(http.StatusInternalServerError).JSON(data)
}

// Returns a parameter as a uint
func (h *Handler) QueryInt(ctx *fiber.Ctx, name string) (uint, error) {
	value := ctx.Query(name, "0")
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return uint(intValue), nil
}

// Sends data as JSON with status 200
func (h *Handler) JSON(ctx *fiber.Ctx, data any) {
	ctx.Status(http.StatusOK).JSON(data)
}

// Sends data as json with status 400
func (h *Handler) BadRequest(ctx *fiber.Ctx, data any) {
	ctx.Status(http.StatusBadRequest).JSON(data)
}

// Sends data as json with status 500
func (h *Handler) Error500(ctx *fiber.Ctx, data any) {
	ctx.Status(http.StatusInternalServerError).JSON(data)
}

// Parse JSON from request body and validate it with the validator package
func (h *Handler) BindJSON(ctx *fiber.Ctx, obj any) (valid bool) {
	if err := ctx.BodyParser(obj); err != nil {
		h.AbortWithMessage(ctx, fiber.StatusBadRequest, "invalid request body")
		return
	}

	// if the struct is not valid sent the error and return false
	if err := h.validator.Validate(obj); err != nil {
		h.SendError(ctx, err)
		return
	}
	return true
}
