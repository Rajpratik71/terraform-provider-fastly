package fastly

import (
	"fmt"
	"net/url"
	"sort"
	"time"
)

// Gzip represents an Gzip logging response from the Fastly API.
type Gzip struct {
	ServiceID      string `mapstructure:"service_id"`
	ServiceVersion int    `mapstructure:"version"`

	Name           string     `mapstructure:"name"`
	ContentTypes   string     `mapstructure:"content_types"`
	Extensions     string     `mapstructure:"extensions"`
	CacheCondition string     `mapstructure:"cache_condition"`
	CreatedAt      *time.Time `mapstructure:"created_at"`
	UpdatedAt      *time.Time `mapstructure:"updated_at"`
	DeletedAt      *time.Time `mapstructure:"deleted_at"`
}

// gzipsByName is a sortable list of gzips.
type gzipsByName []*Gzip

// Len, Swap, and Less implement the sortable interface.
func (s gzipsByName) Len() int      { return len(s) }
func (s gzipsByName) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s gzipsByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// ListGzipsInput is used as input to the ListGzips function.
type ListGzipsInput struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int
}

// ListGzips returns the list of gzips for the configuration version.
func (c *Client) ListGzips(i *ListGzipsInput) ([]*Gzip, error) {
	if i.ServiceID == "" {
		return nil, ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return nil, ErrMissingServiceVersion
	}

	path := fmt.Sprintf("/service/%s/version/%d/gzip", i.ServiceID, i.ServiceVersion)
	resp, err := c.Get(path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gzips []*Gzip
	if err := decodeBodyMap(resp.Body, &gzips); err != nil {
		return nil, err
	}
	sort.Stable(gzipsByName(gzips))
	return gzips, nil
}

// CreateGzipInput is used as input to the CreateGzip function.
type CreateGzipInput struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int

	Name           string `url:"name,omitempty"`
	ContentTypes   string `url:"content_types,omitempty"`
	Extensions     string `url:"extensions,omitempty"`
	CacheCondition string `url:"cache_condition,omitempty"`
}

// CreateGzip creates a new Fastly Gzip.
func (c *Client) CreateGzip(i *CreateGzipInput) (*Gzip, error) {
	if i.ServiceID == "" {
		return nil, ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return nil, ErrMissingServiceVersion
	}

	path := fmt.Sprintf("/service/%s/version/%d/gzip", i.ServiceID, i.ServiceVersion)
	resp, err := c.PostForm(path, i, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gzip *Gzip
	if err := decodeBodyMap(resp.Body, &gzip); err != nil {
		return nil, err
	}
	return gzip, nil
}

// GetGzipInput is used as input to the GetGzip function.
type GetGzipInput struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int

	// Name is the name of the Gzip to fetch.
	Name string
}

// GetGzip gets the Gzip configuration with the given parameters.
func (c *Client) GetGzip(i *GetGzipInput) (*Gzip, error) {
	if i.ServiceID == "" {
		return nil, ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return nil, ErrMissingServiceVersion
	}

	if i.Name == "" {
		return nil, ErrMissingName
	}

	path := fmt.Sprintf("/service/%s/version/%d/gzip/%s", i.ServiceID, i.ServiceVersion, url.PathEscape(i.Name))
	resp, err := c.Get(path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var b *Gzip
	if err := decodeBodyMap(resp.Body, &b); err != nil {
		return nil, err
	}
	return b, nil
}

// UpdateGzipInput is used as input to the UpdateGzip function.
type UpdateGzipInput struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int

	// Name is the name of the Gzip to update.
	Name string

	NewName        *string `url:"name,omitempty"`
	ContentTypes   *string `url:"content_types,omitempty"`
	Extensions     *string `url:"extensions,omitempty"`
	CacheCondition *string `url:"cache_condition,omitempty"`
}

// UpdateGzip updates a specific Gzip.
func (c *Client) UpdateGzip(i *UpdateGzipInput) (*Gzip, error) {
	if i.ServiceID == "" {
		return nil, ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return nil, ErrMissingServiceVersion
	}

	if i.Name == "" {
		return nil, ErrMissingName
	}

	path := fmt.Sprintf("/service/%s/version/%d/gzip/%s", i.ServiceID, i.ServiceVersion, url.PathEscape(i.Name))
	resp, err := c.PutForm(path, i, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var b *Gzip
	if err := decodeBodyMap(resp.Body, &b); err != nil {
		return nil, err
	}
	return b, nil
}

// DeleteGzipInput is the input parameter to DeleteGzip.
type DeleteGzipInput struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int

	// Name is the name of the Gzip to delete (required).
	Name string
}

// DeleteGzip deletes the given Gzip version.
func (c *Client) DeleteGzip(i *DeleteGzipInput) error {
	if i.ServiceID == "" {
		return ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return ErrMissingServiceVersion
	}

	if i.Name == "" {
		return ErrMissingName
	}

	path := fmt.Sprintf("/service/%s/version/%d/gzip/%s", i.ServiceID, i.ServiceVersion, url.PathEscape(i.Name))
	resp, err := c.Delete(path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r *statusResp
	if err := decodeBodyMap(resp.Body, &r); err != nil {
		return err
	}
	if !r.Ok() {
		return ErrNotOK
	}
	return nil
}
