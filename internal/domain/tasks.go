package domain

import "github.com/google/uuid"

type EmailVerificationTask struct {
	Email string `json:"email"`
}

// type TransferTask struct {
// 	EmailFrom string    `json:"emailFrom"`
// 	EmailTo   string    `json:"emailTo"`
// 	AccIdFrom uuid.UUID `json:"accIdFrom"`
// 	AccIdTo   uuid.UUID `json:"accIdTo"`
// 	Amount    int       `json:"amount"`
// }

type CashoutTask struct {
	MachineId uuid.UUID `json:"machineId"`
	Email     string    `json:"email"`
	AccId     uuid.UUID `json:"accId"`
	Amount    int       `json:"amount"`
	NewMoney  int       `json:"new_money"`
}

type DepositTask struct {
	MachineId uuid.UUID `json:"machineId"`
	Email     string    `json:"email"`
	AccId     uuid.UUID `json:"accId"`
	Amount    int       `json:"amount"`
	NewMoney  int       `json:"new_money"`
}
