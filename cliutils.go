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

