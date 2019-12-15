
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

func ServeChangeInstallMethod(w http.ResponseWriter, r *http.Request, baseurl string, installPath string, client_secret string, directory string, foldersMap map[string]string, urlMask string) {

    // generate body of bash case with repo packages
    repoPackages := repoPackagesCasePrint(foldersMap, true, directory, baseurl, installPath)

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
    echo -e "\n\n  About to change install method for all managed files from \e[35mGIT symlinks\e[0m to \e[35mfile copies\e[0m."
else
    echo -e "\n\n  About to change install method for all managed files from \e[35mfile copies\e[0m to \e[35mGIT symlinks\e[0m."
fi
echo -e "\n\n  \e[33;5mWarning!\e[0m This will update all dotfiles managed by dotman."

confirmPrompt

[ -d "$HOME/.dotman/dotfiles" ] && rm -rf "$HOME/.dotman/dotfiles" || mkdir -p "$HOME/.dotman/dotfiles"

if [ -d "$HOME/.dotman/dotfiles" ]; then 
    gitCloneIfPresent "$SECRET"
fi

for NAME in $(cat "$HOME/.dotman/managed"); do
selectPackage $NAME 
done
`
