package fastly

import (
	"fmt"
	"net/url"
	"sort"
	"time"
)

type S3Redundancy string

func S3RedundancyPtr(v S3Redundancy) *S3Redundancy { return &v }

type S3ServerSideEncryption string

func S3ServerSideEncryptionPtr(v S3ServerSideEncryption) *S3ServerSideEncryption { return &v }

type S3AccessControlList string

func S3AccessControlListPtr(v S3AccessControlList) *S3AccessControlList { return &v }

const (
	S3RedundancyStandard                 S3Redundancy = "standard"
	S3RedundancyIntelligentTiering       S3Redundancy = "intelligent_tiering"
	S3RedundancyStandardIA               S3Redundancy = "standard_ia"
	S3RedundancyOneZoneIA                S3Redundancy = "onezone_ia"
	S3RedundancyGlacierInstantRetrieval  S3Redundancy = "glacier_ir"
	S3RedundancyGlacierFlexibleRetrieval S3Redundancy = "glacier"
	S3RedundancyGlacierDeepArchive       S3Redundancy = "deep_archive"
	S3RedundancyReduced                  S3Redundancy = "reduced_redundancy"

	S3ServerSideEncryptionAES S3ServerSideEncryption = "AES256"
	S3ServerSideEncryptionKMS S3ServerSideEncryption = "aws:kms"

	S3AccessControlListPrivate                S3AccessControlList = "private"
	S3AccessControlListPublicRead             S3AccessControlList = "public-read"
	S3AccessControlListPublicReadWrite        S3AccessControlList = "public-read-write"
	S3AccessControlListAWSExecRead            S3AccessControlList = "aws-exec-read"
	S3AccessControlListAuthenticatedRead      S3AccessControlList = "authenticated-read"
	S3AccessControlListBucketOwnerRead        S3AccessControlList = "bucket-owner-read"
	S3AccessControlListBucketOwnerFullControl S3AccessControlList = "bucket-owner-full-control"
)

// S3 represents a S3 response from the Fastly API.
type S3 struct {
	ServiceID      string `mapstructure:"service_id"`
	ServiceVersion int    `mapstructure:"version"`

	Name                         string                 `mapstructure:"name"`
	BucketName                   string                 `mapstructure:"bucket_name"`
	Domain                       string                 `mapstructure:"domain"`
	AccessKey                    string                 `mapstructure:"access_key"`
	SecretKey                    string                 `mapstructure:"secret_key"`
	IAMRole                      string                 `mapstructure:"iam_role"`
	Path                         string                 `mapstructure:"path"`
	Period                       uint                   `mapstructure:"period"`
	CompressionCodec             string                 `mapstructure:"compression_codec"`
	GzipLevel                    uint8                  `mapstructure:"gzip_level"`
	Format                       string                 `mapstructure:"format"`
	FormatVersion                uint                   `mapstructure:"format_version"`
	ResponseCondition            string                 `mapstructure:"response_condition"`
	MessageType                  string                 `mapstructure:"message_type"`
	TimestampFormat              string                 `mapstructure:"timestamp_format"`
	Placement                    string                 `mapstructure:"placement"`
	PublicKey                    string                 `mapstructure:"public_key"`
	Redundancy                   S3Redundancy           `mapstructure:"redundancy"`
	ServerSideEncryptionKMSKeyID string                 `mapstructure:"server_side_encryption_kms_key_id"`
	ServerSideEncryption         S3ServerSideEncryption `mapstructure:"server_side_encryption"`
	CreatedAt                    *time.Time             `mapstructure:"created_at"`
	UpdatedAt                    *time.Time             `mapstructure:"updated_at"`
	DeletedAt                    *time.Time             `mapstructure:"deleted_at"`
	ACL                          S3AccessControlList    `mapstructure:"acl"`
}

// s3sByName is a sortable list of S3s.
type s3sByName []*S3

// Len, Swap, and Less implement the sortable interface.
func (s s3sByName) Len() int      { return len(s) }
func (s s3sByName) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s s3sByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// ListS3sInput is used as input to the ListS3s function.
type ListS3sInput struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int
}

// ListS3s returns the list of S3s for the configuration version.
func (c *Client) ListS3s(i *ListS3sInput) ([]*S3, error) {
	if i.ServiceID == "" {
		return nil, ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return nil, ErrMissingServiceVersion
	}

	path := fmt.Sprintf("/service/%s/version/%d/logging/s3", i.ServiceID, i.ServiceVersion)
	resp, err := c.Get(path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var s3s []*S3
	if err := decodeBodyMap(resp.Body, &s3s); err != nil {
		return nil, err
	}
	sort.Stable(s3sByName(s3s))
	return s3s, nil
}

// CreateS3Input is used as input to the CreateS3 function.
type CreateS3Input struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int

	Name                         string                 `url:"name,omitempty"`
	BucketName                   string                 `url:"bucket_name,omitempty"`
	Domain                       string                 `url:"domain,omitempty"`
	AccessKey                    string                 `url:"access_key,omitempty"`
	SecretKey                    string                 `url:"secret_key,omitempty"`
	IAMRole                      string                 `url:"iam_role,omitempty"`
	Path                         string                 `url:"path,omitempty"`
	Period                       uint                   `url:"period,omitempty"`
	CompressionCodec             string                 `url:"compression_codec,omitempty"`
	GzipLevel                    uint8                  `url:"gzip_level,omitempty"`
	Format                       string                 `url:"format,omitempty"`
	MessageType                  string                 `url:"message_type,omitempty"`
	FormatVersion                uint                   `url:"format_version,omitempty"`
	ResponseCondition            string                 `url:"response_condition,omitempty"`
	TimestampFormat              string                 `url:"timestamp_format,omitempty"`
	Redundancy                   S3Redundancy           `url:"redundancy,omitempty"`
	Placement                    string                 `url:"placement,omitempty"`
	PublicKey                    string                 `url:"public_key,omitempty"`
	ServerSideEncryptionKMSKeyID string                 `url:"server_side_encryption_kms_key_id,omitempty"`
	ServerSideEncryption         S3ServerSideEncryption `url:"server_side_encryption,omitempty"`
	ACL                          S3AccessControlList    `url:"acl,omitempty"`
}

