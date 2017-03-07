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
	"time"

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
const deliveryStreamName string = "beginnerd_firehose_stream"

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
		t, err := tail.TailFile(fileName, tail.Config{Follow: true, ReOpen: true, Poll: true})
		if err != nil {
			fmt.Println("TailFile error:", err)
		}

		printlnChan := make(chan string)
		messages := make(chan string)

		go sendLogs(printlnChan, messages)
		go func() {
			for line := range t.Lines {
				printlnChan <- line.Text
				messages <- line.Text
			}
		}()

		for {
			fmt.Println(<-printlnChan)
		}

	},
}

func sendLogs(printlnChan, messages chan string) {
	svc := firehose.New(session.New())
	records := make([]*firehose.Record, 0, maxBatchSize)
	timeoutFlag := false
	for {
		select {
		case text := <-messages:
			records = append(
				records,
				&firehose.Record{Data: append([]byte(text), '\n')},
			)
		case <-time.After(time.Second * 15):
			timeoutFlag = len(records) > 0
		}

		if len(records) >= maxBatchSize || timeoutFlag {
			_, err := sendToKinesis(svc, records)
			if err != nil {
				fmt.Println("Firehose error:", err)
			} else {
				records = records[:0]
			}
		}
		if timeoutFlag {
			printlnChan <- "15s timeout, remaining data sent"
			timeoutFlag = false
		}
	}
}

func sendToKinesis(svc *firehose.Firehose, records []*firehose.Record) (*firehose.PutRecordBatchOutput, error) {
	return svc.PutRecordBatch(
		&firehose.PutRecordBatchInput{
			DeliveryStreamName: aws.String(deliveryStreamName),
			Records:            records,
		},
	)
}

func init() {
	RootCmd.AddCommand(agglogsCmd)
	agglogsCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "", "path to file")
}
