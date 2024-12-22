package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/akthe-at/go_task/data"
	"github.com/akthe-at/go_task/db"
	"github.com/akthe-at/go_task/sqlc"
	"github.com/akthe-at/go_task/tui"
	"github.com/akthe-at/go_task/tui/formInput"
	"github.com/akthe-at/go_task/utils"
	"github.com/spf13/cobra"
)

var (
	rawFlag      bool
	archived     bool
	NewNote      bool
	noteAliases  string
	noteBody     string
	noteTags     string
	openInEditor bool
)

// addCmd Used for adding new tasks, projects, notes, etc.
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Parent command for adding tasks/projects/notes/etc.",
	Long: `This command is used for adding new tasks, projects, notes, etc. 
	There are subcommands for each of these.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`You invoked the "add" cmd without providing any further subcommands or further arguments,
please complete the command to achieve the desired outcome.`)
	},
}

// addTaskCmd represents the new command
var addTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Creates a new task with a form (or optionally raw string input)",
	Long: `
	Form Example: 'go_task add task' and follow the prompts.
	Raw Example: 'go_task add task <task_title> <task_priority> <task_status>'
	Valid task priorities: low, medium, high, urgent
	Valid task statuses: todo, planning, doing, done
	You can also optionally provided an archived status for the task using the --archived flag.

`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputTitle    = args[0]
			inputPriority = args[1]
			inputStatus   = args[2]
		)

		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()
		queries := sqlc.New(conn)

		if rawFlag {
			validPriority, err := mapToPriorityType(inputPriority)
			if err != nil {
				log.Fatalf("Invalid priority type: %v", err)
			}

			validStatus, err := mapToStatusType(inputStatus)
			if err != nil {
				log.Fatalf("Invalid status type: %v", err)
			}

			newTask := sqlc.CreateTaskParams{
				Title:    inputTitle,
				Priority: sql.NullString{String: string(validPriority), Valid: true},
				Status:   sql.NullString{String: string(validStatus), Valid: true},
				Archived: archived,
			}

			newTaskID, err := queries.CreateTask(ctx, newTask)
			if err != nil {
				log.Fatalf("Error creating task: %v", err)
			}
			fmt.Println("Successfully created a task and it was assigned the following ID: ", newTaskID)

			ok, projectDir, err := utils.CheckIfProjDir()
			if err != nil {
				log.Fatalf("Error while checking if project directory: %v", err)
			}
			if ok {
				projID, err := queries.CheckProgProjectExists(ctx, projectDir)
				if err != nil {
					log.Fatalf("Error while checking if project exists: %v", err)
				} else if projID == 0 {
					project, err := queries.InsertProgProject(ctx, projectDir)
					if err != nil {
						log.Fatalf("Error inserting project: %v", err)
					}
					err = queries.CreateProjectTaskLink(ctx,
						sqlc.CreateProjectTaskLinkParams{
							ProjectID:    sql.NullInt64{Int64: project, Valid: true},
							ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
							ParentTaskID: sql.NullInt64{Int64: newTaskID, Valid: true},
						},
					)
					if err != nil {
						log.Fatalf("Error inserting project link: %v", err)
					}
				}
			}

		} else {
			theme := tui.GetSelectedTheme()
			form := &formInput.NewTaskForm{}

			err := form.NewTaskForm(*tui.ThemeGoTask(theme))
			if err != nil {
				log.Fatalf("Error creating form: %v", err)
			}

			if form.Submit {
				result, err := queries.CreateTask(ctx, sqlc.CreateTaskParams{
					Title:    form.TaskTitle,
					Priority: sql.NullString{String: string(form.Priority), Valid: true},
					Status:   sql.NullString{String: string(form.Status), Valid: true},
				})
				if err != nil {
					log.Fatalf("Error creating task: %v", err)
				}
				fmt.Println("Successfully created a task and it was assigned the following ID: ", result)

				ok, projectDir, err := utils.CheckIfProjDir()
				if err != nil {
					log.Fatalf("Error checking if project directory: %v", err)
				}
				if ok {
					projID, err := queries.CheckProgProjectExists(ctx, projectDir)
					if err != nil {
						log.Fatalf("Error checking if project exists: %v", err)
					} else if projID == 0 {
						project, err := queries.InsertProgProject(ctx, projectDir)
						if err != nil {
							log.Fatalf("Error inserting project: %v", err)
						}
						err = queries.CreateProjectTaskLink(ctx,
							sqlc.CreateProjectTaskLinkParams{
								ProjectID:    sql.NullInt64{Int64: project, Valid: true},
								ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
								ParentTaskID: sql.NullInt64{Int64: result, Valid: true},
							},
						)
						if err != nil {
							log.Fatalf("Error inserting project link: %v", err)
						}
					}

				}

				fmt.Println("Successfully assigned the new task a programming project ID: ", projectDir)
			}
		}
	},
}

// addAreaCmd represents the new command
var addAreaCmd = &cobra.Command{
	Use:   "area",
	Short: "Creates a new area with a form or optionally raw string input",
	Long: `
	The command is used for creating new areas.

	New areas can be created using a form or directly from the command line by passing the --raw or -r flag.
	You can also optionally provide an archived status for the area using the --archived flag.

	Valid area statuses: todo, planning, doing, done

	Raw Example: 'go_task add area "<area_title> <area_status>"'
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputTitle  = args[0]
			inputStatus = args[1]
		)
		ctx := context.Background()
		conn, err := db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close()

		queries := sqlc.New(conn)

		if rawFlag {

			if len(args) < 2 {
				log.Fatalf(`
You must provide at least 2 arguments!
Usage: add area <area_title> <area_status>
if one of your arguments has white space, please wrap it in "" marks.`)
			}
			validStatus, err := mapToStatusType(inputStatus)
			if err != nil {
				log.Fatalf("Invalid status type: %v", err)
			}

			_, err = queries.CreateArea(ctx, sqlc.CreateAreaParams{
				Title:    inputTitle,
				Status:   sql.NullString{String: string(validStatus), Valid: true},
				Archived: archived,
			},
			)
			if err != nil {
				log.Fatalf("Error creating new area: %v", err)
			} else {
				fmt.Println("Successfully created a new area")
			}

		} else {
			form := &formInput.NewAreaForm{}
			theTheme := tui.GetSelectedTheme()

			err := form.NewAreaForm(*tui.ThemeGoTask(theTheme))
			if err != nil {
				log.Fatalf("Error creating form: %v", err)
			}

			if form.Submit {
				_, err = queries.CreateArea(ctx, sqlc.CreateAreaParams{
					Title:    form.AreaTitle,
					Status:   sql.NullString{String: string(form.Status), Valid: true},
					Archived: form.Archived,
				})
				if err != nil {
					log.Fatalf("AddAreaCmd: Error creating task: %v", err)
				} else {
					fmt.Println("Successfully created a new area")
				}
			}
		}
	},
}

