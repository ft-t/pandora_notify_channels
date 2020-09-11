package db

import (
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

var migrations = []*migrate.Migration{
	{
		ID: "initial_202009111829",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync2(&TgChat{}); err != nil {
				return err
			}

			if err := tx.Sync2(&PandoraDevice{}); err != nil {
				return err
			}

			return tx.Sync2(&PandoraNotification{})
		},
	},
}
