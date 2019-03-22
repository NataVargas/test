# RSA key pair REST API 

This example is an Rest API that has an endpoint to generate a RSA key pair. The endpoint save the name of the key, the public key and the private key in a database, where de private one is stored with AES-256.

It also has:

1. An endpoint that allows you to list the keys and search by name.
2. An endpoint that receives a plain text and the ID of the key, and return the encrypted text with the key.
3. An endpoint that receives the encrypted text and the ID of the key, and return the plain text.

It was used:
1. Go language.
2. CockroachDB
3. Vue.js and bootstrap-vue.js
4. API Router: chi

**You must install CockroachDB and go language to use this example.**

## Install CockroachDB on Windows

1. Download and extract the CockroachDB v2.1.6 archive for Windows (from [here](https://www.cockroachlabs.com/docs/v2.1/install-cockroachdb-windows.html)).

2. Open Command Prompt, navigate to the directory containing the binary, and make sure the CockroachDB executable works:

`C:\cockroach-v2.1.6.windows-6.2-amd64> .\cockroach.exe version`

3. Execute the binary.

`C:\cockroach-v2.1.6.windows-6.2-amd64> start cockroach.exe`

Note: if you have another operative system, please follow the steps noted in [Cockroach Lab](https://www.cockroachlabs.com/docs/v2.1/install-cockroachdb-windows.html).

## Install the Go pq driver

`go get -u github.com/lib/pq`

## Start a Local Cluster (Insecure)

1. Navigate to the Cockroach directory and start a node.

`cockroach start --insecure --listen-addr=localhost`

2. Open a new terminal and connect to the built-in SQL client.

`cockroach sql --insecure --host=localhost:26257`

3. Create a database and assign a user to have all permissions.

`CREATE DATABASE keypairrsa;`
`CREATE USER IF NOT EXISTS nata;`
`GRANT ALL ON DATABASE bank TO nata;`

## Execute main script

1. Open a new command prompt, navigate to the **test** directory and run go file.

`go run main.go`

Note: you must have a Local Cluster (Insecure) running.
