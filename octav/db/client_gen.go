package db

// Automatically generated by gendb utility. DO NOT EDIT!

import (
	"bytes"
	"database/sql"
	"strconv"
	"time"

	"github.com/builderscon/octav/octav/tools"
	"github.com/lestrrat/go-pdebug"
	"github.com/pkg/errors"
)

const ClientStdSelectColumns = "clients.oid, clients.eid, clients.secret, clients.name, clients.created_on, clients.modified_on"
const ClientTable = "clients"

type ClientList []Client

func (c *Client) Scan(scanner interface {
	Scan(...interface{}) error
}) error {
	return scanner.Scan(&c.OID, &c.EID, &c.Secret, &c.Name, &c.CreatedOn, &c.ModifiedOn)
}

func init() {
	hooks = append(hooks, func() {
		stmt := tools.GetBuffer()
		defer tools.ReleaseBuffer(stmt)

		stmt.Reset()
		stmt.WriteString(`DELETE FROM `)
		stmt.WriteString(ClientTable)
		stmt.WriteString(` WHERE oid = ?`)
		library.Register("sqlClientDeleteByOIDKey", stmt.String())

		stmt.Reset()
		stmt.WriteString(`UPDATE `)
		stmt.WriteString(ClientTable)
		stmt.WriteString(` SET eid = ?, secret = ?, name = ? WHERE oid = ?`)
		library.Register("sqlClientUpdateByOIDKey", stmt.String())

		stmt.Reset()
		stmt.WriteString(`SELECT `)
		stmt.WriteString(ClientStdSelectColumns)
		stmt.WriteString(` FROM `)
		stmt.WriteString(ClientTable)
		stmt.WriteString(` WHERE `)
		stmt.WriteString(ClientTable)
		stmt.WriteString(`.eid = ?`)
		library.Register("sqlClientLoadByEIDKey", stmt.String())

		stmt.Reset()
		stmt.WriteString(`DELETE FROM `)
		stmt.WriteString(ClientTable)
		stmt.WriteString(` WHERE eid = ?`)
		library.Register("sqlClientDeleteByEIDKey", stmt.String())

		stmt.Reset()
		stmt.WriteString(`UPDATE `)
		stmt.WriteString(ClientTable)
		stmt.WriteString(` SET eid = ?, secret = ?, name = ? WHERE eid = ?`)
		library.Register("sqlClientUpdateByEIDKey", stmt.String())
	})
}

func (c *Client) LoadByEID(tx *Tx, eid string) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker(`Client.LoadByEID %s`, eid).BindError(&err)
		defer g.End()
	}
	stmt, err := library.GetStmt("sqlClientLoadByEIDKey")
	if err != nil {
		return errors.Wrap(err, `failed to get statement`)
	}
	row := tx.Stmt(stmt).QueryRow(eid)
	if err := c.Scan(row); err != nil {
		return err
	}
	return nil
}

func (c *Client) Create(tx *Tx, opts ...InsertOption) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("db.Client.Create").BindError(&err)
		defer g.End()
		pdebug.Printf("%#v", c)
	}
	if c.EID == "" {
		return errors.New("create: non-empty EID required")
	}

	c.CreatedOn = time.Now()
	doIgnore := false
	for _, opt := range opts {
		switch opt.(type) {
		case insertIgnoreOption:
			doIgnore = true
		}
	}

	stmt := bytes.Buffer{}
	stmt.WriteString("INSERT ")
	if doIgnore {
		stmt.WriteString("IGNORE ")
	}
	stmt.WriteString("INTO ")
	stmt.WriteString(ClientTable)
	stmt.WriteString(` (eid, secret, name, created_on, modified_on) VALUES (?, ?, ?, ?, ?)`)
	result, err := tx.Exec(stmt.String(), c.EID, c.Secret, c.Name, c.CreatedOn, c.ModifiedOn)
	if err != nil {
		return err
	}

	lii, err := result.LastInsertId()
	if err != nil {
		return err
	}

	c.OID = lii
	return nil
}

func (c Client) Update(tx *Tx) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker(`Client.Update`).BindError(&err)
		defer g.End()
	}
	if c.OID != 0 {
		if pdebug.Enabled {
			pdebug.Printf(`Using OID (%d) as key`, c.OID)
		}
		stmt, err := library.GetStmt("sqlClientUpdateByOIDKey")
		if err != nil {
			return errors.Wrap(err, `failed to get statement`)
		}
		_, err = tx.Stmt(stmt).Exec(c.EID, c.Secret, c.Name, c.OID)
		return err
	}
	if c.EID != "" {
		if pdebug.Enabled {
			pdebug.Printf(`Using EID (%s) as key`, c.EID)
		}
		stmt, err := library.GetStmt("sqlClientUpdateByEIDKey")
		if err != nil {
			return errors.Wrap(err, `failed to get statement`)
		}
		_, err = tx.Stmt(stmt).Exec(c.EID, c.Secret, c.Name, c.EID)
		return err
	}
	return errors.New("either OID/EID must be filled")
}

func (c Client) Delete(tx *Tx) error {
	if c.OID != 0 {
		stmt, err := library.GetStmt("sqlClientDeleteByOIDKey")
		if err != nil {
			return errors.Wrap(err, `failed to get statement`)
		}
		_, err = tx.Stmt(stmt).Exec(c.OID)
		return err
	}

	if c.EID != "" {
		stmt, err := library.GetStmt("sqlClientDeleteByEIDKey")
		if err != nil {
			return errors.Wrap(err, `failed to get statement`)
		}
		_, err = tx.Stmt(stmt).Exec(c.EID)
		return err
	}

	return errors.New("either OID/EID must be filled")
}

func (v *ClientList) FromRows(rows *sql.Rows, capacity int) error {
	var res []Client
	if capacity > 0 {
		res = make([]Client, 0, capacity)
	} else {
		res = []Client{}
	}

	for rows.Next() {
		vdb := Client{}
		if err := vdb.Scan(rows); err != nil {
			return err
		}
		res = append(res, vdb)
	}
	*v = res
	return nil
}

func (v *ClientList) LoadSinceEID(tx *Tx, since string, limit int) error {
	var s int64
	if id := since; id != "" {
		vdb := Client{}
		if err := vdb.LoadByEID(tx, id); err != nil {
			return err
		}

		s = vdb.OID
	}
	return v.LoadSince(tx, s, limit)
}

func (v *ClientList) LoadSince(tx *Tx, since int64, limit int) error {
	rows, err := tx.Query(`SELECT `+ClientStdSelectColumns+` FROM `+ClientTable+` WHERE clients.oid > ? ORDER BY oid ASC LIMIT `+strconv.Itoa(limit), since)
	if err != nil {
		return err
	}

	if err := v.FromRows(rows, limit); err != nil {
		return err
	}
	return nil
}
