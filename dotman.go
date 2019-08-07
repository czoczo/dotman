package main

// go get -u gopkg.in/src-d/go-git.v4/

import (
	"fmt"
	"net/http"
    "log"
    "regexp"
	"os"
    "strings"
    "io/ioutil"
    "path/filepath"
    "golang.org/x/crypto/ssh"
    "gopkg.in/src-d/go-git.v4/plumbing/transport"
    gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	git "gopkg.in/src-d/go-git.v4"
    githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	. "gopkg.in/src-d/go-git.v4/_examples"
)

// define ssh or http connection string
var url string
var username string
var password string
var directory string
var secret string

// server config
var baseurl string
var alphabet string

// authorization object for git connection
var auth transport.AuthMethod

// serve static
// containsDotFile reports whether name contains a path element starting with a period.
// The name is assumed to be a delimited by forward slashes, as guaranteed
// by the http.FileSystem interface.
func containsDotFile(name string) bool {
//    parts := strings.Split(name, "/")
//    for _, part := range parts {
        if strings.HasPrefix(name, "/.") || name=="/README.md" {
            return true
        }
//    }
    return false
}

// dotFileHidingFile is the http.File use in dotFileHidingFileSystem.
// It is used to wrap the Readdir method of http.File so that we can
// remove files and directories that start with a period from its output.
type dotFileHidingFile struct {
    http.File
}

// Readdir is a wrapper around the Readdir method of the embedded File
// that filters out all files that start with a period in their name.
func (f dotFileHidingFile) Readdir(n int) (fis []os.FileInfo, err error) {
    files, err := f.File.Readdir(n)
    for _, file := range files { // Filters out the dot files
        if !strings.HasPrefix(file.Name(), ".") && file.Name()!="README.md" {
            fis = append(fis, file)
        }
    }
    return
}

// dotFileHidingFileSystem is an http.FileSystem that hides
// hidden "dot files" from being served.
type dotFileHidingFileSystem struct {
    http.FileSystem
}

// Open is a wrapper around the Open method of the embedded FileSystem
// that serves a 403 permission error when name has a file or directory
// with whose name starts with a period in its path.
func (fs dotFileHidingFileSystem) Open(name string) (http.File, error) {
    if containsDotFile(name) { // If dot file, return 403 response
        return nil, os.ErrPermission
    }

    file, err := fs.FileSystem.Open(name)
    if err != nil {
        return nil, err
    }
    return dotFileHidingFile{file}, err
}



func availableOptsCasePrint(w http.ResponseWriter, foldersMap map[string]string, byName bool)  {
    // itertate over available option
    for key, val := range foldersMap {
            index := key
            if byName { index = val }
            fmt.Fprint(w,"        "+index+")\n            if [ $2 ]; then\n                printf \"\\e[0;32m*\\e[0m)\\e[0;35m %s\\e[0m\" \""+val+"\"\n                return\n            fi\n")
            fmt.Fprint(w,"             echo -e \"Installing \\e[0;35m"+val+"\\e[0m\"\n")
            // search folders and add mkdir and download commands
            // recursive walk thorough the dir
            err := filepath.Walk(directory+"/"+val,
                func(path string, info os.FileInfo, err error) error {
                if err != nil {
                    return err
                }

                // skip of not a dir
                if info.IsDir() {
                    return nil
                }

                // print mkdir commands
                output := strings.TrimPrefix(path, directory+"/"+val+"/")
                if filepath.Dir(output) != "." {
                    fmt.Fprint(w,"             #mkdir -p $HOME/" + filepath.Dir(output)+"\n")
                }

                // print download commands
                fmt.Fprint(w,"             #curl -H\"secret:" + secret + "\" \"" + baseurl + "/" + path + "\" > $HOME/" + output+"\n")
                // add option to update list
                fmt.Fprint(w,"             cat \"$HOME/.dotman/managed\" | grep -q \""+val+"\" || echo \""+val+"\" >> \"$HOME/.dotman/managed\" \n")
                return nil
            })
            if err != nil {
                log.Println(err)
            }
            fmt.Fprint(w,"             ;;\n")
    }
}

