# Tella Upload Server

### Protocol
For any request sent to the server, the client is required to authenticate using HTTP Basic auth.

#### Getting file information
At any time client can issue HTTP HEAD request and get current file information from server.
```http request
HEAD /<file> HTTP/1.1
authorization: Basic <base64_auth>
```
Server will reply with current file size in `content-length` header. If file is unknown to the server, it will reply with zero length.
```http request
HTTP/1.1 200 OK
content-length: <file size>
```

#### Uploading file data
At any time client can append data to the files on TUS server using HTTP PUT requests.
```http request
PUT /<file> HTTP/1.1
authorization: Basic <base64_auth>
content-length: <upload_length>
content-type: <uplod_media_type>

<upload_body>
```
Upload requests can be repeated and the client needs to check the current size on the server using HEAD request. For any error reported by TUS server, the client needs to repeat HEAD request to get accurate offset to start upload from.

#### Closing file
After upload of data is complete without errors the client must close the file on TUS server. The server will deny any further appending on closed files.
```http request
POST /<file> HTTP/1.1
authorization: Basic <base64_auth>
content-length: 0
```
