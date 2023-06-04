# pkgcount
A simple tool to count internal &amp; external package use for go projects


Example from this project:


**Internal Package Counts**

| Package | Count |
| :---  |  ---: |
| github.com/frisbm/pkgcount/utils | 1 |

**External Package Counts**

| Package | Count |
| :---  |  ---: |
| strings | 2 |
| log | 2 |
| os | 2 |
| os/exec | 1 |
| bufio | 1 |
| fmt | 1 |
| golang.org/x/exp/slices | 1 |
| path/filepath | 1 |
| flag | 1 |
| math | 1 |
| regexp | 1 |
| sync | 1 |


### install

    go install github.com/frisbm/pkgcount@latest


### Usage:

	pkgcount [flags]

#### The flags are:

	-h
	    Print help message with args and descriptions
	-u
	    Set to make the final resulting markdown return unrendered
	-o
		Output to a file, the rendered markdown looks kind of funny in a file, will
		most likely be used in conjuction with -u
	-d
		Which director/filepath to run on, defaults to current working directory
	-exclude
		Supply a Regular Expression here to exclude certain files, directories, etc.
	-lte
		Will filter the counts of the packages and only return those where counts are
		less than or equal to the int you supply
	-gte
		Will filter the counts of the packages and only return those where counts are
		greater than or equal to the int you supply
