default:
	go build -o passrider database.go http.go api.go parsing.go misc.go passrider.go structs.go
report: 
	go build -o report database.go misc.go structs.go report.go
install:
	go install passrider.go database.go http.go api.go parsing.go misc.go structs.go