// addTaskNoteCmd represents the new command
var addTaskNoteCmd = &cobra.Command{
	Use:   "note",
	Short: "Add a note to a specific task",
	Long: `
To add a new note to an existing task, you can use the 'note' command.
This command will create a new note in the obsidian vault and create a bridge between the note and the task in the database.

To do this:
	Type in: 'go_task add task note <task_id> <note_title> <note_path>'
OR to generate a new note AND add it to a specific task:
	Type in: 'go_task add task note <task_id> <note_title> -t <note_tags> -a <note_aliases> -b <note_body>'
`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputTaskID    = args[0]
			inputNoteTitle = args[1]
			inputNotePath  = args[2]
		)

		taskID, err := strconv.Atoi(inputTaskID)
		if err != nil {
			log.Fatalf("Invalid task ID: %v", err)
		}

		if NewNote {
			if len(args) < 2 {
				log.Fatalf("You must provide at least 2 arguments to generate a new note! Usage: note <task_id> <note_title> -t <note_tags> -a <note_aliases> -b <note_body>")
			}
			fmt.Println("Creating a new note for task: ", inputTaskID)
			theTags := strings.Split(noteTags, " ")
			theAliases := strings.Split(noteAliases, " ")

			newNoteID := data.GenerateNoteID(inputNoteTitle)
			outputPath, err := data.TemplateMarkdownNote(inputNoteTitle, newNoteID, noteBody, theAliases, theTags)
			if err != nil {
				log.Fatal("Error with generating Template!", err)
			}
			if openInEditor {
				editor := GetEditorConfig()
				cmdr := exec.Command(editor, outputPath)
				cmdr.Stdin = os.Stdin
				cmdr.Stdout = os.Stdout
				cmdr.Stderr = os.Stderr
				err = cmdr.Run()
				if err != nil {
					log.Fatalf("There was an error running the command: %v", err)
				}
			}

			ctx := context.Background()
			conn, err := db.ConnectDB()
			if err != nil {
				log.Fatalf("Error connecting to database: %v", err)
			}

			tx, err := conn.Begin()
			if err != nil {
				log.Fatalf("addTaskNoteCmd: Error beginning transaction: %v", err)
			}

			defer tx.Rollback()
			defer conn.Close()

			queries := sqlc.New(conn)
			qtx := queries.WithTx(tx)
			noteID, err := qtx.CreateNote(ctx, sqlc.CreateNoteParams{
				Title: inputNoteTitle,
				Path:  outputPath,
			})
			if err != nil {
				log.Fatalf("addTaskNoteCmd: There was an error creating the note: %v", err)
			}

			_, err = qtx.CreateTaskBridgeNote(ctx, sqlc.CreateTaskBridgeNoteParams{
				NoteID:       sql.NullInt64{Int64: noteID, Valid: true},
				ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
				ParentTaskID: sql.NullInt64{Int64: int64(taskID), Valid: true},
			})
			if err != nil {
				log.Fatalf("addTaskNoteCmd: Error creating task bridge note: %v", err)
			}
			err = tx.Commit()
			if err != nil {
				log.Fatalf("addTaskNoteCmd: Error committing transaction: %v", err)
			}
			fmt.Println("Note added to task successfully")

			ok, projectDir, err := utils.CheckIfProjDir()
			if err != nil {
				log.Fatalf("Error while checking if in a project directory: %v", err)
			}
			if ok {
				projID, err := queries.CheckProgProjectExists(ctx, projectDir)
				if err != nil {
					log.Fatalf("Error while checking if project exists: %v", err)
				}
				if projID == 0 {
					projID, err = queries.InsertProgProject(ctx, projectDir)
					if err != nil {
						log.Fatalf("Error inserting project: %v", err)
					}
				}

				err = queries.CreateProjectTaskLink(ctx,
					sqlc.CreateProjectTaskLinkParams{
						ProjectID:    sql.NullInt64{Int64: projID, Valid: true},
						ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
						ParentTaskID: sql.NullInt64{Int64: int64(taskID), Valid: true},
					})
				if err != nil {
					log.Fatalf("Error inserting project link: %v", err)
				}
			}
		} else {
			if len(args) < 3 {
				log.Fatalf("You must provide at least 3 arguments! Usage: note <task_id> <note_title> <note_path>")
			}

			ctx := context.Background()
			conn, err := db.ConnectDB()
			if err != nil {
				log.Fatalf("Error connecting to database: %v", err)
			}
			tx, err := conn.Begin()
			if err != nil {
				log.Fatalf("addTaskNoteCmd: Error beginning transaction: %v", err)
			}
			defer tx.Rollback()
			defer conn.Close()

			queries := sqlc.New(conn)
			qtx := queries.WithTx(tx)
			noteID, err := qtx.CreateNote(ctx, sqlc.CreateNoteParams{
				Title: inputNoteTitle,
				Path:  inputNotePath,
			})
			if err != nil {
				fmt.Printf("addTaskNoteCmd: There was an error creating the note: %v", err)
			}

			_, err = qtx.CreateTaskBridgeNote(ctx, sqlc.CreateTaskBridgeNoteParams{
				NoteID:       sql.NullInt64{Int64: noteID, Valid: true},
				ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
				ParentTaskID: sql.NullInt64{Int64: int64(taskID), Valid: true},
			})
			if err != nil {
				log.Fatalf("addTaskNoteCmd: Error creating task bridge note: %v", err)
			}

			ok, projectDir, err := utils.CheckIfProjDir()
			if err != nil {
				log.Fatalf("Error while checking if project directory: %v", err)
			}
			if ok {
				projID, err := queries.CheckProgProjectExists(ctx, projectDir)
				if err != nil {
					log.Fatalf("Error while checking if project exists: %v", err)
				}
				if projID == 0 {
					project, err := queries.InsertProgProject(ctx, projectDir)
					if err != nil {
						log.Fatalf("Error inserting project: %v", err)
					}
					err = queries.CreateProjectTaskLink(ctx,
						sqlc.CreateProjectTaskLinkParams{
							ProjectID:    sql.NullInt64{Int64: project, Valid: true},
							ParentCat:    sql.NullInt64{Int64: int64(data.TaskNoteType), Valid: true},
							ParentTaskID: sql.NullInt64{Int64: int64(taskID), Valid: true},
						})
					if err != nil {
						log.Fatalf("Error inserting project link: %v", err)
					}
				}
			}
			err = tx.Commit()
			if err != nil {
				log.Fatalf("addTaskNoteCmd: Error committing transaction: %v", err)
			} else {
				fmt.Println("Note added to task successfully")
			}
		}
	},
}

