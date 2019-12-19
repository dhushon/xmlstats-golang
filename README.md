# xmlstats-golang
Go Example to access XMLStats.com JSON streams for sporting events and box scores

There remains a problem with using time.Time JSON marshalling of data from XMLStats, work around in place.

To build:
```
go build xmlstats.go
```

To run:
register with xmlstats to pickup your authorization string
```
cp env-orig.bash env.bash
```
edit env.bash to add authorization string and email address associated with authorization string 
```
source env.bash //with your credentials
go run xmlstats
```
if you get this response, you have not properly registered, or setup the environment variables to ensure that they carry into the xmlstats function:
```
API server listening at: 127.0.0.1:3166
Loading environment configuration constants
XMLSTATS_URL not found, should include your website or email credential
XMLSTATS_BEARERTOKEN not found, you should get your token from xmlstats registration 
XMLSTATS_USERAGENT not found, should include your website or email credential
doing HTTP GET
The HTTP request header building failed with error 403 Forbidden : {
  "error" : {
    "code" : "403",
    "description" : "Invalid user agent: null. Please refer to the user agent guidelines specified at https://erikberg.com/api#ua."
  }
}
Terminating the application...
```

