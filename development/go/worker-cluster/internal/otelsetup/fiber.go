package otelsetup

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func FiberMiddleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.UserContext(), c.Method()+" "+c.Path())
		defer span.End()

		c.SetUserContext(ctx)
		err := c.Next()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.SetAttributes(
			attribute.Int("http.status_code", c.Response().StatusCode()),
			attribute.String("http.route", c.Route().Path),
		)
		return err
	}
}
