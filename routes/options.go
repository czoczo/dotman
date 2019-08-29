
// Options view - used by other views to generate options menu or install comands
// =================================================

package routes

import (
	"fmt"
    "log"
    "strings"
	"net/http"
	"os"
    "path/filepath"
)

// print bash CASE operator body with available dotfiles folders as options.
// This CASE operator is used to eihter print options for menu, or output command for installing dotfiles
func repoOptsCasePrint(w http.ResponseWriter, foldersMap map[string]string, byName bool, directory string, baseurl string )  {

    // itertate over available option
    for key, val := range foldersMap {
            index := key

            // map either by alphabet char, or folder name
            if byName { index = val }

            // print bash condition. If $2 passed, print menu option and return
            fmt.Fprint(w,"        "+index+")\n            if [ \"$2\" ]; then\n                printf \"\\e[0;32m*\\e[0m)\\e[0;35m %s\\e[0m\" \""+val+"\"\n                return\n            fi\n")

            // else print download commands
            fmt.Fprint(w,"             echo -e \"Installing \\e[0;35m"+val+"\\e[0m\"\n")

            // search folders and add mkdir and download commands
            // recursive walk thorough the dir
            err := filepath.Walk(directory+"/"+val,
                func(path string, info os.FileInfo, err error) error {
                if err != nil {
                    return err
                }

                // skip if not a dir
                if info.IsDir() {
                    return nil
                }

                // if not, strip absolute path and filename. Print mkdir commands
                output := strings.TrimPrefix(path, directory+"/"+val+"/")
                if filepath.Dir(output) != "." {
                    fmt.Fprint(w,"             mkdir -p $HOME/" + filepath.Dir(output)+"\n")
                }

                // print download commands
                fmt.Fprint(w,"             echo -n \"downloading file - "+output+" : \"\n")
                fmt.Fprint(w,"             curl -sH\"secret:$SECRET\" \"" + baseurl + "/" + path + "\" > \"$HOME/" + output+"\"\n")
                fmt.Fprint(w,"             RESULT=$?; [ $RESULT -eq 0 ] && echo -e \"\\e[0;32mok\\e[0m\" || echo -e \"\\e[0;31merror\\e[0m\"\n")

                // if not present, add option to managed dotfiles list
                fmt.Fprint(w,"             cat \"$HOME/.dotman/managed\" | grep -q \""+val+"\" || echo \""+val+"\" >> \"$HOME/.dotman/managed\" \n")
                return nil
            })

            if err != nil {
                log.Println(err)
            }
            fmt.Fprint(w,"             ;;\n")
    }
}
