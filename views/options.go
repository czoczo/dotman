
// Options view - used by other views to generate options menu or install comands
// =================================================

package views

import (
    "log"
    "strings"
    "os"
    "path/filepath"
)

// checking if file exists and is not a directory
func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

// print bash CASE operator body with available dotfiles folders as options.
// This CASE operator is used to eihter print options for menu, or output command for installing dotfiles
func repoPackagesCasePrint(foldersMap map[string]string, byName bool, directory string, baseurl string, installPath string ) string {

    result := ""

    // itertate over available option
    for key, val := range foldersMap {
            index := key

            // map either by alphabet char, or folder name
            if byName { index = val }

            // print bash condition. If $2 passed, print menu option and return
            result = result + "        "+index+")\n            if [ \"$2\" ]; then\n                printf \"\\e[0;32m*\\e[0m)\\e[0;35m %s\\e[0m\" \""+val+"\"\n                return\n            fi\n"

            // else print download commands
            result = result + "             echo -e \"\\n  Installing \\e[0;35m"+val+"\\e[0m\"\n"

            // search folders and add mkdir and download commands
            // recursive walk thorough the dir
            err := filepath.Walk(directory+"/"+val, func(path string, info os.FileInfo, err error) error {
                if err != nil {
                    return err
                }

                // skip if not a dir
                if info.IsDir() {
                    return nil
                }

                // skip if dotautorun.sh file
                if strings.HasSuffix(path, "dotautorun.sh") {
                    return nil
                }

                // if not, strip absolute path and filename. Print mkdir commands
                output := strings.TrimPrefix(path, directory+"/"+val+"/")
                if filepath.Dir(output) != "." {
                    result = result + "             mkdir -p " + installPath + "/" + filepath.Dir(output)+"\n"
                }

                // print download commands
                result = result + "             [ -d \"$HOME/.dotman/dotfiles\" ] && echo -n \"  symlinking file\" || echo -n \"  downloading file\"\n"
                result = result + "             echo -n \" - "+output+" : \"\n"
//                result = result + "             curl -sH\"secret:$SECRET\" \"" + baseurl + "/" + path + "\" > \"$HOME/" + output+"\"\n"
                result = result + "             if [ -d \"$HOME/.dotman/dotfiles\" ]; then\n"
                result = result + "                 ln -sf \"$HOME/.dotman/" + directory + "/" + val + "/" + output + "\" \"" + installPath + "/" + output+"\"\n             else\n"
                result = result + "                 test -h \"" + installPath + "/" + output+"\" && rm \"" + installPath + "/" + output+"\"\n                 curl -sH\"secret:$SECRET\" \"" + baseurl + "/" + path + "\" > \"" + installPath + "/" + output+"\"\n             fi\n"
                result = result + "             RESULT=$?; [ $RESULT -eq 0 ] && echo -e \"\\e[0;32mok\\e[0m\" || echo -e \"\\e[0;31merror\\e[0m\"\n"

                // if not present, add option to managed dotfiles list
                result = result + "             touch \"$HOME/.dotman/managed\"; cat \"$HOME/.dotman/managed\" | grep -q \""+val+"\" || echo \""+val+"\" >> \"$HOME/.dotman/managed\" \n"
                return nil
            })

            // if dotautorun.sh present, run it
            if fileExists(directory + "/" + val + "/dotautorun.sh") {
                result = result + "             echo -n \"  dotautorun.sh found. Running dotautorun.sh : \"\n"
                result = result + "             curl -sH\"secret:$SECRET\" \"" + baseurl + "/" + directory + "/" + val + "/dotautorun.sh\" | bash - \n"
                //result = result + "             source /dev/stdin <<< \"$(curl -H\"SECRET:$SECRET\" \"" + baseurl + "/" + directory + "/" + val + "/dotautorun.sh)\"\n"
                result = result + "             RESULT=$?; [ $RESULT -eq 0 ] && echo -e \"\\e[0;32mok\\e[0m\" || echo -e \"\\e[0;31merror\\e[0m\"\n"
            }

            if err != nil {
                log.Println(err)
            }
            result = result + "             ;;\n"
    }
    return result

}