func gitPull(directory string) {
    // We instance\iate a new repository targeting the given path (the .git folder)
    gitr, err := git.PlainOpen(directory)
    CheckIfError(err)

    // Get the working directory for the repository
    gitw, err := gitr.Worktree()
    CheckIfError(err)

    // Pull the latest changes from the origin remote and merge into the current branch
    Info("git pull origin")
    err = gitw.Pull(&git.PullOptions{
        Auth: auth,
        RemoteName: "origin",
    })
//    CheckIfError(err)

    // ... retrieving the branch being pointed by HEAD
    ref, err := gitr.Head()
    CheckIfError(err)
    // ... retrieving the commit object
    commit, err := gitr.CommitObject(ref.Hash())
    CheckIfError(err)

    fmt.Println(commit)
}

func main() {
    // define ssh or http connection string
    url = "ssh://git@cz0.cz:2222/czoczo/dotrepo.git"
    //url := "https://cz0.cz/git/czoczo/dotfiles.git"
    username = "czoczo"
    password = ""
    directory = "dotfiles"
    secret = "dupa.8"

    // server config
    baseurl = "http://127.0.0.1:1337"
    alphabet = "01234567890abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ"

    // available options list
    foldersMap := make(map[string]string)

    logo := `
\e[2;35m                $$.  ,$$    $$    $$,   $$
                $$$,.$$$  ,$/\$.  $$\.  $$
\e[0;35m          .$$$. \e[2;35m$$'$$'$$ ,$$  $$. $$'$$.$$
\e[0;35m          $$$$$ \e[2;35m$$    $$ $$&&&&$$ $$  '\$$
\e[0;35m          '$$$' \e[2;35m$$    $$ $$    $$ $$   '$$\e[0m
`

    // switch protocols
    if strings.HasPrefix(url, "http") {
        auth = &githttp.BasicAuth {
                Username: username,
                Password: password,
        }
    }
    if strings.HasPrefix(url, "ssh") {
        s := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
        sshKey, _ := ioutil.ReadFile(s)
        signer, _ := ssh.ParsePrivateKey([]byte(sshKey))
        auth = &gitssh.PublicKeys{User: "git", Signer: signer}
    }

    // clone repo locally if not already here
    if _, err := os.Stat(directory); os.IsNotExist(err) {
        // Clone the given repository to the given directory
        Info("git clone %s %s", url, directory)

        gitr, err := git.PlainClone(directory, false, &git.CloneOptions{
            Auth: auth,
            URL:      url,
            Progress: os.Stdout,
        })

        CheckIfError(err)

        // ... retrieving the branch being pointed by HEAD
        ref, err := gitr.Head()
        CheckIfError(err)
        // ... retrieving the commit object
        commit, err := gitr.CommitObject(ref.Hash())
        CheckIfError(err)

        fmt.Println(commit)

    } else {
        gitPull(directory)
    }

    // 
    checkHeaderThenServe := func(h http.Handler) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            client_secret := r.Header.Get("secret")
            if client_secret != secret {
                fmt.Fprint(w,"Ups! No secret given.\n")
                w.WriteHeader(http.StatusForbidden)
                return
            }
            // Serve with the actual handler.
            h.ServeHTTP(w, r)
        }
    }

    // serve locally cloned repo with dotfiles through file server under 'directory' variable name
    fs := checkHeaderThenServe(http.FileServer(dotFileHidingFileSystem{http.Dir(directory)}))
    http.Handle("/"+directory+"/", http.StripPrefix("/"+directory+"/", fs))

    // handle other http requests
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		log.Println( r.URL.Path)
        //commaListRegex, _ := regexp.Compile("^/([0-9a-zA-Z]+,?)+$")
        //listRegex, _ := regexp.Compile("^/[0-9a-zA-Z]+$")


        // list folders in reposiotory with dotfiles
        files, err := ioutil.ReadDir(directory)
        if err != nil {
            log.Fatal(err)
        }

        charCounter := 0
        for _, f := range files {
                // skip .git and README.mdi - case insensitive
                match, _ := regexp.MatchString("(?i)(.git|README.md)", f.Name())
                if match {
                    continue
                }

                // print next alphabet character and mapped folder name
                foldersMap[string(alphabet[charCounter])] = f.Name()
                charCounter = charCounter+1

        }
        // hanlde main request, print main script
        if r.URL.Path == "/" {
            // start with shebang
            fmt.Fprint(w,"#!/bin/sh \n")

            fmt.Fprint(w,"tput clear\n")
            fmt.Fprint(w,"echo -e '"+strings.ReplaceAll(logo,"'","'\"'\"'")+"'\n")
            fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0;37m\"\n")
            client_secret := r.Header.Get("secret")
            if client_secret != secret {
                fmt.Fprint(w,`
exec 3<>/dev/tty
printf "secret: "
read -u 3 -s SECRET
curl -H"secret:$SECRET" ` + baseurl + " | sh -")
                return
            }
            // print error if more folders than available alphabet
            if len(files) > len(alphabet) {
                fmt.Fprint(w,"printf \"%2s%s\\n%4s%s\\n%4s%s\\n%4s%s\\n\\n%2s%s\\n\\n\" \"\" \"Congratz - you reached the limit of number of supported folders. Either:\" \"\" \"a) Wait for me to have the same problem someday.\" \"\" \"b) Increase number of unique characters in 'alphabet' variable.\" \"\" \"c) Implement other solution yourself.\" \"\" \"Decide which is the fastes option on your own ;)\"\n")
            }
            fmt.Fprint(w,"printf \"%2s%s\\n%2s%s\\n\\n\" \"\" \"Choose dotfiles to be installed.\" \"\" \"Select by typing keys (green) and confirm with enter.\"\n")
            fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0m\"\n")

            // vars for iterating available options
            count := 0
            nl := ""
            // iterate over options
            for char := range alphabet {
                    // exit if no keys in map anymore
                    key := string(char)
                    if _, ok := foldersMap[key]; !ok {
                        continue
                    }
                    count += 1
                    if count % 3 == 0 {
                        nl = "\\n"
                    }
                    fmt.Fprint(w,"printf \"  \\e[32m%s\\e[0m)\\e[35m %-15s\\e[0m"+nl+"\" \""+key+"\" \""+foldersMap[key]+"\" \n")
                    nl = ""
            }

            // touch list of dotfiles to be update
            fmt.Fprint(w,"mkdir -p \"$HOME/.dotman\"; touch \"$HOME/.dotman/managed\"\n")
            // print case function
            fmt.Fprint(w,"selectOption() {\n    case \"$1\" in\n")
            availableOptsCasePrint(w, foldersMap, false)
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

            return
        }
        client_secret := r.Header.Get("secret")
        if client_secret != secret {
		    fmt.Fprintf(w, "echo \"Secret not given.\"")
            return
        }
//        if commaListRegex.MatchString(r.URL.Path) {
//            slice := strings.Split(strings.Replace(r.URL.Path,"/","",-1), ",")
//            var response string
//            for _, num := range slice {
//                response = response +" | "+ num
//            }
//		    fmt.Fprintf(w, response)
//            return
//        }

//        // if URI with chosen options print download script
//        if listRegex.MatchString(r.URL.Path) {
//            choice := strings.Replace(r.URL.Path,"/","",-1)
//            var response string
//            for _, char :=  range choice {
//                response = response +" | "+ string(char)
//            }
//		    fmt.Fprintf(w, response)
//            return
//        }

        // pull git repo
        if r.URL.Path == "/sync" {
            gitPull(directory)
            fmt.Fprintf(w, "echo \"Repo synced\"")
            return
        }

        // if none aboue catched, return 404
        if r.URL.Path == "/update" {
            // print case function
            fmt.Fprint(w,"tput clear\n")
            fmt.Fprint(w,"selectOption() {\n    case \"$1\" in\n")
            availableOptsCasePrint(w, foldersMap, true)
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
            return
        }

        fmt.Fprintf(w, "echo \"404 - Not Found\"")
	})

	http.ListenAndServe(":1337", nil)
}
