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

// serve static
// containsDotFile reports whether name contains a path element starting with a period.
// The name is assumed to be a delimited by forward slashes, as guaranteed
// by the http.FileSystem interface.
func containsDotFile(name string) bool {
    parts := strings.Split(name, "/")
    for _, part := range parts {
        if strings.HasPrefix(part, ".") && part!="README.md" {
            return true
        }
    }
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


func main() {
	//url := "ssh://git@cz0.cz:2222/czoczo/dotrepo.git"
	url := "https://cz0.cz/git/czoczo/dotfiles.git"
    username := "czoczo"
    password := ""
    directory := "dotfiles"
    baseurl := "127.0.0.1:1337"
    alphabet := "01234567890abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ"


    var auth transport.AuthMethod
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

    if _, err := os.Stat(directory); os.IsNotExist(err) {
        // Clone the given repository to the given directory
        Info("git clone %s %s", url, directory)

        r, err := git.PlainClone(directory, false, &git.CloneOptions{
            Auth: auth,
            URL:      url,
            Progress: os.Stdout,
        })


        //r, err := git.PlainClone(directory, false, &git.CloneOptions{
        //    Auth: auth,
        //    URL:      url,
        //    Progress: os.Stdout,
        //})

        CheckIfError(err)

        // ... retrieving the branch being pointed by HEAD
        ref, err := r.Head()
        CheckIfError(err)
        // ... retrieving the commit object
        commit, err := r.CommitObject(ref.Hash())
        CheckIfError(err)

        fmt.Println(commit)

    } else {
        // We instance\iate a new repository targeting the given path (the .git folder)
        r, err := git.PlainOpen(directory)
        CheckIfError(err)

        // Get the working directory for the repository
        w, err := r.Worktree()
        CheckIfError(err)

        // Pull the latest changes from the origin remote and merge into the current branch
        Info("git pull origin")
        err = w.Pull(&git.PullOptions{
            Auth: auth,
            RemoteName: "origin",
        })
    //    CheckIfError(err)

        // ... retrieving the branch being pointed by HEAD
        ref, err := r.Head()
        CheckIfError(err)
        // ... retrieving the commit object
        commit, err := r.CommitObject(ref.Hash())
        CheckIfError(err)

        fmt.Println(commit)
    }


    fs := http.FileServer(dotFileHidingFileSystem{http.Dir(directory)})
    http.Handle("/"+directory+"/", http.StripPrefix("/"+directory+"/", fs))

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		log.Println( r.URL.Path)
        //commaListRegex, _ := regexp.Compile("^/([0-9a-zA-Z]+,?)+$")
        listRegex, _ := regexp.Compile("^/[0-9a-zA-Z]+$")
        if r.URL.Path == "/" {


            files, err := ioutil.ReadDir(directory)
            if err != nil {
                log.Fatal(err)
            }

            if len(files) > len(alphabet) {
                    fmt.Fprint(w,"Congratz, you reached the limit of number of supported folders. Reach me or implement it yourself (or both ;) )")
            }

            charCounter := 0

            for _, f := range files {

                    match, _ := regexp.MatchString("(.git|README.md)", f.Name())
                    if match {
                        continue
                    }

                    fmt.Fprint(w,"\n"+string(alphabet[charCounter])+") "+f.Name()+" \n")
                    charCounter = charCounter+1

                    err = filepath.Walk(directory+"/"+f.Name(),
                        func(path string, info os.FileInfo, err error) error {
                        if err != nil {
                            return err
                        }

                        if info.IsDir() {
                            return nil
                        }
                        output := strings.TrimPrefix(path, directory+"/"+f.Name()+"/")
                        if filepath.Dir(output) != "." {
                            fmt.Fprint(w,"mkdir -p $HOME/" + filepath.Dir(output)+"\n")
                        }


                        fmt.Fprint(w,"curl " + "http://" + baseurl + "/" + path + " > $HOME/" + output+"\n")
                        return nil
                    })
                    if err != nil {
                        log.Println(err)
                    }
            }
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
        if listRegex.MatchString(r.URL.Path) {
            choice := strings.Replace(r.URL.Path,"/","",-1)
            var response string
            for _, char :=  range choice {
                response = response +" | "+ string(char)
            }
		    fmt.Fprintf(w, response)
            return
        }
        fmt.Fprintf(w, "404 - Not Found")
	})

	http.ListenAndServe(":1337", nil)
}
