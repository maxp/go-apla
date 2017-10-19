package model

type MigrationHistory struct {
	Version     string `gorm:"not null"`
	DateApplied int64  `gorm:"not null"`
}

func (mh *MigrationHistory) TableName() string {
	return "migration_history"
}

func (mh *MigrationHistory) Get() error {
	return DBConn.First(mh).Error
}

func (mh *MigrationHistory) Create() error {
	return DBConn.Create(mh).Error
}

func (mh *MigrationHistory) Save() error {
	return DBConn.Save(mh).Error
}