// addAreaNoteCmd represents the command for adding new Project/Area notes to an existing Project/Area
var addAreaNoteCmd = &cobra.Command{
	Use:   "note",
	Short: "Add a note to a specific area.",
	Long: `
To add a new note to an existing area, you can use the 'note' command.
This command will create a new note in the obsidian vault and create a bridge between the note and the area in the database.

To do this:
	Type in: 'go_task add area note <area_id> <note_title> <note_path>'
OR to generate a new note AND add it to a specific area:
	Type in: 'go_task add area note <area_id> <note_title> -t <note_tags> -a <note_aliases> -b <note_body>'
`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputAreaID    = args[0]
			inputNoteTitle = args[1]
			inputNotePath  = args[2]
		)

		areaID, err := strconv.Atoi(inputAreaID)
		if err != nil {
			log.Fatalf("Invalid area ID: %v", err)
		}

		if NewNote {
			if len(args) < 2 {
				log.Fatalf("You must provide at least 2 arguments to generate a new note! Usage: note <area_id> <note_title> -t <note_tags> -a <note_aliases> -b <note_body>")
			}
			fmt.Println("Creating a new note for area: ", inputAreaID)
			theTags := strings.Split(noteTags, " ")
			theAliases := strings.Split(noteAliases, " ")

			newNoteID := data.GenerateNoteID(inputNoteTitle)
			outputPath, err := data.TemplateMarkdownNote(inputNoteTitle, newNoteID, noteBody, theAliases, theTags)
			if err != nil {
				log.Fatal("An error occurred while generating the template: ", err)
			}

			if openInEditor {
				editor := GetEditorConfig()
				cmdr := exec.Command(editor, outputPath)
				cmdr.Stdin = os.Stdin
				cmdr.Stdout = os.Stdout
				cmdr.Stderr = os.Stderr
				err := cmdr.Run()
				if err != nil {
					log.Fatalf("There was an error running the command: %v", err)
				}
			}
			ctx := context.Background()
			conn, err := db.ConnectDB()
			if err != nil {
				log.Fatalf("Error connecting to database: %v", err)
			}

			tx, err := conn.Begin()
			if err != nil {
				log.Fatalf("addAreaNoteCmd: Error beginning transaction: %v", err)
			}

			defer tx.Rollback()
			defer conn.Close()

			queries := sqlc.New(conn)
			qtx := queries.WithTx(tx)
			noteID, err := qtx.CreateNote(ctx, sqlc.CreateNoteParams{
				Title: inputNoteTitle,
				Path:  outputPath,
			})
			if err != nil {
				log.Fatalf("addAreaNoteCmd: There was an error creating the note: %v", err)
			}

			_, err = qtx.CreateAreaBridgeNote(ctx, sqlc.CreateAreaBridgeNoteParams{
				NoteID:       sql.NullInt64{Int64: noteID, Valid: true},
				ParentCat:    sql.NullInt64{Int64: int64(data.AreaNoteType), Valid: true},
				ParentAreaID: sql.NullInt64{Int64: int64(areaID), Valid: true},
			})
			if err != nil {
				log.Fatalf("addAreaNoteCmd: Error creating area bridge note: %v", err)
			}
			err = tx.Commit()
			if err != nil {
				log.Fatalf("addAreaNoteCmd: Error committing transaction: %v", err)
			}
			fmt.Println("Note added to Area successfully")

			ok, projectDir, err := utils.CheckIfProjDir()
			if err != nil {
				log.Fatalf("Error while checking if in a project directory: %v", err)
			}
			if ok {
				projID, err := queries.CheckProgProjectExists(ctx, projectDir)
				if err != nil {
					log.Fatalf("Error while checking if project exists: %v", err)
				}
				if projID == 0 {
					projID, err = queries.InsertProgProject(ctx, projectDir)
					if err != nil {
						log.Fatalf("Error inserting project: %v", err)
					}
				}

				err = queries.CreateProjectAreaLink(ctx, sqlc.CreateProjectAreaLinkParams{
					ProjectID:    sql.NullInt64{Int64: projID, Valid: true},
					ParentCat:    sql.NullInt64{Int64: int64(data.AreaNoteType), Valid: true},
					ParentAreaID: sql.NullInt64{Int64: int64(areaID), Valid: true},
				})
				if err != nil {
					log.Fatalf("Error inserting project link: %v", err)
				}
			}
		} else {
			if len(args) < 3 {
				log.Fatalf("You must provide at least 3 arguments! Usage: add area note <area_id> <note_title> <note_path>")
			}

			ctx := context.Background()
			conn, err := db.ConnectDB()
			if err != nil {
				log.Fatalf("Error connecting to database: %v", err)
			}
			tx, err := conn.Begin()
			if err != nil {
				log.Fatalf("addAreaNoteCmd: Error beginning transaction: %v", err)
			}

			defer tx.Rollback()
			defer conn.Close()

			queries := sqlc.New(conn)
			qtx := queries.WithTx(tx)
			noteID, err := qtx.CreateNote(ctx, sqlc.CreateNoteParams{
				Title: inputNoteTitle,
				Path:  inputNotePath,
			})
			if err != nil {
				fmt.Printf("addAreaNoteCmd: There was an error creating the note: %v", err)
			}

			_, err = qtx.CreateAreaBridgeNote(ctx, sqlc.CreateAreaBridgeNoteParams{
				NoteID:       sql.NullInt64{Int64: noteID, Valid: true},
				ParentCat:    sql.NullInt64{Int64: int64(data.AreaNoteType), Valid: true},
				ParentAreaID: sql.NullInt64{Int64: int64(areaID), Valid: true},
			})
			if err != nil {
				log.Fatalf("addAreaNoteCmd: Error creating task bridge note: %v", err)
			}

			ok, projectDir, err := utils.CheckIfProjDir()
			if err != nil {
				log.Fatalf("Error while checking if in a project directory: %v", err)
			}
			if ok {
				projID, err := queries.CheckProgProjectExists(ctx, projectDir)
				if err != nil {
					log.Fatalf("Error while checking if project exists: %v", err)
				}
				if projID == 0 {
					project, err := queries.InsertProgProject(ctx, projectDir)
					if err != nil {
						log.Fatalf("Error while inserting project: %v", err)
					}
					err = queries.CreateProjectAreaLink(ctx,
						sqlc.CreateProjectAreaLinkParams{
							ProjectID:    sql.NullInt64{Int64: project, Valid: true},
							ParentCat:    sql.NullInt64{Int64: int64(data.AreaNoteType), Valid: true},
							ParentAreaID: sql.NullInt64{Int64: int64(areaID), Valid: true},
						})
					if err != nil {
						log.Fatalf("Error inserting project link: %v", err)
					}
				}
			}
			err = tx.Commit()
			if err != nil {
				log.Fatalf("addAreaNoteCmd: Error committing transaction: %v", err)
			} else {
				fmt.Println("Note added to Area successfully")
			}
		}
	},
}

