//go:build example
// +build example

/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2021 MinIO, Inc.
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
	"log"

	minio "github.com/jie123108/minio-go/v7"
	"github.com/jie123108/minio-go/v7/pkg/credentials"
)

func main() {
	// Note: YOUR-ACCESSKEYID, YOUR-SECRETACCESSKEY, my-bucketname, my-objectname and
	// my-testfile are dummy values, please replace them with original values.

	// Requests are always secure (HTTPS) by default. Set secure=false to enable insecure (HTTP) access.
	// This boolean value is the last argument for New().

	// New returns an Amazon S3 compatible client object. API compatibility (v2 or v4) is automatically
	// determined based on the Endpoint value.
	s3Client, err := minio.New("s3.amazonaws.com", &minio.Options{
		Creds:  credentials.NewStaticV4("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
		Secure: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	opts := minio.RestoreRequest{}
	opts.SetType(minio.RestoreSelect)
	opts.SetTier(minio.TierStandard)

	selectParameters := minio.SelectParameters{
		Expression:     "SELECT * FROM object",
		ExpressionType: minio.QueryExpressionTypeSQL,
		InputSerialization: minio.SelectObjectInputSerialization{
			CSV: &minio.CSVInputOptions{
				FileHeaderInfo: minio.CSVFileHeaderInfoUse,
			},
		},
		OutputSerialization: minio.SelectObjectOutputSerialization{
			CSV: &minio.CSVOutputOptions{},
		},
	}

	opts.SetSelectParameters(selectParameters)

	outputLocation := minio.OutputLocation{S3: minio.S3{BucketName: "your-bucket", Prefix: "sql-request-output.csv"}}
	opts.SetOutputLocation(outputLocation)

	err = s3Client.RestoreObject(context.Background(), "your-bucket", "input.csv", "", opts)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Restore SQL request Succeeded.")
}
