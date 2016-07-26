package model

// Automatically generated by genmodel utility. DO NOT EDIT!

import (
	"encoding/json"
	"github.com/builderscon/octav/octav/tools"
	"time"

	"github.com/builderscon/octav/octav/db"
	"github.com/lestrrat/go-pdebug"
)

var _ = time.Time{}

type rawSponsor struct {
	ID           string `json:"id"`
	ConferenceID string `json:"conference_id"`
	Name         string `json:"name" l10n:"true"`
	LogoURL1     string `json:"logo_url1,omitempty"`
	LogoURL2     string `json:"logo_url2,omitempty"`
	LogoURL3     string `json:"logo_url3,omitempty"`
	URL          string `json:"url"`
	GroupName    string `json:"group_name"`
	SortOrder    int    `json:"sort_order"`
}

func (v Sponsor) MarshalJSON() ([]byte, error) {
	var raw rawSponsor
	raw.ID = v.ID
	raw.ConferenceID = v.ConferenceID
	raw.Name = v.Name
	raw.LogoURL1 = v.LogoURL1
	raw.LogoURL2 = v.LogoURL2
	raw.LogoURL3 = v.LogoURL3
	raw.URL = v.URL
	raw.GroupName = v.GroupName
	raw.SortOrder = v.SortOrder
	buf, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	return tools.MarshalJSONWithL10N(buf, v.LocalizedFields)
}

func (v *Sponsor) Load(tx *db.Tx, id string) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("model.Sponsor.Load %s", id).BindError(&err)
		defer g.End()
	}
	vdb := db.Sponsor{}
	if err := vdb.LoadByEID(tx, id); err != nil {
		return err
	}

	if err := v.FromRow(vdb); err != nil {
		return err
	}
	return nil
}

func (v *Sponsor) FromRow(vdb db.Sponsor) error {
	v.ID = vdb.EID
	v.ConferenceID = vdb.ConferenceID
	v.Name = vdb.Name
	if vdb.LogoURL1.Valid {
		v.LogoURL1 = vdb.LogoURL1.String
	}
	if vdb.LogoURL2.Valid {
		v.LogoURL2 = vdb.LogoURL2.String
	}
	if vdb.LogoURL3.Valid {
		v.LogoURL3 = vdb.LogoURL3.String
	}
	v.URL = vdb.URL
	v.GroupName = vdb.GroupName
	v.SortOrder = vdb.SortOrder
	return nil
}

func (v *Sponsor) ToRow(vdb *db.Sponsor) error {
	vdb.EID = v.ID
	vdb.ConferenceID = v.ConferenceID
	vdb.Name = v.Name
	vdb.LogoURL1.Valid = true
	vdb.LogoURL1.String = v.LogoURL1
	vdb.LogoURL2.Valid = true
	vdb.LogoURL2.String = v.LogoURL2
	vdb.LogoURL3.Valid = true
	vdb.LogoURL3.String = v.LogoURL3
	vdb.URL = v.URL
	vdb.GroupName = v.GroupName
	vdb.SortOrder = v.SortOrder
	return nil
}
