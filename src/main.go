package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var versionFlag = flag.String("version", "1.21.4", "Minecraft Version")

	flag.Parse()

	var tempDir = os.TempDir() + "/CEMCAU_" + *versionFlag
	var dlURL = "https://github.com/InventivetalentDev/minecraft-assets/zipball/refs/heads/" + *versionFlag

	out, err := os.Create(tempDir + ".zip")

	if err != nil {
		panic(err)
	}

	defer out.Close()

	resp, err := http.Get(dlURL)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println(n)

	zipReader, err := zip.OpenReader(tempDir + ".zip")

	if err != nil {
		panic(err)
	}

	defer zipReader.Close()

	for _, file := range zipReader.File {
		filePath := filepath.Join(tempDir, file.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(tempDir)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := file.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}
