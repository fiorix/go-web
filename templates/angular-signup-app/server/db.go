// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import "database/sql"

type User struct {
	Id       int
	Email    string
	Passwd   string
	FullName sql.NullString
	IsActive bool
}

func NewUser(email, passwd, fullname string, active bool) (*User, error) {
	stmt, err := MySQL.Prepare(`
		insert into User (Email, Passwd, FullName, IsActive)
		values (?, SHA1(?), ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(email, passwd, fullname, active); err != nil {
		return nil, err
	}
	return GetUser(email)
}

func UserExists(email string) (bool, error) {
	stmt, err := MySQL.Prepare("select count(*) from User where Email=?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var count int
	if err = stmt.QueryRow(email).Scan(&count); err != nil {
		return false, err
	}
	return count >= 1, nil

}

// TODO: cache
func GetUser(email string) (*User, error) {
	stmt, err := MySQL.Prepare("select * from User where Email=?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var u User
	if err = stmt.QueryRow(email).Scan(&u.Id, &u.Email, &u.Passwd,
		&u.FullName, &u.IsActive); err != nil {
		return nil, err
	}
	return &u, nil
}

// TODO: cache
func GetUserById(id int) (*User, error) {
	stmt, err := MySQL.Prepare("select * from User where Id=?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var u User
	if err = stmt.QueryRow(id).Scan(&u.Id, &u.Email, &u.Passwd,
		&u.FullName, &u.IsActive); err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserWithPasswd(email, passwd string) (*User, error) {
	stmt, err := MySQL.Prepare(`
		select * from User where Email=? and Passwd=SHA1(?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var u User
	if err = stmt.QueryRow(email, passwd).Scan(&u.Id, &u.Email, &u.Passwd,
		&u.FullName, &u.IsActive); err != nil {
		return nil, err
	}
	return &u, nil
}

func DelUser(u *User) error {
	stmt, err := MySQL.Prepare("delete from Users where Id=?")
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(u.Id); err != nil {
		return err
	}
	return nil
}

func UpdateUser(u *User) error {
	// TODO: Do something better than this.
	stmt, err := MySQL.Prepare(`
		update User
			set Passwd=?, FullName=?, IsActive=?
			where Id=?
		`)
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(u.Passwd, u.FullName.String, u.IsActive, u.Id); err != nil {
		return err
	}
	return nil
}
