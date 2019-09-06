// Tags handling
// =================================================

package main

import (
    "gopkg.in/yaml.v2"
	"fmt"
    "log"
    "io/ioutil"
    "path/filepath"
)

type TagsData struct {
    Tags map[string][]string
}

func populateTagsMap() {
    tagsFile := directory + "/tags.yaml"

    if ! fileExists(tagsFile) {
        log.Println("tags.yaml: file not found. skipping populating tags")
        return
    }

    filename, _ := filepath.Abs(tagsFile)
    yamlFile, err := ioutil.ReadFile(filename)

    if err != nil {
        log.Println("tags.yaml: error reading file. skipping populating tags")
        return
    }

    var tagsData TagsData

    err = yaml.Unmarshal(yamlFile, &tagsData)
        if err != nil {
        log.Println("tags.yaml: error parsing file. skipping populating tags")
        return
    }

    //DEBUG fmt.Printf("Value: %#v\n", tagsData)

    // show loaded tags
    for tagkey, tagval := range tagsData.Tags {
        log.Println(tagkey + ",")
        log.Println(tagval)
    }
}
