package model

import (
	"fmt"
)

type Key struct {
	tableName string
	ID        int64  `gorm:"primary_key;not null"`
	PublicKey []byte `gorm:"column:pub;not null"`
	Amount    string `gorm:"not null"`
	RbID      int64  `gorm:"not null"`
}

func (m *Key) SetTablePrefix(prefix int64) *Key {
	if prefix == 0 {
		prefix = 1
	}
	m.tableName = fmt.Sprintf("%d_keys", prefix)
	return m
}

func (m Key) TableName() string {
	return m.tableName
}

func (m *Key) Get(wallet int64) error {
	return DBConn.Where("id = ?", wallet).First(m).Error
}
