package main

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
    gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	git "gopkg.in/src-d/go-git.v4"
	. "gopkg.in/src-d/go-git.v4/_examples"
)

func main() {
	url := "ssh://git@cz0.cz:2222/czoczo/dotrepo.git"
    directory := "dotfiles"
    baseurl := "127.0.0.1:1337"
    alphabet := "01234567890abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ"

    s := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
    sshKey, _ := ioutil.ReadFile(s)
    signer, _ := ssh.ParsePrivateKey([]byte(sshKey))
    auth := &gitssh.PublicKeys{User: "git", Signer: signer}

    if _, err := os.Stat(directory); os.IsNotExist(err) {
        // Clone the given repository to the given directory
        Info("git clone %s %s", url, directory)

        //	r, err := git.PlainClone(directory, false, &git.CloneOptions{
        //		Auth: &http.BasicAuth{
        //			Username: username,
        //			Password: password,
        //		},
        //		URL:      url,
        //		Progress: os.Stdout,
        //	})


        r, err := git.PlainClone(directory, false, &git.CloneOptions{
            Auth: auth,
            URL:      url,
            Progress: os.Stdout,
        })

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

    // serve static
    fs := http.FileServer(http.Dir(directory))
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
