package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "log"

  "github.com/gorilla/mux"
)

func main(){
  r := mux.NewRouter()
  r.HandleFunc("/api/get-combinations/{number:[0-9]+}", PermutationHandler)
  port:=":8082"
  log.Printf("Server running on port %s", port)
  http.ListenAndServe(port, &MyServer{r})
}

type MyServer struct {
  r *mux.Router
}

func (s *MyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
  if origin := req.Header.Get("Origin"); origin != "" {
    rw.Header().Set("Access-Control-Allow-Origin", origin)
    //        rw.Header().Set("Access-Control-Allow-Origin","*")
    rw.Header().Set("Access-Control-Allow-Methods", "GET")
    rw.Header().Set("Access-Control-Allow-Headers",
      "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-CurrentStep")
    rw.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
  }
  // Stop here if its Preflighted OPTIONS request
  if req.Method == "OPTIONS" {
    return
  }
  // Let Gorilla work
  //rw.Header().Set("Content-Type", "application/json")
  s.r.ServeHTTP(rw, req)
}


var allPerm []string
func PermutationHandler(rw http.ResponseWriter, r *http.Request){
  params:=mux.Vars(r)
  number:="12345"
  input:=[]rune(number)
  allPerm=[]string{}
  perm(input, 0, len(input)-1)
  fmt.Println(allPerm)
  js,err:=json.Marshal(allPerm)
  if err!=nil{
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Write([]byte(js))
}


func perm(input []rune, begin int , end int){
  
  if begin==end {
    val:=fmt.Sprint(string(input))
    allPerm=append(allPerm,val)
  } else {

    for x:=begin; x<=end; x++{
      input[begin], input[x]=input[x], input[begin]
      fmt.Println(x,begin,end,string(input[begin]),string(input[end]),",",string(input))
      perm(input, begin+1,end)
      input[begin], input[x]=input[x], input[begin]
    }
  }
}
