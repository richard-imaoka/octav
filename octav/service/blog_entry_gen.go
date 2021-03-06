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
var _ = model.BlogEntry{}
var _ = db.BlogEntry{}
var _ = pdebug.Enabled

var blogEntrySvc BlogEntrySvc
var blogEntryOnce sync.Once

func BlogEntry() *BlogEntrySvc {
	blogEntryOnce.Do(blogEntrySvc.Init)
	return &blogEntrySvc
}

func (v *BlogEntrySvc) LookupFromPayload(tx *db.Tx, m *model.BlogEntry, payload *model.LookupBlogEntryRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.BlogEntry.LookupFromPayload").BindError(&err)
		defer g.End()
	}
	if err = v.Lookup(tx, m, payload.ID); err != nil {
		return errors.Wrap(err, "failed to load model.BlogEntry from database")
	}
	return nil
}

func (v *BlogEntrySvc) Lookup(tx *db.Tx, m *model.BlogEntry, id string) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.BlogEntry.Lookup").BindError(&err)
		defer g.End()
	}

	var r model.BlogEntry
	c := Cache()
	key := c.Key("BlogEntry", id)
	var cacheMiss bool
	_, err = c.GetOrSet(key, &r, func() (interface{}, error) {
		if pdebug.Enabled {
			cacheMiss = true
		}
		if err := r.Load(tx, id); err != nil {
			return nil, errors.Wrap(err, "failed to load model.BlogEntry from database")
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
func (v *BlogEntrySvc) Create(tx *db.Tx, vdb *db.BlogEntry, payload *model.CreateBlogEntryRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.BlogEntry.Create").BindError(&err)
		defer g.End()
	}

	if err := v.populateRowForCreate(vdb, payload); err != nil {
		return errors.Wrap(err, `failed to populate row`)
	}

	if err := vdb.Create(tx, payload.DatabaseOptions...); err != nil {
		return errors.Wrap(err, `failed to insert into database`)
	}

	if err := v.PostCreateHook(tx, vdb); err != nil {
		return errors.Wrap(err, `post create hook failed`)
	}
	return nil
}

func (v *BlogEntrySvc) Update(tx *db.Tx, vdb *db.BlogEntry) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.BlogEntry.Update (%s)", vdb.EID).BindError(&err)
		defer g.End()
	}

	if vdb.EID == `` {
		return errors.New("vdb.EID is required (did you forget to call vdb.Load(tx) before hand?)")
	}

	if err := vdb.Update(tx); err != nil {
		return errors.Wrap(err, `failed to update database`)
	}
	c := Cache()
	key := c.Key("BlogEntry", vdb.EID)
	if pdebug.Enabled {
		pdebug.Printf(`CACHE DEL %s`, key)
	}
	cerr := c.Delete(key)
	if pdebug.Enabled {
		if cerr != nil {
			pdebug.Printf(`CACHE ERR: %s`, cerr)
		}
	}
	if err := v.PostUpdateHook(tx, vdb); err != nil {
		return errors.Wrap(err, `post update hook failed`)
	}
	return nil
}

func (v *BlogEntrySvc) UpdateFromPayload(ctx context.Context, tx *db.Tx, payload *model.UpdateBlogEntryRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("service.BlogEntry.UpdateFromPayload (%s)", payload.ID).BindError(&err)
		defer g.End()
	}
	var vdb db.BlogEntry
	if err := vdb.LoadByEID(tx, payload.ID); err != nil {
		return errors.Wrap(err, `failed to load from database`)
	}

	if err := v.populateRowForUpdate(&vdb, payload); err != nil {
		return errors.Wrap(err, `failed to populate row data`)
	}

	if err := v.Update(tx, &vdb); err != nil {
		return errors.Wrap(err, `failed to update row in database`)
	}
	return nil
}

func (v *BlogEntrySvc) Delete(tx *db.Tx, id string) error {
	if pdebug.Enabled {
		g := pdebug.Marker("BlogEntry.Delete (%s)", id)
		defer g.End()
	}
	original := db.BlogEntry{EID: id}
	if err := original.LoadByEID(tx, id); err != nil {
		return errors.Wrap(err, `failed load before delete`)
	}

	vdb := db.BlogEntry{EID: id}
	if err := vdb.Delete(tx); err != nil {
		return errors.Wrap(err, `failed to delete from database`)
	}
	c := Cache()
	key := c.Key("BlogEntry", id)
	c.Delete(key)
	if pdebug.Enabled {
		pdebug.Printf(`CACHE DEL %s`, key)
	}
	if err := v.PostDeleteHook(tx, &original); err != nil {
		return errors.Wrap(err, `post delete hook failed`)
	}
	return nil
}