// CreateS3 creates a new Fastly S3.
func (c *Client) CreateS3(i *CreateS3Input) (*S3, error) {
	if i.ServiceID == "" {
		return nil, ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return nil, ErrMissingServiceVersion
	}

	if i.ServerSideEncryption == S3ServerSideEncryptionKMS && i.ServerSideEncryptionKMSKeyID == "" {
		return nil, ErrMissingServerSideEncryptionKMSKeyID
	}

	path := fmt.Sprintf("/service/%s/version/%d/logging/s3", i.ServiceID, i.ServiceVersion)
	resp, err := c.PostForm(path, i, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var s3 *S3
	if err := decodeBodyMap(resp.Body, &s3); err != nil {
		return nil, err
	}
	return s3, nil
}

// GetS3Input is used as input to the GetS3 function.
type GetS3Input struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int

	// Name is the name of the S3 to fetch.
	Name string
}

// GetS3 gets the S3 configuration with the given parameters.
func (c *Client) GetS3(i *GetS3Input) (*S3, error) {
	if i.ServiceID == "" {
		return nil, ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return nil, ErrMissingServiceVersion
	}

	if i.Name == "" {
		return nil, ErrMissingName
	}

	path := fmt.Sprintf("/service/%s/version/%d/logging/s3/%s", i.ServiceID, i.ServiceVersion, url.PathEscape(i.Name))
	resp, err := c.Get(path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var s3 *S3
	if err := decodeBodyMap(resp.Body, &s3); err != nil {
		return nil, err
	}
	return s3, nil
}

// UpdateS3Input is used as input to the UpdateS3 function.
type UpdateS3Input struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int

	// Name is the name of the S3 to update.
	Name string

	NewName                      *string                 `url:"name,omitempty"`
	BucketName                   *string                 `url:"bucket_name,omitempty"`
	Domain                       *string                 `url:"domain,omitempty"`
	AccessKey                    *string                 `url:"access_key,omitempty"`
	SecretKey                    *string                 `url:"secret_key,omitempty"`
	IAMRole                      *string                 `url:"iam_role,omitempty"`
	Path                         *string                 `url:"path,omitempty"`
	Period                       *uint                   `url:"period,omitempty"`
	CompressionCodec             *string                 `url:"compression_codec,omitempty"`
	GzipLevel                    *uint8                  `url:"gzip_level,omitempty"`
	Format                       *string                 `url:"format,omitempty"`
	FormatVersion                *uint                   `url:"format_version,omitempty"`
	ResponseCondition            *string                 `url:"response_condition,omitempty"`
	MessageType                  *string                 `url:"message_type,omitempty"`
	TimestampFormat              *string                 `url:"timestamp_format,omitempty"`
	Redundancy                   *S3Redundancy           `url:"redundancy,omitempty"`
	Placement                    *string                 `url:"placement,omitempty"`
	PublicKey                    *string                 `url:"public_key,omitempty"`
	ServerSideEncryptionKMSKeyID *string                 `url:"server_side_encryption_kms_key_id,omitempty"`
	ServerSideEncryption         *S3ServerSideEncryption `url:"server_side_encryption,omitempty"`
	ACL                          *S3AccessControlList    `url:"acl,omitempty"`
}

// UpdateS3 updates a specific S3.
func (c *Client) UpdateS3(i *UpdateS3Input) (*S3, error) {
	if i.ServiceID == "" {
		return nil, ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return nil, ErrMissingServiceVersion
	}

	if i.Name == "" {
		return nil, ErrMissingName
	}

	if i.ServerSideEncryption != nil && *i.ServerSideEncryption == S3ServerSideEncryptionKMS && *i.ServerSideEncryptionKMSKeyID == "" {
		return nil, ErrMissingServerSideEncryptionKMSKeyID
	}

	path := fmt.Sprintf("/service/%s/version/%d/logging/s3/%s", i.ServiceID, i.ServiceVersion, url.PathEscape(i.Name))
	resp, err := c.PutForm(path, i, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var s3 *S3
	if err := decodeBodyMap(resp.Body, &s3); err != nil {
		return nil, err
	}
	return s3, nil
}

// DeleteS3Input is the input parameter to DeleteS3.
type DeleteS3Input struct {
	// ServiceID is the ID of the service (required).
	ServiceID string

	// ServiceVersion is the specific configuration version (required).
	ServiceVersion int

	// Name is the name of the S3 to delete (required).
	Name string
}

// DeleteS3 deletes the given S3 version.
func (c *Client) DeleteS3(i *DeleteS3Input) error {
	if i.ServiceID == "" {
		return ErrMissingServiceID
	}

	if i.ServiceVersion == 0 {
		return ErrMissingServiceVersion
	}

	if i.Name == "" {
		return ErrMissingName
	}

	path := fmt.Sprintf("/service/%s/version/%d/logging/s3/%s", i.ServiceID, i.ServiceVersion, url.PathEscape(i.Name))
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
