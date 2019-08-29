
// Install view - shows install menu, and executes install upon chosen options
// =================================================

package routes

import (
	"fmt"
	"net/http"
)

func ServeUpdate(w http.ResponseWriter, r *http.Request, baseurl string, client_secret string, logo string, directory string, alphabet string, foldersMap map[string]string) {

    // print case function
//            fmt.Fprint(w,"tput clear\n")
    fmt.Fprint(w,"SECRET=\""+client_secret+"\"\n")
    fmt.Fprint(w,"selectOption() {\n    case \"$1\" in\n")
    repoOptsCasePrint(w, foldersMap, true, directory, baseurl)

    // close case 
    fmt.Fprint(w,"    esac\n}\n")
    fmt.Fprint(w,`
if [ ! -f "$HOME/.dotman/managed" ]; then 
echo "It appears, you don't manage any dotfiles using dotman. Exiting."
exit 1
fi

for NAME in $(cat "$HOME/.dotman/managed"); do
selectOption $NAME 
done
    `)
}
