package main

import (
  "fmt"
  "net/http"
  "encoding/json"
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
  mux.HandleFunc("POST /api/validate_chirp",validateChirp)


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

func validateChirp(w http.ResponseWriter, req *http.Request){
  fmt.Println("Validating Chirp..")
  type parameters struct{
    Body string `json:"body"`
  }

  type errorJSON struct{
    Error string `json:"error"`
  }

  type validJSON struct{
    Valid bool `json:"valid"`
  }

  decoder := json.NewDecoder(req.Body)
  params := parameters{}
  err := decoder.Decode(&params)

  if err != nil{
    fmt.Println("Error decoding parameters")
    error_msg := errorJSON{
      Error: "Something went wrong",
    }
    error_json, _ := json.Marshal(error_msg)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
    w.Write(error_json)
    return
  }

  if len(params.Body) > 140{
    error_msg := errorJSON{
      Error: "Chirp is too long",
    }
    error_json, _ := json.Marshal(error_msg)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(400)
    w.Write(error_json)
    return
  }

  valid_msg := validJSON{
    Valid: true,
  }
  valid_json, _ := json.Marshal(valid_msg)
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(200)
  w.Write(valid_json)
  return

}

