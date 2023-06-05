# pkgcount
A simple tool to count internal &amp; external package use for go projects


Example from this project:

**Internal Package Counts**

| Package        |        Count |
| :---           |         ---: |
|github.com/frisbm/pkgcount/utils |1 |

**External Package Counts**

| Package        |        Count |
| :---           |         ---: |
|strings |2 |
|os |2 |
|log |2 |
|golang.org/x/exp/slices |1 |
|os/exec |1 |
|bytes |1 |
|math |1 |
|sync |1 |
|flag |1 |
|fmt |1 |
|path/filepath |1 |
|regexp |1 |
|text/template |1 |
|bufio |1 |


### Install

    go install github.com/frisbm/pkgcount@latest


### Usage:

	pkgcount [flags]

#### The flags are:

    -h
        Display the help message along with the list of arguments and their descriptions.
    -u
        Set this option to retrieve the final resulting markdown in an unrendered format.
    -o
        Save the output to a file. Please note that the rendered markdown might appear 
        differently in a file. This option is typically used in conjunction with -u.
    -d
        Specify the directory or file path to execute the operation on. By default, 
        it uses the current working directory.
    -lte
        Filter the package counts and only display those with counts less than or 
        equal to the specified integer.
    -gte
        Filter the package counts and only display those with counts greater than or
        equal to the specified integer.
    -exclude
        Enter a regular expression here to exclude specific files, directories,
        or other entities.
