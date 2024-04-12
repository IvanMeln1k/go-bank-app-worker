package repository

import (
	"context"
	"errors"

	"github.com/IvanMeln1k/go-bank-app-worker/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	machinesStatsTable = "machines_stats"
)

var (
	ErrMachinesStatsNotFound = errors.New("machines stats not found")
	ErrInternal              = errors.New("error internal")
)

type Machines interface {
	GetStats(ctx context.Context, id uuid.UUID) (domain.MachineStats, error)
	UpdateStats(ctx context.Context, id uuid.UUID,
		data domain.MachineStatsUpdate) (domain.MachineStats, error)
}

type Repository struct {
	Machines
}

type Deps struct {
	DB *sqlx.DB
}

func NewRepository(deps Deps) *Repository {
	return &Repository{
		Machines: NewMachinesRepository(deps.DB),
	}
}
