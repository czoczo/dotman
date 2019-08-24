// Utilities
// =================================================

package main

import (
	"os"
	"fmt"
    "log"
    "strings"
    "io/ioutil"
    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/knownhosts"
    "gopkg.in/src-d/go-git.v4/plumbing/transport"
	git "gopkg.in/src-d/go-git.v4"
    gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
    githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[35;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// checking if file exists and is not a directory
func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

// git pull method used to sync shared folder
func gitPull(directory string) {

    // we instance\iate a new repository targeting the given path (the .git folder)
    gitr, err := git.PlainOpen(directory)
    CheckIfError(err)

    // get the working directory for the repository
    gitw, err := gitr.Worktree()
    CheckIfError(err)

    // pull the latest changes from the origin remote and merge into the current branch
    Info("git pull origin")
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
}

// git sync repository by either doing git clone, or pulling if existant
func gitSync(auth transport.AuthMethod, url string, directory string) {
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
        if err != nil && err != fmt.Errorf("ssh: handshake failed: ssh: unable to authenticate, attempted methods [none], no supported methods remain") {
            log.Println(err)
        }

        log.Println("Gonna remote host key...")
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

