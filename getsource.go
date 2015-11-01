package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func getSource(source string) {

	url := "https://github.com/laverboy/base-plugin/archive/master.zip"

	// Create output file
	output, err := os.Create(source)
	if err != nil {
		fmt.Println("Error while creating", source, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}

}
