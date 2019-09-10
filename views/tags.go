package views


import (
    "strings"
	"net/http"
    "text/template"
)

type TagsInstallData struct {
    ClientSecret string
    RepoOpts string
    Packages string
}

func ServeTags(w http.ResponseWriter, r *http.Request, baseurl string, secret string, logo string, directory string, foldersMap map[string]string, tagsData map[string][]string) {
    // check for secret presence in HTTP header
    client_secret := r.Header.Get("secret")
    requestPath := strings.TrimSuffix(r.URL.Path,"/")

    // build data for template
    passTmplData := passPromptData{baseurl, strings.ReplaceAll(logo,"'","'\"'\"'"), client_secret, requestPath}

    // set default template 
    tmplSrc := passPromptView

    // if correct secret, print menu (index)
    if client_secret != secret {
        // render pass prompt template
        tmpl, err := template.New("passPrompt").Parse(tmplSrc)
        if err != nil { panic(err) }
        err = tmpl.Execute(w, passTmplData)
        if err != nil { panic(err) }
        return
    }

    // generate body of bash case with repo packages
    repoPackages := repoPackagesCasePrint(foldersMap, true, directory, baseurl)

    // find packages
    tags := strings.Split(strings.Replace(requestPath,"/t/","",-1), ",")
    packagesList := ""
    for _, key := range tags {
        if val, ok := tagsData[key]; ok {
            packages := strings.Join(val[:],"\n")+"\n"
            packagesList = packagesList + packages
        }
    }

    // build data for template
    tagsTmplData := TagsInstallData{client_secret, repoPackages, packagesList}

    // render template
    tmpl, err := template.New("tagsInstall").Parse(tmplTagsInstall)
    if err != nil { panic(err) }
    err = tmpl.Execute(w, tagsTmplData)
    if err != nil { panic(err) }
}

var tmplTagsInstall = `
SECRET="{{ .ClientSecret }}"
selectPackage() {
    case "$1" in
    {{ .RepoOpts }}
    esac
}

PACKAGES=$(cat <<-END
{{ .Packages }}END
)

echo -e "\n\e[97m-========================================================-"
echo -ne "\n  Follwing dotfiles will be installed in order:\n  "
COMMA=""
for PACKAGE in $PACKAGES; do
echo -en "$COMMA$PACKAGE" 
COMMA=", "
done

exec 3<>/dev/tty
echo -ne  "\n\n  Proceed? [Y/n]"
read -u 3 -n 1 -r -s
echo ""
if [[ $REPLY =~ ^[Nn]$ ]]
then
echo "  Aborted."
exit 0
fi

echo -e "\\n\e[97m-========================================================-\e[0m\n"
for NAME in $PACKAGES; do
selectPackage $NAME 
done
`

