// Copyright © 2017 Rafael Feodrippe <pfeodrippe@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"

	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

var fileName string

const maxBatchSize int = 400

func readDir(dirname string) []os.FileInfo {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		panic(err)
	}
	return files
}

var agglogsCmd = &cobra.Command{
	Use:   "agglogs",
	Short: "Agg logs from a specified file to a s3 bucket",
	Long: `Aggregate logs from a specified file to a s3 bucket
specified by the user.`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if fileName == "" {
			return errors.New("You must give a --file or -f")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		records := make([]*firehose.Record, 0, maxBatchSize)
		t, err := tail.TailFile(fileName, tail.Config{Follow: true, ReOpen: true, Poll: true})
		if err != nil {
			fmt.Println(err)
		}

		for line := range t.Lines {
			fmt.Println(line.Text)
			records = append(
				records,
				&firehose.Record{Data: append([]byte(line.Text), '\n')},
			)
			if len(records) >= maxBatchSize {
				_, err := sendToKinesis(records)
				if err != nil {
					fmt.Println("ERRO:", err)
				} else {
					fmt.Println("!!! sendToKinesis")
					records = records[:0]
				}
			}
		}

	},
}

func sendToKinesis(records []*firehose.Record) (*firehose.PutRecordBatchOutput, error) {
	svc := firehose.New(session.New())
	return svc.PutRecordBatch(
		&firehose.PutRecordBatchInput{
			DeliveryStreamName: aws.String("terraform-kinesis-firehose-test-stream"),
			Records:            records,
		},
	)
}

func init() {
	RootCmd.AddCommand(agglogsCmd)

	agglogsCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "", "path to file")
}
