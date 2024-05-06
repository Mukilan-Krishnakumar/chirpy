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
  // mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request){
  //   fmt.Println("Asking for home page huh")
  //   http.FileServer(http.Dir("."))
  //   return
  // })
  fmt.Println("Trying")
  mux.Handle("/app/*", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
  mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
    w.WriteHeader(200)
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    body := []byte("OK")
    _, err := w.Write(body)
    if err != nil{
      fmt.Println("Error sending Body")
    }
  })
  //corsMux := middlewareCors(mux)
  var server http.Server
  server.Addr = ":8080"
  server.Handler = mux
  server.ListenAndServe()

}
