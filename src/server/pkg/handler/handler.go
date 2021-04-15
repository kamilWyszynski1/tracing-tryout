package handler

import (
	"traefik-tryout/src/server/pkg/models"
	"traefik-tryout/src/server/pkg/service"
	"traefik-tryout/src/utils"

	"github.com/opentracing/opentracing-go"
)

type ServerI interface {
	CreateCustomer(models.Customer)
	GetCustomers() []models.Customer
}

type Handler struct {
	s      *service.Service
	tracer opentracing.Span
}

func NewHandler(tracer opentracing.Span, s *service.Service) *Handler {
	return &Handler{
		s: s,
		tracer: opentracing.StartSpan(
			"repository",
			opentracing.ChildOf(tracer.Context()),
		),
	}
}

func (h *Handler) Tracer() opentracing.Span {
	return h.tracer
}

func (h *Handler) WithTracer(tracer opentracing.Span) *Handler {
	h.tracer = opentracing.StartSpan(
		"handler",
		opentracing.ChildOf(tracer.Context()),
	)
	return h
}

func (h Handler) CreateCustomer(customer models.Customer) {
	defer utils.WrapTrace(
		h.tracer,
		utils.WithTag("method", "CreateCustomer"),
		utils.WithLog("name", customer.Name),
	)
	h.s.WithTracer(h.tracer).AddCustomer(customer)
}

func (h Handler) GetCustomers() []models.Customer {
	defer utils.WrapTrace(
		h.tracer,
		utils.WithTag("method", "GetCustomers"),
	)
	return h.s.WithTracer(h.tracer).GetCustomers()
}
