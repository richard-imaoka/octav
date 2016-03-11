// DO NOT EDIT. Automatically generated by hsup
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/builderscon/octav/octav"
	"github.com/lestrrat/go-pdebug"
	"github.com/lestrrat/go-urlenc"
)

type Client struct {
	Client   *http.Client
	Endpoint string
}

func New(s string) *Client {
	return &Client{
		Client:   &http.Client{},
		Endpoint: s,
	}
}

func (c *Client) CreateConference(in *octav.CreateConferenceRequest) (ret *octav.Conference, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.CreateConference").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/conference/create")
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.Conference
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) CreateRoom(in *octav.Room) (ret *octav.Room, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.CreateRoom").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/room/create")
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.Room
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) CreateSession(in *octav.CreateSessionRequest) (ret *octav.Session, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.CreateSession").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/session/create")
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.Session
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) CreateUser(in *octav.CreateUserRequest) (ret *octav.User, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.CreateUser").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/user/create")
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.User
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) CreateVenue(in *octav.Venue) (ret *octav.Venue, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.CreateVenue").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/venue/create")
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.Venue
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) DeleteConference(in *octav.DeleteConferenceRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.DeleteConference").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/conference/delete")
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	return nil
}

func (c *Client) DeleteRoom(in *octav.DeleteRoomRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.DeleteRoom").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/room/delete")
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	return nil
}

func (c *Client) DeleteUser(in *octav.DeleteUserRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.DeleteUser").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/user/delete")
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	return nil
}

func (c *Client) DeleteVenue(in *octav.DeleteVenueRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.DeleteVenue").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/venue/delete")
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	return nil
}

func (c *Client) ListRooms(in *octav.ListRoomRequest) (ret []octav.Room, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.ListRooms").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/room/list")
	if err != nil {
		return nil, err
	}
	buf, err := urlenc.Marshal(in)
	if err != nil {
		return nil, err
	}
	u.RawQuery = string(buf)
	if pdebug.Enabled {
		pdebug.Printf("GET to %s", u.String())
	}
	res, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload []octav.Room
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) ListSessionsByConference(in *octav.ListSessionsByConferenceRequest) (ret interface{}, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.ListSessionsByConference").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/schedule/list")
	if err != nil {
		return nil, err
	}
	buf, err := urlenc.Marshal(in)
	if err != nil {
		return nil, err
	}
	u.RawQuery = string(buf)
	if pdebug.Enabled {
		pdebug.Printf("GET to %s", u.String())
	}
	res, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload interface{}
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) ListVenues(in *octav.ListVenueRequest) (ret []octav.Venue, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.ListVenues").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/venue/list")
	if err != nil {
		return nil, err
	}
	buf, err := urlenc.Marshal(in)
	if err != nil {
		return nil, err
	}
	u.RawQuery = string(buf)
	if pdebug.Enabled {
		pdebug.Printf("GET to %s", u.String())
	}
	res, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload []octav.Venue
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) LookupConference(in *octav.LookupConferenceRequest) (ret *octav.Conference, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.LookupConference").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/conference/lookup")
	if err != nil {
		return nil, err
	}
	buf, err := urlenc.Marshal(in)
	if err != nil {
		return nil, err
	}
	u.RawQuery = string(buf)
	if pdebug.Enabled {
		pdebug.Printf("GET to %s", u.String())
	}
	res, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.Conference
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) LookupRoom(in *octav.LookupRoomRequest) (ret *octav.Room, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.LookupRoom").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/room/lookup")
	if err != nil {
		return nil, err
	}
	buf, err := urlenc.Marshal(in)
	if err != nil {
		return nil, err
	}
	u.RawQuery = string(buf)
	if pdebug.Enabled {
		pdebug.Printf("GET to %s", u.String())
	}
	res, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.Room
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) LookupSession(in *octav.LookupSessionRequest) (ret *octav.Session, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.LookupSession").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/session/lookup")
	if err != nil {
		return nil, err
	}
	buf, err := urlenc.Marshal(in)
	if err != nil {
		return nil, err
	}
	u.RawQuery = string(buf)
	if pdebug.Enabled {
		pdebug.Printf("GET to %s", u.String())
	}
	res, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.Session
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) LookupUser(in *octav.LookupUserRequest) (ret *octav.User, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.LookupUser").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/user/lookup")
	if err != nil {
		return nil, err
	}
	buf, err := urlenc.Marshal(in)
	if err != nil {
		return nil, err
	}
	u.RawQuery = string(buf)
	if pdebug.Enabled {
		pdebug.Printf("GET to %s", u.String())
	}
	res, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.User
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) LookupVenue(in *octav.LookupVenueRequest) (ret *octav.Venue, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.LookupVenue").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/venue/lookup")
	if err != nil {
		return nil, err
	}
	buf, err := urlenc.Marshal(in)
	if err != nil {
		return nil, err
	}
	u.RawQuery = string(buf)
	if pdebug.Enabled {
		pdebug.Printf("GET to %s", u.String())
	}
	res, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	var payload octav.Venue
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Client) UpdateConference(in *octav.UpdateConferenceRequest) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("client.UpdateConference").BindError(&err)
		defer g.End()
	}
	u, err := url.Parse(c.Endpoint + "/v1/conference/update")
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return err
	}
	if pdebug.Enabled {
		pdebug.Printf("POST to %s", u.String())
	}
	res, err := c.Client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(`Invalid response: '%s'`, res.Status)
	}
	return nil
}
