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
)

var dir string

func readDir(dirname string) []os.FileInfo {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		panic(err)
	}
	return files
}

// agglogsCmd represents the agglogs command
var agglogsCmd = &cobra.Command{
	Use:   "agglogs",
	Short: "Agg logs from a specified file to a s3 bucket",
	Long: `Aggregate logs from a specified file to a s3 bucket
specified by the user.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if dir == "" {
			return errors.New("You must give a --path or -p indicating where is located the directory you want to watch")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		if dir[len(dir)-1:] != "/" {
			dir = dir + "/"
		}
		fmt.Printf("Waiting for files in %s...", dir)
		for len(readDir(dir)) == 0 {
		}
		fmt.Printf("Logging!")
		files, _ := ioutil.ReadDir(dir)
		fmt.Println(files)
		for _, f := range files {
			t, err := tail.TailFile(dir+f.Name(), tail.Config{Follow: true, ReOpen: true, Poll: true})
			if err != nil {
				fmt.Println(err)
			}
			for line := range t.Lines {
				fmt.Println(line.Text)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(agglogsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.
	// agglogsCmd.PersistentFlags().String("foo", "", "A help for foo")
	agglogsCmd.PersistentFlags().StringVarP(&dir, "path", "p", "", "path (directory) you want to watch")
}
