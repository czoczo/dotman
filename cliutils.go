// CLI Utilities
// =================================================

package main

import (
	"os"
	"fmt"
)


// return ASCII logo
func getLogo() string {
    return `
\e[2;35m                $$.  ,$$    $$    $$,   $$
                $$$,.$$$  ,$$$$.  $$$.  $$
\e[0;35m          .$$$. \e[2;35m$$'$$'$$ ,$$.,$$. $$'$$.$$
\e[0;35m          $$$$$ \e[2;35m$$    $$ $$$$$$$$ $$  '$$$
\e[0;35m          '$$$' \e[2;35m$$    $$ $$    $$ $$   '$$\e[0m
`
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[35;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}