func init() {
	// root commands
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addTaskCmd)
	addCmd.AddCommand(addAreaCmd)
	// subcommands
	addTaskCmd.AddCommand(addTaskNoteCmd)
	addAreaCmd.AddCommand(addAreaNoteCmd)
	// flags
	addCmd.PersistentFlags().BoolVarP(&rawFlag, "raw", "r", false, "Bypass using the form and use raw input instead")
	addCmd.PersistentFlags().BoolVar(&archived, "archived", false, "Archive the task or area upon creation")
	addCmd.PersistentFlags().BoolVar(&NewNote, "new", false, "this flag is used to add a new note to an existing task or area")
	addCmd.PersistentFlags().BoolVar(&openInEditor, "open", false, "this flag is used to open the note in an editor after creation")
	addCmd.PersistentFlags().StringVarP(&noteTags, "tags", "t", "", "Tags for the note")
	addCmd.PersistentFlags().StringVarP(&noteAliases, "aliases", "a", "", "Aliases for the note")
	addCmd.PersistentFlags().StringVarP(&noteBody, "body", "b", "", "Text for the Note Body")
}

// mapToPriorityType maps a string to a PriorityType
func mapToPriorityType(input string) (data.PriorityType, error) {
	switch input {
	case "low":
		return data.PriorityTypeLow, nil
	case "medium":
		return data.PriorityTypeMedium, nil
	case "high":
		return data.PriorityTypeHigh, nil
	case "urgent":
		return data.PriorityTypeUrgent, nil
	default:
		return "", fmt.Errorf("invalid priority type ( %v ) is not one of the valid priority values (low, medium, high, urgent)", input)
	}
}

// mapToStatusType maps a string to a StatusType
func mapToStatusType(input string) (data.StatusType, error) {
	switch input {
	case "todo":
		return data.StatusToDo, nil
	case "planning":
		return data.StatusPlanning, nil
	case "doing":
		return data.StatusDoing, nil
	case "done":
		return data.StatusDone, nil
	default:
		return "", fmt.Errorf("invalid status type ( %v ) is not one of the valid status values (todo, planning, doing, done)", input)
	}
}
