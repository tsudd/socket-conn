# socket-conn
Server/client communication powered by native *net* Go package using UDP only.

## Main features
- communication using UDP only
- configuration handling
- connected users managing
- own simple protocol
- chat between users
- chat history restoring after user reconnection
- heartbeating of connections

## How to launch
Basic sequence to test apps is the next:
1. Start a server using default or custom configuration
2. Start a client for the first user in on terminal
3. Start a client for the second user in other terminal
4. Try to establish a connection between them by using command `chat <username>` and have a little chat
### Server configuration and start
Server placed in the *server* folder. By simply compiling and starting *server.go* server will be online and start to listen for connections from clients. 
You can use `go run` command to start the server. Optionally, you are able to provide a config file. In other case, server will use *default_config.yaml*.
```
cd server
go run server.go <config file name, placed in the config folder>
```
Or, you can build executable and work using it:
```
cd server
go build server.go server
./server <config name>
```

### Client configuration and start
Don't forget to prepare a config for each client and place it in the *config* folder of the client app. Config is **mandatory** parameter for client start. Same as for server, you can *run* or *build* the app.
```
cd client
go run client.go <config name>
```
For build:
```
cd client
go build client
./client <config name>
```
### Client usage
After a client started, it will take configurations from the provided config file and try to connect to the server, which should be online to establish a connection.
Then you will be able to use internal commands to communicate with the server. Some of them:

- `help` - get info about all existing commands
- `online` - get a list of connected to the server users
- `chat <username>` - start chatting with one of the connected users
- `exit` - exit client app
