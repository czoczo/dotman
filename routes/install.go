
// Install view - shows install menu, and executes install upon chosen options
// =================================================

package routes

import (
	"fmt"
    "strings"
	"net/http"
)

func ServeInstall(w http.ResponseWriter, r *http.Request, baseurl string, client_secret string, logo string, directory string, alphabet string, foldersMap map[string]string) {

    // start with shebang
    fmt.Fprint(w,`#!/bin/bash
tput clear
`)
    // print ASCII logo
    fmt.Fprint(w,"echo -e '"+strings.ReplaceAll(logo,"'","'\"'\"'")+"'\n")
    fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0;37m\"\n")

    // print menu
    fmt.Fprint(w,"printf \"%2s%s\\n%2s%s\\e[32m%s\\e[0;37m%s\\n\\n\" \"\" \"Choose dotfiles to be installed.\" \"\" \"Select by typing keys (\" \"green\" \") and confirm with enter.\"\n")
    fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0m\"\n")

    // vars for iterating available options
    count := 0
    nl := ""

    // iterate over options
    for i := 0; i < len(alphabet); i++ {

            // exit if no keys in map anymore
            key := string(alphabet[i])
            if _, ok := foldersMap[key]; !ok {
                continue
            }

            // handle deviding menu in 3 columns
            count += 1
            if count % 3 == 0 {
                nl = "\\n"
            }

            // print menu option
            fmt.Fprint(w,"printf \"  \\e[32m%s\\e[0m)\\e[35m %-15s\\e[0m"+nl+"\" \""+key+"\" \""+foldersMap[key]+"\" \n")
            nl = ""
    }

    // touch list of dotfiles to be update
    fmt.Fprint(w,"mkdir -p \"$HOME/.dotman\"; touch \"$HOME/.dotman/managed\"\n")

    // print case function
    fmt.Fprint(w,"SECRET=\""+client_secret+"\"\n")
    fmt.Fprint(w,"selectOption() {\n    case \"$1\" in\n")
    repoOptsCasePrint(w, foldersMap, false, directory, baseurl)

    // close case 
    fmt.Fprint(w,"    esac\n}\n")

    // print variable with available options for cross checking input
    keys := make([]string, 0)
    for key := range foldersMap {
        keys = append(keys, key)
    }
    fmt.Fprint(w,"OPTS=\""+strings.Join(keys, "")+"\"\n")

    // print rest of script
    fmt.Fprint(w,`
exec 3<>/dev/tty
echo ""
read -u 3 -p "  Chosen options: " words
echo ""
if [ -z $words ]; then
echo -e "\e[0;37mNothing to do... exiting."
exit 0
fi
echo -e "\e[97m-========================================================-\e[0;37m"
printf "%2s\n" "" "Follwing dotfiles will be installed in order:"
COMMA=""
for CHAR in $(echo "$words" | fold -w1); do
test "${OPTS#*$CHAR}" != "$OPTS" || continue
echo -en "$COMMA" 
selectOption $CHAR False
COMMA=", "
done

if [ "$COMMA" == "" ]; then
echo "Nothing to do... exiting."
exit 0
fi

printf "\n%2s" "" "Proceed? [y/N]"
read -u 3 -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
exit 0
fi

echo -e "\\n\e[97m-========================================================-\e[0m\n"
echo "Installing dotfiles:"

for CHAR in $(echo "$words" | fold -w1); do
test "${OPTS#*$CHAR}" != "$OPTS" || continue
selectOption $CHAR 
done
    `)
}
