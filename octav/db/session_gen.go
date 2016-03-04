// Automatically generated by gendb utility. DO NOT EDIT!
package db

import (
	"errors"
)

var SessionTable = "sessions"

func (s *Session) Scan(scanner interface {
	Scan(...interface{}) error
}) error {
	return scanner.Scan(&s.OID, &s.EID, &s.ConferenceID, &s.RoomID, &s.SpeakerID, &s.Title, &s.Abstract, &s.Memo, &s.StartsOn, &s.Duration, &s.MaterialLevel, &s.Tags, &s.Category, &s.SpokenLanguage, &s.SlideLanguage, &s.SlideSubtitles, &s.SlideURL, &s.VideoURL, &s.PhotoPermission, &s.VideoPermission, &s.HasInterpretation, &s.Status, &s.SortOrder, &s.Confirmed, &s.CreatedOn, &s.ModifiedOn)
}

func (s *Session) LoadByEID(tx *Tx, eid string) error {
	row := tx.QueryRow(`SELECT oid, eid, conference_id, room_id, speaker_id, title, abstract, memo, starts_on, duration, material_level, tags, category, spoken_language, slide_language, slide_subtitles, slide_url, video_url, photo_permission, video_permission, has_interpretation, status, sort_order, confirmed, created_on, modified_on FROM `+SessionTable+` WHERE eid = ?`, eid)
	if err := s.Scan(row); err != nil {
		return err
	}
	return nil
}

func (s *Session) Create(tx *Tx) error {
	if s.EID == "" {
		return errors.New("create: non-empty EID required")
	}

	result, err := tx.Exec(`INSERT INTO eid, conference_id, room_id, speaker_id, title, abstract, memo, starts_on, duration, material_level, tags, category, spoken_language, slide_language, slide_subtitles, slide_url, video_url, photo_permission, video_permission, has_interpretation, status, sort_order, confirmed, created_on, modified_on VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, s.EID, s.ConferenceID, s.RoomID, s.SpeakerID, s.Title, s.Abstract, s.Memo, s.StartsOn, s.Duration, s.MaterialLevel, s.Tags, s.Category, s.SpokenLanguage, s.SlideLanguage, s.SlideSubtitles, s.SlideURL, s.VideoURL, s.PhotoPermission, s.VideoPermission, s.HasInterpretation, s.Status, s.SortOrder, s.Confirmed, s.CreatedOn, s.ModifiedOn)
	if err != nil {
		return err
	}

	lii, err := result.LastInsertId()
	if err != nil {
		return err
	}

	s.OID = lii
	return nil
}