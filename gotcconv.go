package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func EscapeText(text string) string {
	buf := new(bytes.Buffer)
	xml.EscapeText(buf, []byte(text))
	return buf.String()
}

func DataRace(logPath string) {
	writer := bufio.NewWriter(os.Stdout)

	buf := new(bytes.Buffer)
	err := filepath.Walk(filepath.Dir(logPath), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasPrefix(path, logPath) {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			buf.Write(bytes)
			buf.WriteString("\n")
			buf.WriteString("\n")
		}
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	writer.WriteString(`<!DOCTYPE html><html><head lang="en"><meta charset="UTF-8"><title>Data race report</title></head><body><pre>`)
	writer.WriteString("\n")
	writer.WriteString(EscapeText(buf.String()))
	writer.WriteString("\n")
	writer.WriteString(`</pre></body></html>`)

	if err := writer.Flush(); err != nil {
		log.Fatal(err)
	}
}

func ErrCheck(inFilePath string) {
	// go get github.com/kisielk/errcheck
	// errcheck ./...

	inFile, err := os.Open(inFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	writer := bufio.NewWriter(os.Stdout)

	writer.WriteString(`<checkstyle version="4.3">`)
	writer.WriteString("\n")
	scanner := bufio.NewScanner(inFile)
	var prevFileName string
	for scanner.Scan() {
		origLine := scanner.Text()
		line := origLine
		index := strings.Index(line, "\t")
		if index == -1 {
			log.Fatalf(`Could not parse line "%s"`, origLine)
		}
		message := line[index+1:]
		line = line[:index]
		index1 := strings.Index(line, ":")
		index2 := strings.LastIndex(line, ":")
		if index1 == -1 || index2 == -1 {
			log.Fatalf(`Could not parse line "%s"`, origLine)
		}
		fileName := line[:index1]
		lineNumber := line[index1+1 : index2]
		columnNumber := line[index2+1:]

		if prevFileName != fileName {
			if prevFileName != "" {
				writer.WriteString("\t</file>\n")
			}
			writer.WriteString("\t")
			writer.WriteString(fmt.Sprintf(`<file name="%s">`, EscapeText(fileName)))
			writer.WriteString("\n")
		}
		writer.WriteString("\t\t")
		writer.WriteString(fmt.Sprintf(`<error line="%s" column="%s" severity="warning" message="%s"></error>`, lineNumber, columnNumber, EscapeText(message)))
		writer.WriteString("\n")

		prevFileName = fileName
	}
	if prevFileName != "" {
		writer.WriteString("\t</file>\n")
	}
	writer.WriteString("</checkstyle>\n")

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if err := writer.Flush(); err != nil {
		log.Fatal(err)
	}
}

func Vet(inFilePath string) {
	inFile, err := os.Open(inFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	writer := bufio.NewWriter(os.Stdout)

	writer.WriteString(`<checkstyle version="4.3">`)
	writer.WriteString("\n")
	scanner := bufio.NewScanner(inFile)
	var prevFileName string
	for scanner.Scan() {
		origLine := scanner.Text()
		line := origLine
        if strings.Index(line, "exit status") != -1 { continue }
		index := strings.Index(line, ": ")
		if index == -1 {
			log.Fatalf(`Could not parse line "%s"`, origLine)
		}
		message := line[index+2:]
		line = line[:index]
		index = strings.LastIndex(line, ":")
		if index == -1 {
			log.Fatalf(`Could not parse line "%s"`, origLine)
		}
		fileName := line[:index]
		lineNumber := line[index+1:]

		if prevFileName != fileName {
			if prevFileName != "" {
				writer.WriteString("\t</file>\n")
			}
			writer.WriteString("\t")
			writer.WriteString(fmt.Sprintf(`<file name="%s">`, EscapeText(fileName)))
			writer.WriteString("\n")
		}
		writer.WriteString("\t\t")
		writer.WriteString(fmt.Sprintf(`<error line="%s" column="1" severity="warning" message="%s"></error>`, lineNumber, EscapeText(message)))
		writer.WriteString("\n")

		prevFileName = fileName
	}
	if prevFileName != "" {
		writer.WriteString("\t</file>\n")
	}
	writer.WriteString("</checkstyle>\n")

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if err := writer.Flush(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// /Users/akorzhevskii/Downloads/reports/report
	logPath := flag.String("log_path", "", "go -race log path")
	// /Users/akorzhevskii/Downloads/errcheck.out
	errcheckReport := flag.String("errcheck", "", "errcheck report")
	// /Users/akorzhevskii/Downloads/govet.out
	vetReport := flag.String("vet", "", "go vet report")
	flag.Parse()

	if *logPath != "" {
		DataRace(*logPath)
		return
	}

	if *errcheckReport != "" {
		ErrCheck(*errcheckReport)
		return
	}

	if *vetReport != "" {
		Vet(*vetReport)
		return
	}
}
