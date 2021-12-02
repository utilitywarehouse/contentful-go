package contentful

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// EntriesService service
type EntriesService service

// Entry model
type Entry struct {
	locale string
	Sys    *Sys `json:"sys"`
	Fields map[string]interface{}
}

// GetVersion returns entity version
func (entry *Entry) GetVersion() int {
	version := 1
	if entry.Sys != nil {
		version = entry.Sys.Version
	}

	return version
}

// GetEntryKey returns the entry's keys
func (service *EntriesService) GetEntryKey(ctx context.Context, entry *Entry, key string) (*EntryField, error) {
	ef := EntryField{
		value: entry.Fields[key],
	}

	col, err := service.c.ContentTypes.List(entry.Sys.Space.Sys.ID).Next(ctx)
	if err != nil {
		return nil, err
	}

	for _, ct := range col.ToContentType() {
		if ct.Sys.ID != entry.Sys.ContentType.Sys.ID {
			continue
		}

		for _, field := range ct.Fields {
			if field.ID != key {
				continue
			}

			ef.dataType = field.Type
		}
	}

	return &ef, nil
}

// List returns entries collection
func (service *EntriesService) List(spaceID string) *Collection {
	path := fmt.Sprintf("/spaces/%s/environments/%s/entries", spaceID, service.c.Environment)

	req, err := service.c.newRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return &Collection{}
	}

	col := NewCollection(&CollectionOptions{})
	col.c = service.c
	col.req = req

	return col
}

// Get returns a single entry
func (service *EntriesService) Get(ctx context.Context, spaceID, entryID string) (*Entry, error) {
	path := fmt.Sprintf("/spaces/%s/entries/%s", spaceID, entryID)
	query := url.Values{}
	method := "GET"

	req, err := service.c.newRequest(method, path, query, nil)
	if err != nil {
		return &Entry{}, err
	}

	var entry Entry
	if ok := service.c.do(req.WithContext(ctx), &entry); ok != nil {
		return nil, err
	}

	return &entry, err
}

// Upsert updates or creates a new entry
func (service *EntriesService) Upsert(ctx context.Context, spaceID string, entry *Entry) error {
	fields := map[string]interface{}{
		"fields": entry.Fields,
	}

	bytesArray, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	// Creating/updating an entry requires a content type to be provided
	if entry.Sys.ContentType == nil {
		return fmt.Errorf("creating/updating an entry requires a content type")
	}

	var path string
	var method string

	if entry.Sys != nil && entry.Sys.CreatedAt != "" {
		path = fmt.Sprintf("/spaces/%s/environments/%s/entries/%s", spaceID, service.c.Environment, entry.Sys.ID)
		method = http.MethodPut
	} else {
		path = fmt.Sprintf("/spaces/%s/environments/%s/entries", spaceID, service.c.Environment)
		method = http.MethodPost
	}

	req, err := service.c.newRequest(method, path, nil, bytes.NewReader(bytesArray))
	if err != nil {
		return err
	}

	version := strconv.Itoa(entry.Sys.Version)
	req.Header.Set("X-Contentful-Version", version)
	req.Header.Set("X-Contentful-Content-Type", entry.Sys.ContentType.Sys.ID)

	return service.c.do(req.WithContext(ctx), entry)
}

// Delete the entry
func (service *EntriesService) Delete(ctx context.Context, spaceID string, entryID string) error {
	path := fmt.Sprintf("/spaces/%s/entries/%s", spaceID, entryID)
	method := "DELETE"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return err
	}

	return service.c.do(req.WithContext(ctx), nil)
}

// Publish the entry
func (service *EntriesService) Publish(ctx context.Context, spaceID string, entry *Entry) error {
	path := fmt.Sprintf("/spaces/%s/entries/%s/published", spaceID, entry.Sys.ID)
	method := "PUT"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return err
	}

	version := strconv.Itoa(entry.Sys.Version)
	req.Header.Set("X-Contentful-Version", version)

	return service.c.do(req.WithContext(ctx), nil)
}

// Unpublish the entry
func (service *EntriesService) Unpublish(ctx context.Context, spaceID string, entry *Entry) error {
	path := fmt.Sprintf("/spaces/%s/entries/%s/published", spaceID, entry.Sys.ID)
	method := "DELETE"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return err
	}

	version := strconv.Itoa(entry.Sys.Version)
	req.Header.Set("X-Contentful-Version", version)

	return service.c.do(req.WithContext(ctx), nil)
}
