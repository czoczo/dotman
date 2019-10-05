// Tags handling
// =================================================

package main

import (
    "gopkg.in/yaml.v2"
//	"fmt"
    "strings"
    "log"
    "io/ioutil"
    "path/filepath"
)

type TagsData struct {
    Tags map[string][]string
}

var tagsData TagsData

func populateTagsMap(foldersMap map[string]string) {
    tagsFile := directory + "/config.yaml"

    if ! fileExists(tagsFile) {
        log.Println("config.yaml: file not found. skipping populating tags")
        return
    }

    filename, _ := filepath.Abs(tagsFile)
    yamlFile, err := ioutil.ReadFile(filename)

    if err != nil {
        log.Println("config.yaml: error reading file. skipping populating tags")
        return
    }

    err = yaml.Unmarshal(yamlFile, &tagsData)
        if err != nil {
        log.Println("config.yaml: error parsing file. skipping populating tags")
        return
    }

    // make empty value map with available packages as keys
    availablePackages := make(map[string]struct{})
    for _, val := range foldersMap {
        availablePackages[val] = struct{}{}
    }

    // show loaded tags
    log.Println("Loading tags from config.yaml file...")
    for tagkey, tagval := range tagsData.Tags {

        // print tag
        log.Println(tagkey + ": " + strings.Join(tagval,", "))

        // find and delete not found packages

        // make temporary list with validated packages
        tempList := make([]string,0)

        // for each element in tag
        for _, pack := range tagval {

            // check if corespondend package exists
            if _, ok := availablePackages[pack]; ok {
                tempList = append(tempList, pack)
            } else {
                log.Println("Warning! Package " + pack + " not found! Skipping.")
            }
        }

        // if no element in validates list delete tag, otherwise reassign validated packages
        if len(tempList) == 0 {
            delete(tagsData.Tags, tagkey)
        } else {
            tagsData.Tags[tagkey] = tempList
        }
    }
}
