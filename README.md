# LAWLIPOPS

### How to deploy the project on the server
After logging in to the server, change to the project directory
```bash
cd go/src/github.com/imkostas/LAWLIPOPS
```

Pull the latest changes from github
```bash
git pull
```

Build a new executable to include the changes
```bash
go build
```

Stop the previous service
```bash
sudo service LAWLIPOPS stop
```

Start the service again
```bash
sudo service LAWLIPOPS start
```
