package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"net.craftengine/uploader/internal/app"
)

var version string

// Define a flag for the Minecraft version
var versionFlag = flag.String("version", "1.21.4", "Minecraft Version")
var s3Client *s3.Client
var s3Context = context.Background()

func main() {
	fmt.Println("Asset Uploader Version: " + version)

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Parse command-line flags
	flag.Parse()

	// Retrieve the environment variable for S3_PATH_STYLE and parse it to a boolean
	// This variable determines whether to use path-style or virtual-hosted-style URLs for S3 requests
	pathStyle, _ := strconv.ParseBool(os.Getenv("S3_PATH_STYLE"))

	// Create a new S3 client with specified options
	s3Client = s3.New(s3.Options{
		BaseEndpoint: aws.String(os.Getenv("S3_ENDPOINT")),
		Region:       os.Getenv("S3_REGION"),
		// Provide static credentials for accessing the S3 service, retrieved from environment variables
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			""),
	}, func(o *s3.Options) {
		// Configure the client to use path-style URLs based on the parsed boolean value
		o.UsePathStyle = pathStyle
	})

	app.VersionFlag = *versionFlag
	app.S3Client = s3Client
	app.S3Context = s3Context

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
	app.UnpackZip(tempDir)
}
