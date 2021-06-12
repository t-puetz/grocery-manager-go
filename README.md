# grocery-manager-go
Grocery manager server written in Go

The webserver will start at port 8080 by default when no --port flag <port number> was provided.
  
Usage:
  
  With Go you can the run subcommand which basically is build followed directly by execute (executable resides in /tmp then)
  or you can use build <source files> and then run the resulting executable that will be put into your project path.
  
  go run main.go
  
  go run main.go --port 8082
  
  go build main.go and then run by executing produced main executable
  
  go build main.go --port 8081
