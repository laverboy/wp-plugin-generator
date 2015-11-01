package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/laverboy/plugingenerator/Godeps/_workspace/src/github.com/gorilla/schema"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type options struct {
	PluginName      string `schema:"plugin-name"`
	PluginShortName string `schema:"short-name"`
	Description     string
	Version         string
	Database        bool
	Tests           bool
	Example         bool
}

func (o *options) PluginNameCamelCase() string {
	return strings.Replace(strings.Title(o.PluginName), " ", "", -1)
}

func (o *options) PluginLowercaseName() string {
	return strings.Replace(strings.ToLower(o.PluginName), " ", "", -1)
}

var (
	source = "./plugin.zip"
	tmp    = "./tmp/"
)

func main() {
	go getSource(source)

	assetsHandler := http.FileServer(http.Dir("./assets/"))
	http.Handle("/assets/", http.StripPrefix("/assets/", assetsHandler))

	http.HandleFunc("/", viewHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Println("listening on port", port)
	http.ListenAndServe(":"+port, nil)
}

func viewHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {

	// if we are simply viewing the page then just return template
	case "GET":
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, nil)

	// if form has been posted then process form options
	case "POST":
		req.ParseForm()

		opt := new(options)
		decoder := schema.NewDecoder()
		decoder.Decode(opt, req.PostForm)

		// fmt.Fprintf(w, "Options: %#v", opt)

		unzip(source, tmp)

		updateBaseFile(tmp, opt)
		findAndReplace(tmp, opt)

		os.Rename(tmp+"base-plugin-master/base-plugin.php", tmp+"base-plugin-master/"+opt.PluginShortName+".php")
		os.Rename(tmp+"base-plugin-master", tmp+opt.PluginShortName)

		defer os.RemoveAll(tmp + opt.PluginShortName)

		err := zipup(tmp+opt.PluginShortName, tmp+"PluginStarter.zip")
		if err != nil {
			fmt.Println("Error creating zip file", err)
		}
		defer os.RemoveAll(tmp + "PluginStarter.zip")

		w.Header().Set("Content-Disposition", "attachment; filename=Plugin.zip")
		w.Header().Set("Content-Type", "application/zip, application/octet-stream")
		// w.Header().Set("Content-Length", req.Header.Get("Content-Length"))

		zipfile, err := os.Open(tmp + "PluginStarter.zip")
		if err != nil {
			fmt.Println("Error opening newly created zip file", err)
		}
		defer zipfile.Close()

		io.Copy(w, zipfile)
	}

}

func zipup(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(source, strings.TrimPrefix(path, source))

		if info.IsDir() {
			header.Name += string(os.PathSeparator)
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	err = archive.Close()
	if err != nil {
		return err
	}

	return err
}

func findAndReplace(target string, opt *options) {
	filepath.Walk(target, func(path string, f os.FileInfo, err error) (e error) {
		if err != nil {
			fmt.Println("error occurred", err)
		}

		if !f.IsDir() {
			input, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Println("error reading file", err)
			}

			output := input
			output = bytes.Replace(output, []byte("BasePlugin"), []byte(opt.PluginNameCamelCase()), -1)
			output = bytes.Replace(output, []byte("baseplugin"), []byte(opt.PluginLowercaseName()), -1)
			output = bytes.Replace(output, []byte("Base Plugin"), []byte(opt.PluginName), -1)

			if err = ioutil.WriteFile(path, output, 0666); err != nil {
				fmt.Println("error writing file", err)
			}
		}

		return
	})
}

func updateBaseFile(target string, opt *options) {
	baseFile := filepath.Join(target, "base-plugin-master/base-plugin.php")

	input, err := ioutil.ReadFile(baseFile)
	if err != nil {
		fmt.Println("Error while reading base plugin file", err)
		return
	}

	var regexes []string

	if !opt.Example {
		regexes = append(regexes, `(?ms)^[\r\n]*?//\s?Example.*?};$[\r\n]`)
	}

	if !opt.Database {
		regexes = append(regexes, `(?ms)^[\r\n]*?//\s?DB.*?};$[\r\n]`)
	}

	output := input
	for _, res := range regexes {
		re := regexp.MustCompile(res)
		output = re.ReplaceAll(output, []byte(""))
	}

	output = bytes.Replace(output, []byte("0.1.0"), []byte(opt.Version), 2)
	output = bytes.Replace(output, []byte("The beginnings of yet another awesome plugin."), []byte(opt.Description), 1)

	if err = ioutil.WriteFile(baseFile, output, 0666); err != nil {
		fmt.Println("error writing base plugin file", err)
	}
}
