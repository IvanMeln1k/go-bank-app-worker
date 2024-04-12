package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/IvanMeln1k/go-bank-app-worker/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type MachinesRepository struct {
	db *sqlx.DB
}

func NewMachinesRepository(db *sqlx.DB) *MachinesRepository {
	return &MachinesRepository{
		db: db,
	}
}

func (r *MachinesRepository) GetStats(ctx context.Context, id uuid.UUID) (domain.MachineStats, error) {
	var machineStats domain.MachineStats

	query := fmt.Sprintf("SELECT * FROM %s WHERE machine_id=$1", machinesStatsTable)
	if err := r.db.Get(&machineStats, query, id); err != nil {
		logrus.Errorf("error getting machine stats from db: %s", err)
		if errors.Is(sql.ErrNoRows, err) {
			return machineStats, ErrMachinesStatsNotFound
		}
		return machineStats, ErrInternal
	}

	return machineStats, nil
}

func (r *MachinesRepository) UpdateStats(ctx context.Context, id uuid.UUID,
	data domain.MachineStatsUpdate) (domain.MachineStats, error) {
	var machineStats domain.MachineStats

	values := make([]interface{}, 0)
	fields := make([]string, 0)
	argsId := 1

	addField := func(field string, value interface{}) {
		values = append(values, value)
		fields = append(fields, fmt.Sprintf("%s=$%d", field, argsId))
		argsId++
	}

	if data.Cashout != nil {
		addField("cash_out", *data.Cashout)
	}

	if data.Deposit != nil {
		addField("deposit", *data.Deposit)
	}

	setQuery := strings.Join(fields, ", ")
	query := fmt.Sprintf("UPDATE %s m SET %s WHERE machine_id=$%d RETURNING m.*", machinesStatsTable, setQuery,
		argsId)
	values = append(values, id)
	row := r.db.QueryRowx(query, values...)
	if err := row.StructScan(&machineStats); err != nil {
		logrus.Errorf("error updating machine stats into db: %s", err)
		return machineStats, ErrInternal
	}

	return machineStats, nil
}
