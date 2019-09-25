
// Chane Install Method view - switches between git and copy files install methods
// =================================================

package views

import (
	"net/http"
    "text/template"
)

type ChangeInstallMethodData struct {
    ClientSecret string
    RepoOpts string
    BaseURL string
    URLMask string
}

func ServeChangeInstallMethod(w http.ResponseWriter, r *http.Request, baseurl string, client_secret string, directory string, foldersMap map[string]string, urlMask string) {

    // generate body of bash case with repo packages
    repoPackages := repoPackagesCasePrint(foldersMap, true, directory, baseurl)

    // build data for template
    data := ChangeInstallMethodData{client_secret, repoPackages, baseurl, urlMask}

    // render template
    tmpl, err := template.New("update").Parse(tmplChangeInstallMethod)
    if err != nil { panic(err) }
    err = tmpl.Execute(w, data)
    if err != nil { panic(err) }
}

var tmplChangeInstallMethod = bashTemplHead + gitCloneTmpl + `
SECRET="{{ .ClientSecret }}"
selectPackage() {
    case "$1" in
    {{ .RepoOpts }}
    esac
}

if [ ! -f "$HOME/.dotman/managed" ]; then 
echo "  It appears, you don't manage any dotfiles using dotman. Exiting."
exit 1
fi

command -v git >/dev/null 2>&1 || echo "  Git command not found. Please install git to change install method." 

if [ -d "$HOME/.dotman/dotfiles" ]; then 
    echo -e "\n\n  About to change install method for all managed files from GIT symlinks to file copies."
else
    echo -e "\n\n  About to change install method for all managed files from file copies to GIT symlinks."
fi

confirmPrompt

[ -d "$HOME/.dotman/dotfiles" ] && rm -rf "$HOME/.dotman/dotfiles" || mkdir -p "$HOME/.dotman/dotfiles"

if [ -d "$HOME/.dotman/dotfiles" ]; then 
    gitCloneIfPresent "$SECRET"
fi

for NAME in $(cat "$HOME/.dotman/managed"); do
selectPackage $NAME 
done
`
