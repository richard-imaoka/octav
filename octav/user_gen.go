// Automatically generated by genmodel utility. DO NOT EDIT!
package octav

import (
	"encoding/json"
	"time"

	"github.com/builderscon/octav/octav/db"
	"github.com/builderscon/octav/octav/tools"
	"github.com/lestrrat/go-pdebug"
)

var _ = time.Time{}

func (v User) GetPropNames() ([]string, error) {
	l, _ := v.L10N.GetPropNames()
	return append(l, "id", "first_name", "last_name", "nickname", "email", "tshirt_size"), nil
}

func (v User) GetPropValue(s string) (interface{}, error) {
	switch s {
	case "id":
		return v.ID, nil
	case "first_name":
		return v.FirstName, nil
	case "last_name":
		return v.LastName, nil
	case "nickname":
		return v.Nickname, nil
	case "email":
		return v.Email, nil
	case "tshirt_size":
		return v.TshirtSize, nil
	default:
		return v.L10N.GetPropValue(s)
	}
}

func (v User) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	m["id"] = v.ID
	m["first_name"] = v.FirstName
	m["last_name"] = v.LastName
	m["nickname"] = v.Nickname
	m["email"] = v.Email
	m["tshirt_size"] = v.TshirtSize
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return tools.MarshalJSONWithL10N(buf, v.L10N)
}

func (v *User) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if jv, ok := m["id"]; ok {
		switch jv.(type) {
		case string:
			v.ID = jv.(string)
			delete(m, "id")
		default:
			return ErrInvalidFieldType{Field: "id"}
		}
	}

	if jv, ok := m["first_name"]; ok {
		switch jv.(type) {
		case string:
			v.FirstName = jv.(string)
			delete(m, "first_name")
		default:
			return ErrInvalidFieldType{Field: "first_name"}
		}
	}

	if jv, ok := m["last_name"]; ok {
		switch jv.(type) {
		case string:
			v.LastName = jv.(string)
			delete(m, "last_name")
		default:
			return ErrInvalidFieldType{Field: "last_name"}
		}
	}

	if jv, ok := m["nickname"]; ok {
		switch jv.(type) {
		case string:
			v.Nickname = jv.(string)
			delete(m, "nickname")
		default:
			return ErrInvalidFieldType{Field: "nickname"}
		}
	}

	if jv, ok := m["email"]; ok {
		switch jv.(type) {
		case string:
			v.Email = jv.(string)
			delete(m, "email")
		default:
			return ErrInvalidFieldType{Field: "email"}
		}
	}

	if jv, ok := m["tshirt_size"]; ok {
		switch jv.(type) {
		case string:
			v.TshirtSize = jv.(string)
			delete(m, "tshirt_size")
		default:
			return ErrInvalidFieldType{Field: "tshirt_size"}
		}
	}
	return nil
}

func (v *User) Load(tx *db.Tx, id string) error {
	vdb := db.User{}
	if err := vdb.LoadByEID(tx, id); err != nil {
		return err
	}

	if err := v.FromRow(vdb); err != nil {
		return err
	}
	if err := v.LoadLocalizedFields(tx); err != nil {
		return err
	}
	return nil
}

func (v *User) LoadLocalizedFields(tx *db.Tx) error {
	ls, err := db.LoadLocalizedStringsForParent(tx, v.ID, "User")
	if err != nil {
		return err
	}

	if len(ls) > 0 {
		v.L10N = tools.LocalizedFields{}
		for _, l := range ls {
			v.L10N.Set(l.Language, l.Name, l.Localized)
		}
	}
	return nil
}

func (v *User) FromRow(vdb db.User) error {
	v.ID = vdb.EID
	v.FirstName = vdb.FirstName
	v.LastName = vdb.LastName
	v.Nickname = vdb.Nickname
	v.Email = vdb.Email
	v.TshirtSize = vdb.TshirtSize
	return nil
}

func (v *User) ToRow(vdb *db.User) error {
	vdb.EID = v.ID
	vdb.FirstName = v.FirstName
	vdb.LastName = v.LastName
	vdb.Nickname = v.Nickname
	vdb.Email = v.Email
	vdb.TshirtSize = v.TshirtSize
	return nil
}

func (v *User) Create(tx *db.Tx) error {
	if v.ID == "" {
		v.ID = tools.UUID()
	}

	vdb := db.User{}
	v.ToRow(&vdb)
	if err := vdb.Create(tx); err != nil {
		return err
	}

	if err := v.L10N.CreateLocalizedStrings(tx, "User", v.ID); err != nil {
		return err
	}
	return nil
}

func (v *User) Update(tx *db.Tx) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("User.Update (%s)", v.ID).BindError(&err)
		defer g.End()
	}

	vdb := db.User{}
	v.ToRow(&vdb)
	if err := vdb.Update(tx); err != nil {
		return err
	}

	return v.L10N.Foreach(func(l, k, x string) error {
		ls := db.LocalizedString{
			ParentType: "User",
			ParentID:   v.ID,
			Language:   l,
			Name:       k,
			Localized:  x,
		}
		return ls.Upsert(tx)
	})
}

func (v *User) Delete(tx *db.Tx) error {
	if pdebug.Enabled {
		g := pdebug.Marker("User.Delete (%s)", v.ID)
		defer g.End()
	}

	vdb := db.User{EID: v.ID}
	if err := vdb.Delete(tx); err != nil {
		return err
	}
	if err := db.DeleteLocalizedStringsForParent(tx, v.ID, "User"); err != nil {
		return err
	}
	return nil
}

func (v *UserList) Load(tx *db.Tx, since string, limit int) error {
	vdbl := db.UserList{}
	if err := vdbl.LoadSinceEID(tx, since, limit); err != nil {
		return err
	}
	res := make([]User, len(vdbl))
	for i, vdb := range vdbl {
		v := User{}
		if err := v.FromRow(vdb); err != nil {
			return err
		}
		if err := v.LoadLocalizedFields(tx); err != nil {
			return err
		}
		res[i] = v
	}
	*v = res
	return nil
}
