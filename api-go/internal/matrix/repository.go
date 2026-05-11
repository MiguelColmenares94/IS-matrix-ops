package matrix

import (
	"database/sql"
	"encoding/json"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) SaveQRComputation(userID string, input, q, r [][]float64, success bool, errMsg string) error {
	inputJSON, _ := json.Marshal(input)
	var qJSON, rJSON interface{}
	if q != nil {
		b, _ := json.Marshal(q)
		qJSON = b
		b2, _ := json.Marshal(r)
		rJSON = b2
	}
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	_, err := repo.db.Exec(
		"CALL save_qr_computation($1, $2, $3, $4, $5, $6)",
		userID, inputJSON, qJSON, rJSON, success, errPtr,
	)
	return err
}
