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

	fmt.Println("Dowloading from " + dlURL)

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

	io.Copy(out, resp.Body)

	unpackZip(tempDir)
}

func unpackZip(tempDir string) {
	zipReader, err := zip.OpenReader(tempDir + ".zip")

	if err != nil {
		panic(err)
	}

	defer zipReader.Close()

	for _, file := range zipReader.File {
		filePath := filepath.Join(tempDir, file.Name)
		fileParts := strings.Split(filePath, string(os.PathSeparator))

		var rootIndex = 0

		for i, part := range fileParts {
			if strings.HasPrefix(part, "InventivetalentDev-minecraft-assets") {
				filePath = strings.ReplaceAll(filePath, part, "")
				rootIndex = i
				break
			}
		}

		if len(fileParts) >= 9 && fileParts[rootIndex+1] == "data" {
			os.Remove(tempDir + ".zip")
			generateObjects(tempDir)
			break
		}

		if len(fileParts) >= 11 && (fileParts[rootIndex+3] == "atlases" ||
			fileParts[rootIndex+3] == "equipment" ||
			fileParts[rootIndex+3] == "font" ||
			fileParts[rootIndex+3] == "items" ||
			fileParts[rootIndex+3] == "particles" ||
			fileParts[rootIndex+3] == "post_effect" ||
			fileParts[rootIndex+3] == "resourcepacks" ||
			fileParts[rootIndex+3] == "shaders" ||
			fileParts[rootIndex+3] == "sounds" ||
			fileParts[rootIndex+3] == "texts") {
			continue
		}

		if fileParts[len(fileParts)-1] == "_list.json" ||
			fileParts[len(fileParts)-1] == "sounds.json" ||
			fileParts[len(fileParts)-1] == "_all.json" ||
			fileParts[len(fileParts)-1] == "regional_compliancies.json" ||
			fileParts[len(fileParts)-1] == "gpu_warnlist.json" {
			continue
		}

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

		fmt.Println(file.Name)

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

type IIndex struct {
	Hash     string   `json:"hash"`
	Path     string   `json:"path"`
	Length   int64    `json:"length"`
	Versions []string `json:"versions"`
}

func generateObjects(tempDir string) {
	//	var indexFile []IIndex

	filepath.Walk(filepath.Join(tempDir, "assets", "minecraft"), func(path string, info os.FileInfo, err error) error {
		fmt.Println(path, info.Size())
		return nil
	})
}
