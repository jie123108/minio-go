//go:build example
// +build example

/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2017 MinIO, Inc.
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

package main

import (
	"context"
	"fmt"

	"github.com/jie123108/minio-go/v7"
	"github.com/jie123108/minio-go/v7/pkg/credentials"
)

func main() {
	// Note: YOUR-ACCESSKEYID, YOUR-SECRETACCESSKEY, my-bucketname and my-prefixname
	// are dummy values, please replace them with original values.

	// Requests are always secure (HTTPS) by default. Set secure=false to enable insecure (HTTP) access.
	// This boolean value is the last argument for New().

	// New returns an Amazon S3 compatible client object. API compatibility (v2 or v4) is automatically
	// determined based on the Endpoint value.
	s3Client, err := minio.New("s3.amazonaws.com", &minio.Options{
		Creds:  credentials.NewStaticV4("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// List 'N' number of objects from a bucket-name with a matching prefix.
	listObjectsN := func(bucket, prefix string, recursive bool, N int) (objsInfo []minio.ObjectInfo, err error) {
		ctx, cancel := context.WithCancel(context.Background())
		// Indicate ListObjects go-routine to exit and stop feeding the objectInfo channel.
		defer cancel()
		i := 1
		opts := minio.ListObjectsOptions{
			UseV1:     true,
			Prefix:    prefix,
			Recursive: recursive,
		}
		for object := range s3Client.ListObjects(ctx, bucket, opts) {
			if object.Err != nil {
				return nil, object.Err
			}
			i++
			// Verify if we have printed N objects.
			if i == N {
				break
			}
			objsInfo = append(objsInfo, object)
		}
		return objsInfo, nil
	}

	// List recursively first 100 entries for prefix 'my-prefixname'.
	recursive := true
	objsInfo, err := listObjectsN("my-bucketname", "my-prefixname", recursive, 100)
	if err != nil {
		fmt.Println(err)
	}

	// Print all the entries.
	fmt.Println(objsInfo)
}
