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

 
How to create test data withour a client:

  open slite DB by running sqlite grocery-manager-go.sqlite
  On the sqlite shell issue the following SQL statements:

  INSERT INTO list (title) VALUES ("test_list1");

  INSERT INTO grocery_item (name,current,minimum) VALUES ("Oatmeal",1,1);
  
  INSERT INTO list_item (grocery_item_id,quantity,checked,position,on_list) VALUES (1,1,1,13,1);


