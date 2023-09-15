package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/logotipiwe/dc_go_utils/src/config"
)

var db *sql.DB

func InitDb() error {
	connectionStr := fmt.Sprintf("%v:%v@tcp(%v)/%v", GetConfig("DB_USER"), GetConfig("DB_PASS"),
		GetConfig("DB_HOST"), GetConfig("DB_NAME")) //TODO make idp's own config
	conn, err := sql.Open("mysql", connectionStr)
	if err != nil {
		return err
	}
	if err := conn.Ping(); err != nil {
		println(fmt.Sprintf("Error connecting database: %s", err))
		return err
	}
	db = conn
	println("Database connected!")
	return nil
}

func existsInDbByGoogleId(googleId string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT IF(COUNT(*),'true','false') from users where google_id = ?", googleId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func createUserInDb(user *DcUser) (*DcUser, error) {
	_, err := db.Exec("INSERT INTO users (id, name, picture, google_id) VALUES (?,?,?,?)",
		user.Id, user.Name, user.Picture, user.GoogleId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func getUserFromDbByGoogleId(gId string) (*DcUser, error) {
	user := DcUser{}
	row := db.QueryRow("SELECT id, name, picture, google_id FROM users WHERE google_id = ?", gId)
	err := row.Scan(&user.Id, &user.Name, &user.Picture, &user.GoogleId)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func getUserFromDbById(id string) (*DcUser, error) {
	user := DcUser{}
	row := db.QueryRow("SELECT id, name, picture, google_id FROM users WHERE id = ?", id)
	err := row.Scan(&user.Id, &user.Name, &user.Picture, &user.GoogleId)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
