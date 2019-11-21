// File managment helpers
// =================================================

package main

import (
    "os"
    "fmt"
    "log"
    "time"
    "regexp"
    "strings"
    "io/ioutil"
    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/knownhosts"
    "gopkg.in/src-d/go-git.v4/plumbing/transport"
    "gopkg.in/src-d/go-git.v4/storage/memory"
    "gopkg.in/src-d/go-git.v4/plumbing/object"
	git "gopkg.in/src-d/go-git.v4"
    gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
    githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// checking if file exists and is not a directory
func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

// in memory repo clone
func gitMemClone(auth transport.AuthMethod, url string) *object.Commit {

    // if prefix ssh:// remove from connection string
    if strings.HasPrefix(url,"ssh://") {
        url = strings.Replace(url,"ssh://","",1)
    }

    // clone the given repository to the given directory
    log.Println("checking repository " + url + " for updates")

    gitr, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
        Auth: auth,
        URL:      url,
    })

    CheckIfError(err)

    // ... retrieving the branch being pointed by HEAD
    ref, err := gitr.Head()
    CheckIfError(err)
    // ... retrieving the commit object
    commit, err := gitr.CommitObject(ref.Hash())
    CheckIfError(err)

    return commit
}

// loop for pull-type checking of updates on remote repo
func watchdog(auth transport.AuthMethod, url string, directory string, localHead string) {
    log.Println("Pull watchdog for changes on remote enabled.")
    go func() {
        var remote *object.Commit

        // run forever
        for {
            // sleep for amount of interval{
            time.Sleep(time.Duration(pullInterval) * time.Second)

            // get remote head and compare to local
            remote = gitMemClone(auth, url)
            if remote.Hash.String() != localHead {
                // if different update local
                gitSync(auth, url, directory)
            }
        }
    }()
}

// git pull method used to sync shared folder
func gitPull(directory string) *object.Commit {

    // we instance\iate a new repository targeting the given path (the .git folder)
    gitr, err := git.PlainOpen(directory)
    CheckIfError(err)

    // get the working directory for the repository
    gitw, err := gitr.Worktree()
    CheckIfError(err)

    // pull the latest changes from the origin remote and merge into the current branch
    log.Println("git pull origin")
    err = gitw.Pull(&git.PullOptions{
        Auth: auth,
        RemoteName: "origin",
    })

    // ... retrieving the branch being pointed by HEAD
    ref, err := gitr.Head()
    CheckIfError(err)
    // ... retrieving the commit object
    commit, err := gitr.CommitObject(ref.Hash())
    CheckIfError(err)

    fmt.Println(commit)
    return commit
}

// sync git repository by doing git clone
func gitSync(auth transport.AuthMethod, url string, directory string) *object.Commit {
    // clear data
    os.RemoveAll(directory)

    // if prefix ssh:// remove from connection string
    if strings.HasPrefix(url,"ssh://") {
        url = strings.Replace(url,"ssh://","",1)
    }

    // clone the given repository to the given directory
    log.Println("cloning repository " + url + " to folder " + directory)

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

    log.Println(commit)
    return commit
}

// return auth object based on url
func getAuth(url string, sshkey string, remoteHost string, sshAccept bool) transport.AuthMethod {
    // define ssh or http connection string
    var auth transport.AuthMethod

    // switch available protocols, and generate auth object
    if strings.HasPrefix(url, "http") {
        // create auth object for git  connection
        auth = &githttp.BasicAuth {
                Username: username,
                Password: password,
        }
    }
    if strings.HasPrefix(url, "ssh") {
        // read private key
        sshKey, _ := ioutil.ReadFile(sshkey)
        signer, _ := ssh.ParsePrivateKey([]byte(sshKey))

        // show remote key
        sshConfig := &ssh.ClientConfig{
            HostKeyCallback: KeyPrint,
        }
        _, err := ssh.Dial("tcp", remoteHost, sshConfig)

        // God, forgive me, for I have sinned. Badly. Dirty, very dirty... workaround of unexpected error - didn't have time to debug this.
        if err != nil && ! strings.Contains(err.Error(), "unable to authenticate, attempted methods [none], no supported methods remain") {
            log.Println(err)
        }

        // if accept  scan remote keys
        if sshAccept {
            log.Println("Adding remote host key...")
            // show remote key
            sshConfig := &ssh.ClientConfig{
                HostKeyCallback: TrustKey,
            }
            ssh.Dial("tcp", remoteHost, sshConfig)
        }

        // create known_hosts file
        hostKeyCallback, err := knownhosts.New(ssh_known_hosts)
        if err != nil {
            log.Fatal(err)
        }
        hostKeyCallbackHelper := gitssh.HostKeyCallbackHelper{HostKeyCallback: hostKeyCallback}

        // create auth object for git  connection
        auth = &gitssh.PublicKeys{User: "git", Signer: signer, HostKeyCallbackHelper: hostKeyCallbackHelper}
    }
    return auth
}

// return map with folders from the repository assigned to alphabet characters
func getFoldersMap(directory string, alphabet string) map[string]string {
    foldersMap := make(map[string]string)

    // get folder list made of 1st level of folders in reposiotory contining dotfiles
    files, err := ioutil.ReadDir(directory)
    if err != nil {
        log.Fatal(err)
    }

    // raise error if 
    if len(files) > len(alphabet) {
        log.Println("Congratz - you reached the limit of number of supported folders. Either:\n\na) Wait for author to have the same problem someday.\nb) Increase number of unique characters in 'alphabet' variable.\nc) Implement other solution yourself.\n\nDecide which is the fastes option on your own ;)\n\nExiting...")
        os.Exit(1)
    }

    // iterate over above list, and assign characters from alphabet to folders in map
    charCounter := 0
    for _, f := range files {

            // skip .git and README.md - case insensitive
            match, _ := regexp.MatchString("(?i)(.git|README.md|config.yaml)", f.Name())
            if match {
                continue
            }

            // print next alphabet character and mapped folder name
            foldersMap[string(alphabet[charCounter])] = f.Name()
            charCounter = charCounter+1

    }
    return foldersMap
}
