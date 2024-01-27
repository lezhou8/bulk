package cmd

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"
)

const instructionComment = "#!/bin/sh\n# This file will execute when the file closes\n"

var rootCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Bulk rename files and folders",
	Long: `Bulk is a CLI tool that opens a text editor on temporary file that lists out your selected files and allows you to rename them.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		currDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		if !filesValid(args, currDir) {
			log.Fatal("Error: invalid files")
		}

		fullPaths := allPathsToFull(args, currDir)
		requireFullPath, err := cmd.Flags().GetBool("full-path")
		if err != nil {
			log.Fatal(err)
		}

		if !requireFullPath && filesAllSameDir(fullPaths, currDir) {
			for i, path := range fullPaths {
				fullPaths[i] = filepath.Base(path)
			}
		}

		tmpFile, err := os.CreateTemp("", "bulktemp")
		if err != nil {
			log.Fatal(err)
		}
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(strings.Join(fullPaths, "\n")); err != nil {
			log.Fatal(err)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "xdg-open"
		}

		openEditor := exec.Command(editor, tmpFile.Name())
		openEditor.Stdin = os.Stdin
		openEditor.Stdout = os.Stdout
		openEditor.Stderr = os.Stderr
		err = openEditor.Run()
		if err != nil {
			log.Fatal(err)
		}

		fileContentBytes, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			log.Fatal(err)
		}
		fileContent := string(fileContentBytes)

		expectedLineCount := len(fullPaths)
		fileContentLines := strings.Split(fileContent, "\n")
		if !lineCountGood(fileContentLines, expectedLineCount) {
			log.Fatal(err)
		}

		newNames := createNewNames(fileContentLines, expectedLineCount)
		showCmdsLines := createShowCmdsLines(fullPaths, newNames)

		tmpFileCmds, err := os.CreateTemp("", "bulkCmds")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(tmpFileCmds.Name())

		if _, err := tmpFileCmds.WriteString(instructionComment + strings.Join(showCmdsLines, "\n")); err != nil {
			log.Fatal(err)
		}
		tmpFileCmds.Close()

		openEditor = exec.Command(editor, tmpFileCmds.Name())
		openEditor.Stdin = os.Stdin
		openEditor.Stdout = os.Stdout
		openEditor.Stderr = os.Stderr
		err = openEditor.Run()
		if err != nil {
			log.Fatal(err)
		}
		err = os.Chmod(tmpFileCmds.Name(), 0755)
		if err != nil {
			log.Fatal(err)
		}

		isDryRun, err := cmd.Flags().GetBool("dry-run")
		if isDryRun {
			return
		}

		runCmds := exec.Command(tmpFileCmds.Name())
		runCmds.Stdin = os.Stdin
		runCmds.Stdout = os.Stdout
		runCmds.Stderr = os.Stderr
		err = runCmds.Run()
		if err != nil {
			log.Fatal(err)
		}
	},
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

func filesAllSameDir(fs []string, currDir string) bool {
	dir := mapset.NewSet[string]()
	for _, f := range fs {
		dir.Add(filepath.Dir(f))
	}
	return dir.Cardinality() == 1 && dir.Contains(currDir)
}

func isFullPath(path string) bool {
	return strings.HasPrefix(path, "/") || strings.HasPrefix(path, "~")
}

func lineCountGood(fileContentLines []string, expectedLineCount int) bool {
	goodLineCount := 0
	for _, line := range fileContentLines {
		if !lineToBeCounted(line) {
			continue
		}
		goodLineCount++
	}
	if goodLineCount != expectedLineCount {
		return false
	}
	return true
}

func lineToBeCounted(line string) bool {
	return !strings.HasPrefix(line, "#") && line != ""
}

func createNewNames(fileContentLines []string, expectedLineCount int) []string {
	newNames := make([]string, expectedLineCount)
	i := 0
	for _, line := range fileContentLines {
		if !lineToBeCounted(line) {
			continue
		}
		line = removeComment(line)
		newNames[i] = line
		i++
	}
	return newNames
}

func removeComment(line string) string {
	if strings.Contains(line, "#") {
		line = strings.Split(line, "#")[0]
	}
	return line
}

func createShowCmdsLines(fullPaths, newNames []string) []string {
	expectedLineCount := len(fullPaths)
	showCmdsLines := make([]string, expectedLineCount)
	for i := 0; i < expectedLineCount; i++ {
		showCmdsLines[i] = "mv -vi -- \"" + fullPaths[i] + "\" \"" + newNames[i] + "\""
	}
	return showCmdsLines
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
