package service

// Automatically generated by genmodel utility. DO NOT EDIT!

import (
	"context"
	"sync"
	"time"

	"github.com/builderscon/octav/octav/cache"

	"github.com/builderscon/octav/octav/db"
	"github.com/builderscon/octav/octav/internal/errors"
	"github.com/builderscon/octav/octav/model"
	"github.com/lestrrat/go-pdebug"
)

var _ = time.Time{}
var _ = cache.WithExpires(time.Minute)
var _ = context.Background
var _ = errors.Wrap
var _ = model.Venue{}
var _ = db.Venue{}
var _ = pdebug.Enabled

var venueSvc VenueSvc
var venueOnce sync.Once

func Venue() *VenueSvc {
	venueOnce.Do(venueSvc.Init)
	return &venueSvc
}

func (v *VenueSvc) LookupFromPayload(tx *db.Tx, m *model.Venue, payload *model.LookupVenueRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.Venue.LookupFromPayload").BindError(&err)
		defer g.End()
	}
	if err = v.Lookup(tx, m, payload.ID); err != nil {
		return errors.Wrap(err, "failed to load model.Venue from database")
	}
	if err := v.Decorate(tx, m, payload.TrustedCall, payload.Lang.String); err != nil {
		return errors.Wrap(err, "failed to load associated data for model.Venue from database")
	}
	return nil
}

func (v *VenueSvc) Lookup(tx *db.Tx, m *model.Venue, id string) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.Venue.Lookup").BindError(&err)
		defer g.End()
	}

	var r model.Venue
	key := `api.Venue.` + id
	c := Cache()
	var cacheMiss bool
	_, err = c.GetOrSet(key, &r, func() (interface{}, error) {
		if pdebug.Enabled {
			cacheMiss = true
		}
		if err := r.Load(tx, id); err != nil {
			return nil, errors.Wrap(err, "failed to load model.Venue from database")
		}
		return &r, nil
	}, cache.WithExpires(time.Hour))
	if pdebug.Enabled {
		cacheSt := `HIT`
		if cacheMiss {
			cacheSt = `MISS`
		}
		pdebug.Printf(`CACHE %s: %s`, cacheSt, key)
	}
	*m = r
	return nil
}

// Create takes in the transaction, the incoming payload, and a reference to
// a database row. The database row is initialized/populated so that the
// caller can use it afterwards.
func (v *VenueSvc) Create(tx *db.Tx, vdb *db.Venue, payload *model.CreateVenueRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.Venue.Create").BindError(&err)
		defer g.End()
	}

	if err := v.populateRowForCreate(vdb, payload); err != nil {
		return errors.Wrap(err, `failed to populate row`)
	}

	if err := vdb.Create(tx); err != nil {
		return errors.Wrap(err, `failed to insert into database`)
	}

	if err := payload.LocalizedFields.CreateLocalizedStrings(tx, "Venue", vdb.EID); err != nil {
		return errors.Wrap(err, `failed to populate localized strings`)
	}
	return nil
}

func (v *VenueSvc) Update(tx *db.Tx, vdb *db.Venue) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.Venue.Update (%s)", vdb.EID).BindError(&err)
		defer g.End()
	}

	if vdb.EID == `` {
		return errors.New("vdb.EID is required (did you forget to call vdb.Load(tx) before hand?)")
	}

	if err := vdb.Update(tx); err != nil {
		return errors.Wrap(err, `failed to update database`)
	}
	key := `api.Venue.` + vdb.EID
	if pdebug.Enabled {
		pdebug.Printf(`CACHE DEL %s`, key)
	}
	c := Cache()
	cerr := c.Delete(key)
	if pdebug.Enabled {
		if cerr != nil {
			pdebug.Printf(`CACHE ERR: %%s`, cerr)
		}
	}
	return nil
}

func (v *VenueSvc) UpdateFromPayload(ctx context.Context, tx *db.Tx, payload *model.UpdateVenueRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.Venue.UpdateFromPayload (%s)", payload.ID).BindError(&err)
		defer g.End()
	}
	var vdb db.Venue
	if err := vdb.LoadByEID(tx, payload.ID); err != nil {
		return errors.Wrap(err, `failed to load from database`)
	}

	if err := v.populateRowForUpdate(&vdb, payload); err != nil {
		return errors.Wrap(err, `failed to populate row data`)
	}

	if err := v.Update(tx, &vdb); err != nil {
		return errors.Wrap(err, `failed to update row in database`)
	}

	ls := LocalizedString()
	if err := ls.UpdateFields(tx, "Venue", vdb.EID, payload.LocalizedFields); err != nil {
		return errors.Wrap(err, `failed to update localized fields`)
	}
	return nil
}

func (v *VenueSvc) ReplaceL10NStrings(tx *db.Tx, m *model.Venue, lang string) error {
	if pdebug.Enabled {
		g := pdebug.Marker("service.Venue.ReplaceL10NStrings lang = %s", lang)
		defer g.End()
	}
	ls := LocalizedString()
	list := make([]db.LocalizedString, 0, 2)
	switch lang {
	case "", "en":
		if len(m.Name) > 0 && len(m.Address) > 0 {
			return nil
		}
		for _, extralang := range []string{`ja`} {
			list = list[:0]
			if err := ls.LookupFields(tx, "Venue", m.ID, extralang, &list); err != nil {
				return errors.Wrap(err, `failed to lookup localized fields`)
			}

			for _, l := range list {
				switch l.Name {
				case "name":
					if len(m.Name) == 0 {
						if pdebug.Enabled {
							pdebug.Printf("Replacing for key 'name' (fallback en -> %s", l.Language)
						}
						m.Name = l.Localized
					}
				case "address":
					if len(m.Address) == 0 {
						if pdebug.Enabled {
							pdebug.Printf("Replacing for key 'address' (fallback en -> %s", l.Language)
						}
						m.Address = l.Localized
					}
				}
			}
		}
		return nil
	case "all":
		for _, extralang := range []string{`ja`} {
			list = list[:0]
			if err := ls.LookupFields(tx, "Venue", m.ID, extralang, &list); err != nil {
				return errors.Wrap(err, `failed to lookup localized fields`)
			}

			for _, l := range list {
				if pdebug.Enabled {
					pdebug.Printf("Adding key '%s#%s'", l.Name, l.Language)
				}
				m.LocalizedFields.Set(l.Language, l.Name, l.Localized)
			}
		}
	default:
		for _, extralang := range []string{`ja`} {
			list = list[:0]
			if err := ls.LookupFields(tx, "Venue", m.ID, extralang, &list); err != nil {
				return errors.Wrap(err, `failed to lookup localized fields`)
			}

			for _, l := range list {
				switch l.Name {
				case "name":
					if pdebug.Enabled {
						pdebug.Printf("Replacing for key 'name'")
					}
					m.Name = l.Localized
				case "address":
					if pdebug.Enabled {
						pdebug.Printf("Replacing for key 'address'")
					}
					m.Address = l.Localized
				}
			}
		}
	}
	return nil
}

func (v *VenueSvc) Delete(tx *db.Tx, id string) error {
	if pdebug.Enabled {
		g := pdebug.Marker("Venue.Delete (%s)", id)
		defer g.End()
	}

	vdb := db.Venue{EID: id}
	if err := vdb.Delete(tx); err != nil {
		return err
	}
	key := `api.Venue.` + id
	c := Cache()
	c.Delete(key)
	if pdebug.Enabled {
		pdebug.Printf(`CACHE DEL %s`, key)
	}
	if err := db.DeleteLocalizedStringsForParent(tx, id, "Venue"); err != nil {
		return err
	}
	return nil
}
