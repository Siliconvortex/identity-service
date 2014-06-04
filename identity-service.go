package main

import (
  "fmt"
  "time"
  "encoding/json"
  "net/http"
  "github.com/dchest/uniuri"
  "labix.org/v2/mgo"
)

func main() {
  session, err := mgo.Dial("localhost")
  if err != nil {
    panic(err)
  }
  defer session.Close()

  // Optional. Switch the session to a monotonic behavior.
  session.SetMode(mgo.Monotonic, true)

  // create the collection
  token_collection := session.DB("test").C("tokens")

  // ensure the token id index exists
  index := mgo.Index{
    Key: []string{"id"},
    Unique: true,
  }
  err2 := token_collection.EnsureIndex(index)
  if err2 != nil {
    panic(err2)
  }

  http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    s, _ := json.Marshal(getToken(token_collection))
    fmt.Fprint(w, string(s))
    return
  })

  http.ListenAndServe(":8080", nil)
}

const Token_Length = 256
const Token_Duration = time.Hour

type Token struct {
  Id string `json:"id" bson:"id"`
  Expiration time.Time `json:"expiration" bson:"expiration"`
}

func getToken(token_collection *mgo.Collection) Token {
  // generate and persist unique token
  // retry if token collision
  t := TokenGenerator()
  for err := token_collection.Insert(t); err != nil ; err = token_collection.Insert(t) { 
    if !mgo.IsDup(err) {
      // if this is not a "duplicate key" error, panic
      panic(err)
    } else {
      // otherwise, it's just a random dupe, find another key and try again
      fmt.Printf("dup - generating another token")
      t = TokenGenerator()
    }
  }

  return t
}

func TokenGenerator() Token {
  answer := Token{
    Id : uniuri.NewLen(Token_Length),
    Expiration: time.Now().Add(Token_Duration).UTC(),
  }

  return answer
}
