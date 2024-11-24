/*
Copyright © 2024 Adam Kelly <arkelly111@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// openCmd rRepresents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("open called")
	},
}

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Open a note",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cobra.OnInitialize(initConfig)
		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Errorf("There was an error connecting to the database: %v", err)
		}
		defer conn.Close()

		queries := sqlc.New(conn)
		noteID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Errorf("There was an error converting the noteID to an integer: %v", err)
		}

		editor := getEditorConfig()
		if editor == "" {
			log.Warnf("No editor set in config file, falling back to $EDITOR.")
			editor = os.Getenv("EDITOR")
		}

		note, err := queries.ReadNoteByID(ctx, int64(noteID))
		notePath, err := expandPath(note.Path)
		if err != nil {
			log.Errorf("There was an error expanding the path: %v", err)
		} else {
			cmdr := exec.Command(editor, notePath)
			cmdr.Stdin = os.Stdin
			cmdr.Stdout = os.Stdout
			cmdr.Stderr = os.Stderr
			err = cmdr.Run()
			if err != nil {
				log.Errorf("There was an error running the command: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.AddCommand(noteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// openCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// openCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getEditorConfig() string {
	editor := viper.GetString("selected.editor")
	return editor
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = filepath.Join(usr.HomeDir, path[1:])
	}
	return path, nil
}
