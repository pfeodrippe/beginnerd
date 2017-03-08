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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

var fileName string
var deleteDBFile bool
var willPrintBufferSize bool

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
		if deleteDBFile {
			fmt.Println("Removing beginnerd.db file...")
			os.Remove("./beginnerd.db")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		t, err := tail.TailFile(fileName, tail.Config{Follow: true, ReOpen: true, Poll: true})
		if err != nil {
			fmt.Println("TailFile error:", err)
		}

		// print to stdout from the goroutine
		printlnChan := make(chan string, 10)
		// lines from tailed file
		messages := make(chan string, 10000)

		go processLogs(printlnChan, messages)

		// tail file
		go func() {
			for line := range t.Lines {
				printlnChan <- line.Text
				messages <- line.Text
			}
		}()

		// infinite loop to show
		for {
			fmt.Println(<-printlnChan)
			if willPrintBufferSize {
				fmt.Println("Buffer size:", len(messages))
			}
		}

	},
}

// receive the lines from the file tailed in a channel
// and send them to the firehose stream.
// save to the local db to keep track of the last sent line.
func processLogs(printlnChan, messages chan string) {
	db, err := bolt.Open("beginnerd.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createBucket(db, "beginnerdBucket")
	if err != nil {
		log.Fatal(err)
	}

	bucketCh := make(chan []byte)
	go readFromBucket(db, "beginnerdBucket", "lastData", bucketCh)
	lastIdAtBucket := <-bucketCh // wait until it reads from bucket
	var lastRecordId float64
	if lastIdAtBucket != nil {
		value, _ := strconv.ParseFloat(string(lastIdAtBucket), 64)
		lastRecordId = value
	} else {
		lastRecordId = -1
	}
	printlnChan <- fmt.Sprintln("Last id was", lastRecordId)

	// initiate a session with aws
	svc := firehose.New(session.New())
	// buffer the records from the channel
	records := make([]*firehose.Record, 0, maxBatchSize)
	// notify about timeout
	timeoutFlag := false
	biggestId := lastRecordId

	for {
		select {
		case text := <-messages:
			t, _ := decodeJson([]byte(text))
			actId := t["id"].(float64)
			if actId > biggestId { // if actual id is bigger than the last record stored at local db
				records = append(
					records,
					&firehose.Record{Data: append([]byte(text), '\n')},
				)
				biggestId = actId
			} else {
				printlnChan <- fmt.Sprintln("Skipping", text)
			}
		case <-time.After(time.Second * 3):
			timeoutFlag = len(records) > 0
		}

		if len(records) >= maxBatchSize || timeoutFlag {
			_, err := sendToKinesis(svc, records)
			if err != nil {
				fmt.Println("Firehose error:", err)
			} else {
				biggestIdStr := strconv.FormatFloat(biggestId, 'f', 0, 64)
				saveToBucket(db, "beginnerdBucket", "lastData", biggestIdStr)
				printlnChan <- "Biggest Id (last): " + biggestIdStr
				records = records[:0]
			}
		}
		if timeoutFlag {
			printlnChan <- "3s timeout, remaining data sent"
			timeoutFlag = false
		}
	}

}

// Helper functions
func decodeJson(data []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	return m, err
}

// nothing special here
func saveToBucket(db *bolt.DB, bucket, key, value string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), []byte(value))
		return err
	})
}

// read through a channel
func readFromBucket(db *bolt.DB, bucket, key string, ch chan []byte) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v := b.Get([]byte(key))
		ch <- v
		return nil
	})
}

// create bucket if it doesn't exist
func createBucket(db *bolt.DB, bucket string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

// use kinesis firehose to send the records
func sendToKinesis(svc *firehose.Firehose, records []*firehose.Record) (*firehose.PutRecordBatchOutput, error) {
	return svc.PutRecordBatch(
		&firehose.PutRecordBatchInput{
			DeliveryStreamName: aws.String(deliveryStreamName),
			Records:            records,
		},
	)
}

// config the cmd
func init() {
	RootCmd.AddCommand(agglogsCmd)

	agglogsCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "", "path to file")
	agglogsCmd.PersistentFlags().BoolVarP(&deleteDBFile, "delete", "d", false, "if set, delete the DB file")
	agglogsCmd.PersistentFlags().BoolVarP(&willPrintBufferSize, "showbuffer", "b", false, "if set, track buffer size")
}
