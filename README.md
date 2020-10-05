# Direct Upload Server

### Running the server
#### Download and build Direct-Upload server
To run Direct-Upload server container, first clone the Direct-Upload repository and build docker image:
```shell script
git clone https://github.com/Horizontal-org/direct-upload.git
cd direct-upload
docker build -t direct-upload .
```
Before running actual Direct-Upload server container run following command to get basic usage help:
```shell script
docker run --rm --name direct-upload direct-upload -h
```
Output will show available Direct-Upload commands and flags:
```shell script
Usage:
  direct-upload [command]

Available Commands:
  auth        Manage user authentication.
  help        Help about any command
  server      Start Tella Direct Upload Server

Flags:
  -h, --help         help for direct-upload
  -r, --rpc string   address for rpc server to bind to (default "127.0.0.1:1206")
  -v, --verbose      make logging more talkative

Use "direct-upload [command] --help" for more information about a command.
```
Command `server` starts the server, with following options:
```shell script
Start Tella Direct Upload Server

Usage:
  direct-upload server [flags]

Flags:
  -a, --address string    address for server to bind to (default ":8080")
  -c, --cert string       certificate file, ie. ./fullcert.pem
  -d, --database string   direct-upload database file (default "./direct-upload.db")
  -f, --files string      path where direct-upload server stores uploaded files
  -h, --help              help for server
  -k, --key string        private key file, ie. ./key.pem

Global Flags:
  -r, --rpc string   address for rpc server to bind to (default "127.0.0.1:1206")
  -v, --verbose      make logging more talkative
```
To run container with started Direct-Upload server daemon inside with default parameters and with SSL enabled, 
run the following command:
```shell script
docker run -d -v /opt/data/:/data -p 443:8080 --name direct-upload \
  direct-upload server -c /data/fullchain.pem -k /data/privkey.pem
``` 
File `docker-config.yml` that is copied to the container during Docker build step contains a server's
default options. Flags `-c` and `-k` for enabling HTTPS are optional but are highly recommended as 
protocol uses HTTP Basic authentication for user authentication. Direct-Upload server inside container is 
using `/data` mount volume and that is where all data should reside.

To check server logs, run following command:
```shell script
docker logs -f --tail=100 direct-upload
```
After successful server start, users using the service needs to be provided using `auth` command in the running container.
To get help on the `auth` command run following command while direct-upload container is running:
```shell script
docker exec -it direct-upload direct-upload auth --help
```
Command will print usage help:
```shell script
Manage user authentication.

Usage:
  direct-upload auth [command]

Available Commands:
  add         Add user authentication if doesn't already exists. Will prompt for password.
  backup      Backup auth database.
  del         Delete user authentication.
  list        List usernames.
  passwd      Change user authentication. Will prompt for password.

Flags:
  -h, --help   help for auth

Global Flags:
  -r, --rpc string   address for rpc server to bind to (default "127.0.0.1:1206")
  -v, --verbose      make logging more talkative

Use "direct-upload auth [command] --help" for more information about a command.
```
To start using server, lets create user:
```shell script
docker exec -it direct-upload direct-upload auth add <username>
```
To list available usernames:
```shell script
docker exec -it direct-upload direct-upload auth list
```
Server allows for changing user password, removing user and backing up user database.


### Protocol
For any request sent to the server, the client is required to authenticate using HTTP Basic auth. 
For any request server will reply with 401 status on bad credentials or with status 400 if username 
is not valid. Valid username start with letter, number or underscore character and can contain
letters, numbers and `_-.@` characters.

#### Getting file information
At any time client can issue HTTP HEAD request and get current file information from the server.
```http request
HEAD /<file> HTTP/1.1
authorization: Basic <base64_auth>
```
Server will reply with current file size in `content-length` header. If file is unknown to the server, 
it will reply with zero length.
```http request
HTTP/1.1 200 OK
content-length: <file size>
```

#### Uploading file data
At any time client can append data to the files on Direct-Upload server using HTTP PUT requests.
```http request
PUT /<file> HTTP/1.1
authorization: Basic <base64_auth>
content-length: <upload_length>
content-type: <uplod_media_type>

<upload_body>
```
Upload requests can be repeated and the client needs to check the current size on the server using 
HEAD request. For any error reported by Direct-Upload server, the client needs to repeat HEAD request to get 
accurate offset to start upload from.

#### Closing file
After upload of data is complete without errors the client must close the file on Direct-Upload server. The 
server will deny any further PUT appending on closed files with 409 status.
```http request
POST /<file> HTTP/1.1
authorization: Basic <base64_auth>
content-length: 0
```
