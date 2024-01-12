# Bulk
Simple command-line tool for bulk renaming files and folders inspired by [ranger](https://github.com/ranger/ranger).

## Usage
Call bulk with the files to rename as arguments.

```{sh}
bulk file.txt another_file.txt yet_another_file.txt
```

bulk will open a text editor, first by checking your `$EDITOR` environmental variable, then `xdg-open`. The editor can change then name of each file.

Another file will be opened. This file will contain the shell commands to rename your files to their new names. Upon closing this file, each file will be renamed.

Works well when used with [nav](https://github.com/lezhou8/nav).

<img src="assets/nav_bulk.gif">

## Install

```{sh}
go install github.com/lezhou8/bulk@latest
```

## Dependencies
- [Go](https://golang.org/)

## Built with
- [Go](https://golang.org/)
- [cobra](https://github.com/spf13/cobra)
- [golang-set](https://github.com/deckarep/golang-set)
