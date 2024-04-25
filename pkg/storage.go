/*
		GNU GENERAL PUBLIC LICENSE
		Version 3, 29 June 2007


kyoketsu, a Client-To-Client Network Enumeration System
Copyright (C) 2024 Russell Hrubesky, ChiralWorks Software LLC

 Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License,
or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/

package kyoketsu

import (
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"
)

type TopologyDatabaseIO interface {
	/*
			This interface defines the Input and output methods that will be necessary
		    for an appropriate implementation of the data storage that the distributed system will use.
		    When I get around to implementing the client-to-client format of this, it could be anything.
	*/
	Migrate() error
	Create(host Host) (*Host, error)
	All() ([]Host, error)
	GetByIP(ip string) (*Host, error)
	Update(id int64, updated Host) (*Host, error)
	Delete(id int64) error
}

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type SQLiteRepo struct {
	db *sql.DB
}

// Instantiate a new SQLiteRepo struct
func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{
		db: db,
	}

}

// Creates a new SQL table with necessary data
func (r *SQLiteRepo) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS hosts(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        fqdn TEXT NOT NULL,
        ipv4_address TEXT NOT NULL UNIQUE,
        listening_port TEXT NOT NULL
    );
    `

	_, err := r.db.Exec(query)
	return err
}

/*
Create an entry in the hosts table

	:param host: a Host entry from a port scan
*/
func (r *SQLiteRepo) Create(host Host) (*Host, error) {
	res, err := r.db.Exec("INSERT INTO hosts(fqdn, ipv4_address, listening_port) values(?,?,?)", host.Fqdn, host.IpAddress, host.PortString)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, ErrDuplicate
			}
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	host.Id = id

	return &host, nil
}

// Get all Hosts from the host table
func (r *SQLiteRepo) All() ([]Host, error) {
	rows, err := r.db.Query("SELECT * FROM hosts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Host
	for rows.Next() {
		var host Host
		if err := rows.Scan(&host.Id, &host.Fqdn, &host.IpAddress, &host.PortString); err != nil {
			return nil, err
		}
		all = append(all, host)
	}
	return all, nil
}

// Get a record by its FQDN
func (r *SQLiteRepo) GetByIP(ip string) (*Host, error) {
	row := r.db.QueryRow("SELECT * FROM hosts WHERE ipv4_address = ?", ip)

	var host Host
	if err := row.Scan(&host.Id, &host.Fqdn, &host.IpAddress, &host.PortString); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &host, nil
}

// Update a record by its ID
func (r *SQLiteRepo) Update(id int64, updated Host) (*Host, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec("UPDATE hosts SET fqdn = ?, ipv4_address = ?, listening_port = ? WHERE id = ?", updated.Fqdn, updated.IpAddress, updated.PortString, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}

// Delete a record by its ID
func (r *SQLiteRepo) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM hosts WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
}
