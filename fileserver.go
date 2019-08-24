// Functions for serving cloned repository files on HTTP
// =================================================

package main

import (
	"net/http"
	"fmt"
	"os"
    "strings"
)

// blacklisting some files from being served

// check if path starts with dot or a README.md
func containsDotFile(name string) bool {
        if strings.HasPrefix(name, "/.") || name=="/README.md" {
            return true
        }
    return false
}

// fileHidingFile is the http.File use in fileHidingFileSystem.
// it is used to wrap the Readdir method of http.File so that we can
// remove files and directories that are blacklisted from its output.
type fileHidingFile struct {
    http.File
}

// readdir is a wrapper around the Readdir method of the embedded File
// that filters out all files that start with a period in their name.
func (f fileHidingFile) Readdir(n int) (fis []os.FileInfo, err error) {
    files, err := f.File.Readdir(n)
    for _, file := range files { // filters out the files
        if !strings.HasPrefix(file.Name(), ".") && file.Name()!="README.md" {
            fis = append(fis, file)
        }
    }
    return
}

// fileHidingFileSystem is an http.FileSystem that hides
// hidden "dot files" from being served.
type fileHidingFileSystem struct {
    http.FileSystem
}

// open is a wrapper around the Open method of the embedded FileSystem
// that serves a 403 permission error when name has a file or directory
// with whose name starts with a period in its path.
func (fs fileHidingFileSystem) Open(name string) (http.File, error) {
    if containsDotFile(name) { // If dot file, return 403 response
        return nil, os.ErrPermission
    }

    file, err := fs.FileSystem.Open(name)
    if err != nil {
        return nil, err
    }
    return fileHidingFile{file}, err
}

// Check if secret passed to protect shared dotfiles
func checkSecretThenServe(h http.Handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        client_secret := r.Header.Get("secret")
        if client_secret != secret {
            fmt.Fprint(w,"Ups! No secret given.\n")
            return
        }
        // Serve with the actual handler.
        h.ServeHTTP(w, r)
    }
}


