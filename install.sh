# deploy.sh
#!/bin/bash
go build -o locc
mv locc /usr/local/bin/locc
