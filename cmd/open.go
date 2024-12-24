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
	"strconv"

	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/utils"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// openCmd Represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "This is the root command for opening notes",
	Long:  `The openCmd is used to open notes but it does not have any functionality and requires a subcommand to be run.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The open command requires a subcommand to be run. Please provide a subcommand.")
	},
}

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Open a note",

	Long: `This command is used to open a note. It requires a noteID to be provided as an argument.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputNoteID := args[0]

		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Errorf("There was an error connecting to the database: %v", err)
		}
		defer conn.Close()

		queries := sqlc.New(conn)
		noteID, err := strconv.Atoi(inputNoteID)
		if err != nil {
			log.Errorf("There was an error converting the noteID to an integer: %v", err)
		}

		editor := utils.GetEditorConfig()

		note, err := queries.ReadNoteByID(ctx, int64(noteID))
		if err != nil {
			log.Errorf("ReadNoteByID: There was an error reading the note: %v", err)
		}
		notePath, err := utils.ExpandPath(note.Path)
		if err != nil {
			log.Fatalf("There was an error expanding the path: %v", err)
		}

		openNoteInEditor(editor, notePath)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.AddCommand(noteCmd)
}

// openNoteInEditor Opens the note in the editor
func openNoteInEditor(editor string, notePath string) {
	editorProcess := exec.Command(editor, notePath)
	editorProcess.Stdin = os.Stdin
	editorProcess.Stdout = os.Stdout
	editorProcess.Stderr = os.Stderr
	err := editorProcess.Run()
	if err != nil {
		log.Fatalf("There was an error running the command: %v", err)
	}
}
