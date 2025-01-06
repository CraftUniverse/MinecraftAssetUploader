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

// Define a flag for the Minecraft version
var versionFlag = flag.String("version", "1.21.4", "Minecraft Version")

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Parse command-line flags
	flag.Parse()

	// Create a temporary directory for the downloaded files
	var tempDir = filepath.Join(os.TempDir(), "/CEMCAU_"+*versionFlag)
	// Construct the download URL for the Minecraft assets
	var dlURL = "https://github.com/InventivetalentDev/minecraft-assets/zipball/refs/heads/" + *versionFlag

	fmt.Println("Dowloading from " + dlURL)

	// Create a file to save the downloaded zip
	out, err := os.Create(tempDir + ".zip")

	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Download the zip file
	resp, err := http.Get(dlURL)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Copy the downloaded data to the file
	io.Copy(out, resp.Body)

	// Unpack the downloaded zip file
	unpackZip(tempDir)
}

// Unpacks the zip file to the specified directory
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

		// Find the root directory of the extracted files
		for i, part := range fileParts {
			if strings.HasPrefix(part, "InventivetalentDev-minecraft-assets") {
				filePath = strings.ReplaceAll(filePath, part, "")
				rootIndex = i
				break
			}
		}

		// Skip certain directories and files
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

		// Validate the file path
		if !strings.HasPrefix(filePath, filepath.Clean(tempDir)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}

		// Create directories if necessary
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		fmt.Println(file.Name)

		// Extract the file
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

// Structure for the index file
type IIndex struct {
	Hash     string   `json:"hash"`
	Path     string   `json:"path"`
	Length   int64    `json:"length"`
	Versions []string `json:"versions"`
}

// Generates the index file and objects
func generateObjects(tempDir string) {
	var index []IIndex
	var objectPath = tempDir + "_objects"

	// Walk through the assets directory
	filepath.Walk(filepath.Join(tempDir, "assets", "minecraft"), func(path string, info os.FileInfo, err error) error {
		var fPath = strings.ReplaceAll(strings.ReplaceAll(path, tempDir, ""), "\\", "/")

		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)

		fileParts := strings.Split(fPath, string(os.PathSeparator))

		if err != nil {
			panic(err)
		}

		// Calculate the MD5 hash of the file
		h := md5.New()
		if _, err := io.Copy(h, file); err != nil {
			panic(err)
		}

		hash := fmt.Sprintf("%x", h.Sum(nil))

		// Create the directory for the hash
		var basePath = filepath.Join(objectPath, hash[0:2])
		os.MkdirAll(basePath, os.ModePerm)

		// Read the file content and write it to the new location
		fileContent, _ := os.ReadFile(path)
		os.WriteFile(filepath.Join(basePath, hash), fileContent, 0644)

		fmt.Println(fileParts[len(fileParts)-1] + " -> " + hash[0:2] + "/" + hash)

		// Add the file to the index
		index = append(index, IIndex{
			Hash:     hash,
			Path:     fPath,
			Length:   info.Size(),
			Versions: []string{*versionFlag},
		})
		return nil
	})

	// Marshal the index to JSON and write it to a file
	jsonString, _ := json.Marshal(index)
	indexFile, err := os.Create(filepath.Join(objectPath, "index.json"))

	if err != nil {
		panic(err)
	}
	defer indexFile.Close()

	io.WriteString(indexFile, string(jsonString[:]))

	// Remove the temporary directory
	os.Remove(tempDir)
	uploadToS3(objectPath)
}

// Uploads the generated objects to S3
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

	// Walk through the object directory and upload each file to S3
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

		fmt.Println("Uploading: " + fPath)

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
