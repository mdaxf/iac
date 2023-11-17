package checks

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // import mysql driver
)

type MySQLCheck struct {
	Ctx              context.Context
	ConnectionString string
	DB               *sql.DB
	Error            error
}

func NewMySQLCheck(ctx context.Context, db *sql.DB, connectionString string) MySQLCheck {
	return MySQLCheck{
		Ctx:              ctx,
		ConnectionString: connectionString,
		DB:               db,
		Error:            nil,
	}
}

func CheckMySQLStatus(ctx context.Context, db *sql.DB, connectionString string) error {
	check := NewMySQLCheck(ctx, db, connectionString)
	return check.CheckStatus()
}

func (check *MySQLCheck) CheckbyConnectionString() error {

	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	var checkErr error
	checkErr = nil

	ctx := check.Ctx
	db, err := sql.Open("mysql", check.ConnectionString)
	if err != nil {
		checkErr = fmt.Errorf("MySQL health check failed on connect: %w", err)
		check.Error = checkErr
		return checkErr
	}
	defer func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()
		// override checkErr only if there were no other errors
		if cerr := db.Close(); cerr != nil && checkErr == nil {
			checkErr = fmt.Errorf("MySQL health check failed on connection closing: %w", cerr)
			check.Error = checkErr
			return
		}
	}()

	if err = db.PingContext(ctx); err != nil {
		checkErr = fmt.Errorf("MySQL health check failed on ping: %w", err)
		check.Error = checkErr
		return checkErr
	}

	rows, err := db.QueryContext(ctx, `SELECT VERSION()`)
	if err != nil {
		checkErr = fmt.Errorf("MySQL health check failed on select: %w", err)
		check.Error = checkErr
		return checkErr
	}
	defer func() {
		// override checkErr only if there were no other errors
		if err = rows.Close(); err != nil && checkErr == nil {
			checkErr = fmt.Errorf("MySQL health check failed on rows closing: %w", err)
			check.Error = checkErr
			return
		}
	}()

	return nil

}

func (check *MySQLCheck) CheckStatus() error {
	defer func() {
		if err := recover(); err != nil {
			check.Error = fmt.Errorf("MySQL health check failed on connect: %w", err)
			return
		}
	}()

	ctx := check.Ctx
	db := check.DB
	err := db.PingContext(ctx)
	if err != nil {
		check.Error = fmt.Errorf("MySQL health check failed on ping: %w", err)
		return check.Error
	}
	rows, err := db.QueryContext(ctx, `SELECT VERSION()`)
	defer rows.Close()

	if err != nil {
		check.Error = fmt.Errorf("MySQL health check failed on select: %w", err)
		return check.Error
	}

	err = rows.Close()
	if err != nil {
		check.Error = fmt.Errorf("MySQL health check failed on rows closing: %w", err)
		return check.Error
	}

	return nil
}
