package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type MkFile struct {
	FullFileName string
	FileName     string
}

func main() {

	const BookAuthor = "Kyle Simpson"
	const BookDir = "You-Dont-Know-JS"
	const BookRepoURL = "https://github.com/getify/You-Dont-Know-JS.git"
	const BookEpubCSSURL = "https://gist.githubusercontent.com/bmaupin/6e3649af73120fac2b6907169632be2c/raw/epub.css"
	const MetaTittlePrefix = "You Don't Know JS: "

	//-f --from
	const InputFormart = "gfm" //markdown_github (deprecated) | gfm  (GitHub-Flavored Markdown)
	//-t --to
	const OutputFormat = "epub" // kindle (mobi, awz...) not supported

	var books = map[string]string{
		"up & going":               MetaTittlePrefix + " 1 - Up & Going",
		"scope & closures":         MetaTittlePrefix + " 2 - Scope & Closures",
		"this & object prototypes": MetaTittlePrefix + " 3 - this & Object Prototypes",
		"types & grammar":          MetaTittlePrefix + " 4 - Types & Grammar",
		"async & performance":      MetaTittlePrefix + " 5 - Async & Performance",
		"es6 & beyond":             MetaTittlePrefix + " 6 - ES6 & Beyond",
	}

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(workDir)

	if _, err := os.Stat(BookDir); os.IsNotExist(err) {
		execExternal(workDir, "git", "clone", "--progress", "-v", BookRepoURL)
	} else {
		execExternal(workDir, "git", "-C", BookDir, "reset", "--hard", "HEAD")
		execExternal(workDir, "git", "-C", BookDir, "pull", "--progress", "-v")
	}
	execExternal(workDir, "wget", "-O", BookDir+"/epub.css", BookEpubCSSURL)

	for folder, bookName := range books {
		var bookFolder = BookDir + "/" + folder

		log.Println(bookName)
		var mkFiles = checkExt(bookFolder, ".md")
		var sigleMkNames []string
		for _, mkFile := range mkFiles {
			replaceFileInline(mkFile.FullFileName)
			sigleMkNames = append(sigleMkNames, mkFile.FileName)
		}

		epubOptions := []string{
			"-f", "markdown+smart", "-o", fmt.Sprintf("../%s.epub", bookName),
			"--epub-cover-image=cover.jpg", "--css=../epub.css",
			// "--no-highlight",
			"-M", fmt.Sprintf("author=\"%s\"", BookAuthor),
			"-M", fmt.Sprintf("title=\"%s\"", bookName),
			"-M", "lang=en-US",
			"--verbose", "--log=log.json", "../preface.md"}
		epubOptions = append(epubOptions, sigleMkNames...)

		fmt.Println(epubOptions)
		buildEpub(bookFolder, epubOptions...)
		execExternal(BookDir, "ebook-convert", fmt.Sprintf("%s.epub", bookName), fmt.Sprintf("%s.mobi", bookName))
	}

	execExternal(BookDir, "ls")
}

func replaceFileInline(fileName string) {
	const DeleteYouDontKnow = "# You Don't Know JS.*"
	const BrFind = "<br>"
	const BrRepleace = "<br/>"
	var ReImg = regexp.MustCompile(`(?m)(<img.*[^\/])>`)
	const SubsImg = `$1/>`
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		r, err := regexp.MatchString(DeleteYouDontKnow, line)
		if err == nil && r {
			fmt.Printf("Delete ['%s'] in file >> %s\n", line, fileName)
			line = ""
		}
		line = strings.Replace(line, BrFind, BrRepleace, -1)

		if ReImg.MatchString(line) {
			var imgLine = ReImg.ReplaceAllString(line, SubsImg)
			fmt.Printf("Replace IMG ['%s'] with >> %s\n", line, imgLine)
			line = imgLine
		}

		lines[i] = line
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func checkExt(pathS string, ext string) []MkFile {
	var files []MkFile
	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString(ext, f.Name())
			if err == nil && r {
				mkFile := MkFile{path, f.Name()}
				files = append(files, mkFile)
			}
		}
		return nil
	})
	return files
}

func execExternal(folder string, command string, options ...string) []byte {
	cmd := exec.Command(command, options...)
	cmd.Dir = folder
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("\n\t cmd.Run() failed with %s\n%s\n ==> %s -- %s", err, out, command, options)
	}
	fmt.Printf("\n combined out:\n%s\n", string(out))
	return out
}

func buildEpub(folder string, options ...string) {
	cmd := exec.Command("pandoc", options...)
	cmd.Dir = folder
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("\n\t cmd.Run(pandoc) failed with %s\n%s", err, out)
	}
	fmt.Printf("\n combined out:\n%s\n", string(out))
}
