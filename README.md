# dbx
:smile:

## Install

`go get -u github.com/waldurbas/dbx`

## Example

```go
package main

import (
    "fmt"
    "github.com/waldurbas/dba"
    "github.com/waldurbas/dbx/dbt/fdb"	
)

func main() {
  db := fdb.NewDatabase("user:pwd@127.0.0.1:3051/adb.fdb")

  if !db.Connect() {
     db.Fatal("-> connect: ", db.Err)
  }

  defer func {
    fmt.Println("-> disconnect")
    db.Close()
  }()

  q:= db.ExecQ("select ART,AID from DATA rows 10")
  if q.Err != nil {
    db.Fatal("exec: ", q.Err)
  }

  defer q.Close()

  for q.Fetch() {
    q.PrintOut(os.Stdout)
  }
}
````
