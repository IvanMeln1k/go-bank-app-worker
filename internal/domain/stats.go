package domain

import "github.com/google/uuid"

type MachineStats struct {
	MachineId uuid.UUID `db:"machine_id"`
	Cashout   int       `db:"cash_out"`
	Deposit   int       `db:"deposit"`
}

type MachineStatsUpdate struct {
	Cashout *int
	Deposit *int
}
