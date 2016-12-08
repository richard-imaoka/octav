package model

// Automatically generated by genmodel utility. DO NOT EDIT!

import (
	"time"

	"github.com/builderscon/octav/octav/db"
	"github.com/lestrrat/go-pdebug"
)

var _ = pdebug.Enabled
var _ = time.Time{}

func (v *ConferenceDate) Load(tx *db.Tx, id string) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("model.ConferenceDate.Load %s", id).BindError(&err)
		defer g.End()
	}
	vdb := db.ConferenceDate{}
	if err := vdb.LoadByEID(tx, id); err != nil {
		return err
	}

	if err := v.FromRow(&vdb); err != nil {
		return err
	}
	return nil
}

func (v *ConferenceDate) FromRow(vdb *db.ConferenceDate) error {
	v.ID = vdb.EID
	if vdb.Open.Valid {
		v.Open = vdb.Open.Time
	}
	if vdb.Close.Valid {
		v.Close = vdb.Close.Time
	}
	return nil
}

func (v *ConferenceDate) ToRow(vdb *db.ConferenceDate) error {
	vdb.EID = v.ID
	vdb.Open.Valid = true
	vdb.Open.Time = v.Open
	vdb.Close.Valid = true
	vdb.Close.Time = v.Close
	return nil
}