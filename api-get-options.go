/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2020 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package minio

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jie123108/minio-go/v7/pkg/encrypt"
)

// AdvancedGetOptions for internal use by MinIO server - not intended for client use.
type AdvancedGetOptions struct {
	ReplicationDeleteMarker           bool
	IsReplicationReadyForDeleteMarker bool
	ReplicationProxyRequest           string
}

// GetObjectOptions are used to specify additional headers or options
// during GET requests.
type GetObjectOptions struct {
	headers              map[string]string
	reqParams            url.Values
	ServerSideEncryption encrypt.ServerSide
	VersionID            string
	PartNumber           int

	// Include any checksums, if object was uploaded with checksum.
	// For multipart objects this is a checksum of part checksums.
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html
	Checksum bool

	// To be not used by external applications
	Internal AdvancedGetOptions
}

// StatObjectOptions are used to specify additional headers or options
// during GET info/stat requests.
type StatObjectOptions = GetObjectOptions

// Header returns the http.Header representation of the GET options.
func (o GetObjectOptions) Header() http.Header {
	headers := make(http.Header, len(o.headers))
	for k, v := range o.headers {
		headers.Set(k, v)
	}
	if o.ServerSideEncryption != nil && o.ServerSideEncryption.Type() == encrypt.SSEC {
		o.ServerSideEncryption.Marshal(headers)
	}
	// this header is set for active-active replication scenario where GET/HEAD
	// to site A is proxy'd to site B if object/version missing on site A.
	if o.Internal.ReplicationProxyRequest != "" {
		headers.Set(minIOBucketReplicationProxyRequest, o.Internal.ReplicationProxyRequest)
	}
	if o.Checksum {
		headers.Set("x-amz-checksum-mode", "ENABLED")
	}
	return headers
}

// Set adds a key value pair to the options. The
// key-value pair will be part of the HTTP GET request
// headers.
func (o *GetObjectOptions) Set(key, value string) {
	if o.headers == nil {
		o.headers = make(map[string]string)
	}
	o.headers[http.CanonicalHeaderKey(key)] = value
}

// SetReqParam - set request query string parameter
// supported key: see supportedQueryValues and allowedCustomQueryPrefix.
// If an unsupported key is passed in, it will be ignored and nothing will be done.
func (o *GetObjectOptions) SetReqParam(key, value string) {
	if !isCustomQueryValue(key) && !isStandardQueryValue(key) {
		// do nothing
		return
	}
	if o.reqParams == nil {
		o.reqParams = make(url.Values)
	}
	o.reqParams.Set(key, value)
}

// AddReqParam - add request query string parameter
// supported key: see supportedQueryValues and allowedCustomQueryPrefix.
// If an unsupported key is passed in, it will be ignored and nothing will be done.
func (o *GetObjectOptions) AddReqParam(key, value string) {
	if !isCustomQueryValue(key) && !isStandardQueryValue(key) {
		// do nothing
		return
	}
	if o.reqParams == nil {
		o.reqParams = make(url.Values)
	}
	o.reqParams.Add(key, value)
}

// SetMatchETag - set match etag.
func (o *GetObjectOptions) SetMatchETag(etag string) error {
	if etag == "" {
		return errInvalidArgument("ETag cannot be empty.")
	}
	o.Set("If-Match", "\""+etag+"\"")
	return nil
}

// SetMatchETagExcept - set match etag except.
func (o *GetObjectOptions) SetMatchETagExcept(etag string) error {
	if etag == "" {
		return errInvalidArgument("ETag cannot be empty.")
	}
	o.Set("If-None-Match", "\""+etag+"\"")
	return nil
}

// SetUnmodified - set unmodified time since.
func (o *GetObjectOptions) SetUnmodified(modTime time.Time) error {
	if modTime.IsZero() {
		return errInvalidArgument("Modified since cannot be empty.")
	}
	o.Set("If-Unmodified-Since", modTime.Format(http.TimeFormat))
	return nil
}

// SetModified - set modified time since.
func (o *GetObjectOptions) SetModified(modTime time.Time) error {
	if modTime.IsZero() {
		return errInvalidArgument("Modified since cannot be empty.")
	}
	o.Set("If-Modified-Since", modTime.Format(http.TimeFormat))
	return nil
}

// SetRange - set the start and end offset of the object to be read.
// See https://tools.ietf.org/html/rfc7233#section-3.1 for reference.
func (o *GetObjectOptions) SetRange(start, end int64) error {
	switch {
	case start == 0 && end < 0:
		// Read last '-end' bytes. `bytes=-N`.
		o.Set("Range", fmt.Sprintf("bytes=%d", end))
	case 0 < start && end == 0:
		// Read everything starting from offset
		// 'start'. `bytes=N-`.
		o.Set("Range", fmt.Sprintf("bytes=%d-", start))
	case 0 <= start && start <= end:
		// Read everything starting at 'start' till the
		// 'end'. `bytes=N-M`
		o.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	default:
		// All other cases such as
		// bytes=-3-
		// bytes=5-3
		// bytes=-2-4
		// bytes=-3-0
		// bytes=-3--2
		// are invalid.
		return errInvalidArgument(
			fmt.Sprintf(
				"Invalid range specified: start=%d end=%d",
				start, end))
	}
	return nil
}

// toQueryValues - Convert the versionId, partNumber, and reqParams in Options to query string parameters.
func (o *GetObjectOptions) toQueryValues() url.Values {
	urlValues := make(url.Values)
	if o.VersionID != "" {
		urlValues.Set("versionId", o.VersionID)
	}
	if o.PartNumber > 0 {
		urlValues.Set("partNumber", strconv.Itoa(o.PartNumber))
	}

	if o.reqParams != nil {
		for key, values := range o.reqParams {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
	}

	return urlValues
}
