package main

import (
	"archive/zip"
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/goccy/go-json"
	"github.com/joho/godotenv"
)

var versionFlag = flag.String("version", "1.21.4", "Minecraft Version")

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	flag.Parse()

	var tempDir = filepath.Join(os.TempDir(), "/CEMCAU_"+*versionFlag)
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
	var index []IIndex
	var objectPath = tempDir + "_objects"

	filepath.Walk(filepath.Join(tempDir, "assets", "minecraft"), func(path string, info os.FileInfo, err error) error {
		var fPath = strings.ReplaceAll(strings.ReplaceAll(path, tempDir, ""), "\\", "/")

		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)

		if err != nil {
			panic(err)
		}

		h := md5.New()
		if _, err := io.Copy(h, file); err != nil {
			panic(err)
		}

		hash := fmt.Sprintf("%x", h.Sum(nil))

		var basePath = filepath.Join(objectPath, hash[0:2])

		os.MkdirAll(basePath, os.ModePerm)

		fileContent, _ := os.ReadFile(path)

		os.WriteFile(filepath.Join(basePath, hash), fileContent, 0644)

		index = append(index, IIndex{
			Hash:     hash,
			Path:     fPath,
			Length:   info.Size(),
			Versions: []string{*versionFlag},
		})
		return nil
	})

	jsonString, _ := json.Marshal(index)

	indexFile, err := os.Create(filepath.Join(objectPath, "index.json"))

	if err != nil {
		panic(err)
	}
	defer indexFile.Close()

	io.WriteString(indexFile, string(jsonString[:]))

	os.Remove(tempDir)
	uploadToS3(objectPath)
}

func uploadToS3(objectPath string) {
	pathStyle, _ := strconv.ParseBool(os.Getenv("S3_PATH_STYLE"))

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(os.Getenv("S3_ENDPOINT")),
		Region:       os.Getenv("S3_REGION"),
		Credentials:  credentials.NewStaticCredentialsProvider(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), ""),
	}, func(o *s3.Options) {
		o.UsePathStyle = pathStyle
	})
	uploader := manager.NewUploader(client)

	filepath.Walk(objectPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		pathParts := strings.Split(path, string(os.PathSeparator))

		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		var fPath string

		if pathParts[len(pathParts)-1] == "index.json" {
			fPath = "index.json"
		} else {
			fPath = pathParts[len(pathParts)-2] + "/" + pathParts[len(pathParts)-1]
		}

		fmt.Println(fPath)

		if _, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET")),
			Key:    aws.String(os.Getenv("S3_PREFIX") + "/" + fPath),
			Body:   file,
		}); err != nil {
			panic(err)
		}

		return nil
	})

}
