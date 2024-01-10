package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"
)

const (
	instructionComment string = "# Change the names of these files, then save and quit\n\n"
)

var rootCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Bulk rename files and folders",
	Long: `Bulk is a CLI tool that opens a text editor on temporary file that lists out your selected files and allows you to rename them.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		args = splitArgsIfTogether(args)

		currDir, err := os.Getwd()
		if err != nil {
			fmt.Println("Error: getting current directory")
			return
		}

		if !filesValid(args, currDir) {
			fmt.Println("Error: file found that does not exist or is invalid")
			return
		}

		fullPaths := allPathsToFull(args, currDir)
		requireFullPath, err := cmd.Flags().GetBool("full-path")
		if err != nil {
			fmt.Println("Error: could not get --full-path flag")
			return
		}

		if !requireFullPath && filesAllSameDir(fullPaths) {
			for i, path := range fullPaths {
				fullPaths[i] = filepath.Base(path)
			}
		}

		tmpFile, err := os.CreateTemp("", "bulktemp")
		if err != nil {
			fmt.Println("Error: could not create temp file")
			return
		}
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		data := []byte(instructionComment + strings.Join(fullPaths, "\n"))
		if _, err := tmpFile.Write(data); err != nil {
			fmt.Println("Error: could not write to temp file")
			return
		}


		editor := os.Getenv("EDITOR")
		if editor == "" { // temp
			editor = "xdg-open"
		}

		openEditor := exec.Command(editor, tmpFile.Name())
		openEditor.Stdin = os.Stdin
		openEditor.Stdout = os.Stdout
		openEditor.Stderr = os.Stderr
		err = openEditor.Run()
		if err != nil {
			fmt.Println("Error: could not open editor")
			return
		}

		fileContentBytes, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			fmt.Println("Error: could not read temp file")
			return
		}
		fileContent := string(fileContentBytes)

		goodLineCount := 0
		expectedLineCount := len(fullPaths)
		fileContentLines := strings.Split(fileContent, "\n")
		for _, line := range fileContentLines {
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			goodLineCount++
		}
		if goodLineCount != expectedLineCount {
			fmt.Println("Error: number of lines changed")
			return
		}

		newNames := make([]string, expectedLineCount)
		i := 0
		for _, line := range fileContentLines {
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			newNames[i] = line
			i++
		}
	},
}

func splitArgsIfTogether(args []string) []string {
	re := regexp.MustCompile(`(?:[^'"\s]+|'[^']*'|"[^"]*")+`)
	splitArgs := make([]string, 0)
	for _, arg := range args {
		splitArgs = append(splitArgs, re.FindAllString(arg, -1)...)
	}
	return splitArgs
}

func filesValid(fs []string, currDir string) bool {
	for _, f := range fs {
		if !isFullPath(f) {
			f = filepath.Join(currDir, f)
		}
		if _, err := os.Stat(f); err != nil {
			return false
		}
	}
	return true
}

func allPathsToFull(fs []string, currDir string) []string {
	fullFs := make([]string, len(fs))
	for i, f := range fs {
		if !isFullPath(f) {
			f = filepath.Join(currDir, f)
		}
		fullFs[i] = f
	}
	return fullFs
}

func filesAllSameDir(fs []string) bool {
	dir := mapset.NewSet[string]()
	for _, f := range fs {
		dir.Add(filepath.Dir(f))
	}
	return dir.Cardinality() == 1
}

func isFullPath(path string) bool {
	return strings.HasPrefix(path, "/") || strings.HasPrefix(path, "~")
}

func checkLineCount(fileContent string, expectedLineCount int) bool {
	return true
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("dry-run", "d", false, "Dry run")
	rootCmd.Flags().BoolP("full-path", "f", false, "List out full path of files")
}
