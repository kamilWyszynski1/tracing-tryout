package service

import (
	"traefik-tryout/src/server/pkg/models"
	"traefik-tryout/src/server/pkg/repository"
	"traefik-tryout/src/utils"

	"github.com/opentracing/opentracing-go"
)

type Service struct {
	r      *repository.Repository
	tracer opentracing.Span
}

func NewService(tracer opentracing.Span, r *repository.Repository) *Service {
	return &Service{
		r: r,
		tracer: opentracing.StartSpan(
			"repository",
			opentracing.ChildOf(tracer.Context()),
		),
	}
}

func (s *Service) WithTracer(tracer opentracing.Span) *Service {
	s.tracer = opentracing.StartSpan(
		"service",
		opentracing.ChildOf(tracer.Context()),
	)
	return s
}

func (s Service) AddCustomer(c models.Customer) {
	defer utils.WrapTrace(
		s.tracer,
		utils.WithTag("method", "AddCustomer"),
	)
	s.r.WithTracer(s.tracer).AddCustomer(c)
}

func (s Service) GetCustomers() []models.Customer {
	defer utils.WrapTrace(
		s.tracer,
		utils.WithTag("method", "GetCustomers"),
	)
	return s.r.WithTracer(s.tracer).GetCustomers()
}
