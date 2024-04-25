package main

import (
  "fmt"
  "net/http"
)

func middlewareCors(next http.Handler) http.Handler{
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
    w.Header().Set("Access-Control-Allow-Headers", "*")
    if r.Method == "OPTIONS"{
      w.WriteHeader(http.StatusOK)
      return
    }
    next.ServeHTTP(w, r)
  })
}

func main(){
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request){
    fmt.Println("Serving not found")
    http.NotFound(w, req)
    return
  })
  corsMux := middlewareCors(mux)
  var server http.Server
  server.Addr = ":8080"
  server.Handler = corsMux
  server.ListenAndServe()

}
