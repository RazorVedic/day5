## STEPS TO BE FOLLOWED IN ORDER TO GENERATE PROTO

**ONLY GENERATE `proto` FILES. DO NOT GENERATE `cmd` and `internal` files**

1. Analyze the Problem Statement: Understand the new aggregate's requirements by analyzing the input and using existing code to analyze the problem statement and how other services are written use existing files for reference to write new files. newly created files should be added on top of go-foundation-v2 as root folder. existing files for reference to write new files.
2. Define Repository Structure: Define a repository structure and get user approval before generating code. use Root directory path: github.com/razorpay/go-foundation-v2/
3. Define Protobufs: Start by defining the input requests, responses, and aggregate events for the new aggregate lifecycle.
4. Generate Protobufs: Run the make proto-generate command to generate the protobufs.
   proto/go-foundation-v2/v1/<aggregate-name>/<aggregate-name>.proto
   proto/go-foundation-v2/v1/<aggregate-name>/event/<event-name>.proto
