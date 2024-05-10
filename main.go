package main

import (
  "fmt"
  "net/http"
)

type apiConfig struct{
  fileServerHits int
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler{
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
    cfg.fileServerHits++
    fmt.Println("This is the server hit", cfg.fileServerHits)
    next.ServeHTTP(w, r)
  })

}


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
  var apiCfg apiConfig
  mux := http.NewServeMux()
  // mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request){
  //   fmt.Println("Asking for home page huh")
  //   http.FileServer(http.Dir("."))
  //   return
  // })
  fmt.Println("Trying")
  mux.Handle("/app/*", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
  mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request){
    w.WriteHeader(200)
    html_text := `
<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>
`
    serve_html := []byte(fmt.Sprintf(html_text, apiCfg.fileServerHits))
    _, err := w.Write(serve_html)
    if err != nil{
      fmt.Println("Couldn't load metrics")
    }
  })
  mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
    w.WriteHeader(200)
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    body := []byte("OK")
    _, err := w.Write(body)
    if err != nil{
      fmt.Println("Error sending Body")
    }
  })
  mux.HandleFunc("GET /api/metrics", apiCfg.hitsCalculator)
  mux.HandleFunc("/api/reset", apiCfg.resetHits)


  //corsMux := middlewareCors(mux)
  var server http.Server
  server.Addr = ":8080"
  server.Handler = mux
  server.ListenAndServe()

}


func (cfg *apiConfig) hitsCalculator(w http.ResponseWriter, req *http.Request){
  w.WriteHeader(200)
  w.Header().Add("Content-Type", "text/plain; charset=utf-8")
  hits := fmt.Sprintf("Hits: %v", cfg.fileServerHits)
  body := []byte(hits)
  _, err := w.Write(body)
  if err != nil{
    fmt.Println("Error sending Body")
  }

}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, req *http.Request){
  w.WriteHeader(200)
  cfg.fileServerHits = 0
}
