package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
)

// writeFile writes content to a file
func writeFile(filePath string, data []byte) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("no such file: %s make sure running from the root of the repo directory", filePath)
	}

	fmt.Printf("Writing new %s\n", filePath)
	err := ioutil.WriteFile(filePath, data, os.FileMode(0664))
	if err != nil {
		log.Fatalf("Error writing '%s': %s", filePath, err)
	}
}

type fileInfo struct {
	path        string
	varName     string
	srcPath     string
	description string
}

func buildMetaDataFile() {
	var files []fileInfo
	files = append(files,
		fileInfo{
			path:        "disposable.txt",
			varName:     "disposableDomains",
			srcPath:     "../../metadata_disposable.go",
			description: "// map to store disposable domains data",
		},
		fileInfo{
			path:        "free.txt",
			varName:     "freeDomains",
			srcPath:     "../../metadata_free.go",
			description: "// map to store free domains data",
		},
		fileInfo{
			path:        "role.txt",
			varName:     "roleAccounts",
			srcPath:     "../../metadata_role.go",
			description: "// map to store role-based accounts data",
		},
	)

	for _, f := range files {
		log.Printf("Building map for: %s\n", f.path)
		file, err := os.Open(f.path)
		if err != nil {
			panic(fmt.Sprintf("open meta data f %s fail: %v ", f, err))
		}

		output := bytes.Buffer{}
		output.WriteString("package emailverifier\n\n")
		output.WriteString(f.description + "\n")
		output.WriteString(fmt.Sprintf("var %s = map[string]bool {\n", f.varName))

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		data := make(map[string]bool)
		for scanner.Scan() {
			key := scanner.Text()

			if !data[key] {
				output.WriteString("\t")
				output.WriteString(strconv.Quote(key))
				output.WriteString(": ")
				output.WriteString("true")
				output.WriteString(",\n")

			}
			data[key] = true
		}
		output.WriteString("}")
		log.Printf("Read %d mappings in %s\n", len(data), f.path)

		err = file.Close()
		if err != nil {
			panic(fmt.Sprintf("close role meta data file %s fail: %v ", f.path, err))
		}
		writeFile(f.srcPath, output.Bytes())

	}

}

func updateMetaData() {
	cmd := exec.Command(
		"/bin/bash",
		"-c",
		"./update.sh",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("error calling update.sh to update meta data: %s", err.Error())
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("error calling update.sh to update meta data: %s", err.Error())
	}
	if err = cmd.Start(); err != nil {
		log.Fatalf("error executing update.sh to update meta data: %s", err.Error())
	}
	data, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Fatalf("error reading update.sh result: %s : %s", err.Error(), data)
	}
	outputBuf := bufio.NewReader(stdout)

	for {
		output, _, err := outputBuf.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		log.Println(string(output))
	}

	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	updateMetaData()
	buildMetaDataFile()
}
