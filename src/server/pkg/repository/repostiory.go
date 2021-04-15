package repository

import (
	"traefik-tryout/src/server/pkg/models"
	"traefik-tryout/src/utils"

	"github.com/opentracing/opentracing-go"
)

type Repository struct {
	repository []models.Customer
	tracer     opentracing.Span
}

func NewRepository(tracer opentracing.Span) *Repository {
	return &Repository{
		repository: make([]models.Customer, 0),
		tracer: opentracing.StartSpan(
			"repository",
			opentracing.ChildOf(tracer.Context()),
		),
	}
}

func (r *Repository) WithTracer(tracer opentracing.Span) *Repository {
	r.tracer = opentracing.StartSpan(
		"repository",
		opentracing.ChildOf(tracer.Context()),
	)
	return r
}

func (r Repository) AddCustomer(c models.Customer) {
	defer utils.WrapTrace(
		r.tracer,
		utils.WithTag("method", "AddCustomer"),
	)
	r.repository = append(r.repository, c)
}

func (r Repository) GetCustomers() []models.Customer {
	defer utils.WrapTrace(
		r.tracer,
		utils.WithTag("method", "GetCustomers"),
	)
	return r.repository
}
