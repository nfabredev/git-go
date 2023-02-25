package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/master\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		blob_sha := os.Args[3]

		if blob_sha == "" {
			fmt.Println("Usage : ./your_git.sh cat-file -p <blob_sha>")
			os.Exit(1)
		}

		dir, filename := blob_sha[0:2], blob_sha[2:]

		file, err := os.Open(".git/objects/" + dir + "/" + filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		reader, err := zlib.NewReader(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer reader.Close()

		buf := new(strings.Builder)
		_, err = io.Copy(buf, reader)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fileContent := buf.String()
		fileContentArray := strings.Split(fileContent, "\x00")

		fmt.Print(fileContentArray[1])

	case "hash-object":
		fileToSave := os.Args[3]
		file, err := os.ReadFile(fileToSave)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		header := "blob" + " " + fmt.Sprint(len([]rune(string(file)))) + "\x00"
		content := header + string(file)

		h := sha1.New()
		h.Write([]byte(content))
		fileSHA := hex.EncodeToString(h.Sum(nil))

		var b bytes.Buffer
		writer := zlib.NewWriter(&b)
		writer.Write([]byte(content))
		writer.Close()

		dir, filename := fileSHA[0:2], fileSHA[2:]
		if err := os.Mkdir(".git/objects/"+dir, os.ModePerm); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = os.WriteFile((".git/objects/" + dir + "/" + filename), b.Bytes(), 0666)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Print(fileSHA)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
