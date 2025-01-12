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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/akthe-at/go_task/config"
	"github.com/akthe-at/go_task/tui"
	dataTable "github.com/akthe-at/go_task/tui/dataTable"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go_task",
	Short: "This launches the go_task TUI application",
	Long: `This application consists of a TUI interface and a CLI interface
	To launch the TUI version, simpy run go_task with no arguments. All other subcommands
	interact with the CLI version of the application`,
	Run: func(cmd *cobra.Command, args []string) {
		tui.ClearTerminalScreen()
		model := dataTable.NewRootModel()
		p := tea.NewProgram(&model)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/go_task/config.toml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("initConfig: Error getting user home directory: %v", err)
	}

	var configPath string
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		switch runtime.GOOS {
		case "windows":
			if windowsConfigDir := os.Getenv("LOCALAPPDATA"); windowsConfigDir != "" {
				configPath = filepath.Join(windowsConfigDir, "go_task", "config")
			} else {
				configPath = filepath.Join(home, ".config", "go_task")
			}
		case "linux":
			if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
				configPath = filepath.Join(xdgConfig, "go_task")
			} else {
				configPath = filepath.Join(home, ".config", "go_task")
			}
		}
		err := os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			log.Fatalf("failed to create directory: %v", err)
		}
	}

	// TODO: This needs a better name.
	viper.AddConfigPath(configPath)
	viper.SetConfigType("toml")
	viper.SetConfigName("config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		return
	}

	// FIXME: I wish this wasn't necessary but viper wouldn'tm unmarshal the data otherwise?
	// Directly access the values from Viper and manually assign them to the struct
	config.UserSettings.Selected.Editor = viper.GetString("selected.editor")
	config.UserSettings.Selected.NotesPath = viper.GetString("selected.notes_path")
	config.UserSettings.Selected.UseObsidian = viper.GetBool("selected.use_obsidian")
	config.UserSettings.Selected.Theme = viper.GetString("selected.theme")

	var userThemes tui.ColorThemes
	if err := viper.Unmarshal(&userThemes); err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
		return
	}
	mergeThemes(userThemes)
}

func mergeThemes(userThemes tui.ColorThemes) {
	if userThemes.RosePine != (tui.Theme{}) {
		tui.Themes.RosePine = userThemes.RosePine
	}
	if userThemes.RosePineMoon != (tui.Theme{}) {
		tui.Themes.RosePineMoon = userThemes.RosePineMoon
	}
	if userThemes.RosePineDawn != (tui.Theme{}) {
		tui.Themes.RosePineDawn = userThemes.RosePineDawn
	}
	for name, theme := range userThemes.Additional {
		tui.Themes.Additional[name] = theme
	}
	if userThemes.Selected.Theme != "" {
		tui.Themes.Selected = userThemes.Selected
	}
}
