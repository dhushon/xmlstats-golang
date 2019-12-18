# xmlstats-golang
Go Example to access XMLStats.com JSON streams for sporting events and box scores

There remains a problem with using time.Time JSON marshalling of data from XMLStats, work around in place.

To build:
go build xmlstats.go

To run:
source env.bash //with your credentials
go run xmlstats
