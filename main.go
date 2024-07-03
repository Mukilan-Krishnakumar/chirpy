package main

import (
  "fmt"
  "net/http"
  "encoding/json"
  "strings"
  "sync"
  "os"
  "errors"
)

type apiConfig struct{
  fileServerHits int
}

type DB struct{
  path string
  mux *sync.RWMutex
}


type parameters struct{
  Body string `json:"body"`
}

type cleanedParameters struct{
  CleanedBody string `json:"cleaned_body"`
}

type Chirp struct{
  Id int `json:"id"`
  Body string `json:"body"`
}

type DBStructure struct{
  Chirps map[int]Chirp `json:"chirps"`
}

type errorJSON struct{
  Error string `json:"error"`
}

type validJSON struct{
  Valid bool `json:"valid"`
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
  mux.HandleFunc("POST /api/chirps/",createChirp)

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

func createChirp(w http.ResponseWriter, req *http.Request){
  fmt.Println("Creating Chirp..")

  decoder := json.NewDecoder(req.Body)
  params := parameters{}
  err := decoder.Decode(&params)

  if err != nil{
    respondWithError(w, 500, "Something went wrong")
    return
  }

  if len(params.Body) > 140{
    respondWithError(w, 400, "Chirp is too long")
    return
  }
    
  cleaned_response := badWordReplacement(params.Body)
  cleaned_body := cleanedParameters{
    CleanedBody: cleaned_response,
  }
  db, err :=  NewDB("./database.json")
  if err != nil{
    fmt.Println(err)
  }

  go db.CreateChirp(cleaned_response)

  respondWithJSON(w, 200  , cleaned_body)
  return

}

func respondWithError(w http.ResponseWriter, code int, msg string){
    error_msg := errorJSON{
      Error: msg,
    }
    error_json, _ := json.Marshal(error_msg)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(error_json)
    return

}


func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
    responseJson, _ := json.Marshal(payload)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(responseJson)
    return
}

func badWordReplacement(input_text string) (cleaned_response string){ 
  curse_words := []string{"kerfuffle", "sharbert", "fornax"}
  curse_words_ctr := 0
  cleaned_words := []string{}

  for _, word := range strings.Split(input_text, " "){
    check_word := strings.ToLower(word)
    replace_word := word
    for _, curse_word := range curse_words{
      if curse_word == check_word{
        replace_word = "****"
      }
    }
    if replace_word != word{
        curse_words_ctr = curse_words_ctr + 1
        cleaned_words = append(cleaned_words, replace_word)
      }else{
        cleaned_words = append(cleaned_words, word)
      }
  }

  return strings.Join(cleaned_words, " ")

}

func NewDB (path string) (*DB, error){
  _, err := os.ReadFile(path)
  if errors.Is(err, os.ErrNotExist){
    fmt.Println("File does not exist")
    err := os.WriteFile(path, []byte(""), 0666)
    if err != nil{
      empty_db := &DB{}
      fmt.Println("Couldn't write file")
      return empty_db, errors.New("Couldn't write file")
    }
  } 
  mux := &sync.RWMutex{}
  db := &DB{path: path, mux: mux}
  return db, nil
}


func (db *DB) CreateChirp(body string){
  // Save Chirps in DB as a Map Structure
  db.mux.Lock()
  _, err := os.ReadFile(db.path)
  if errors.Is(err, os.ErrNotExist){
    fmt.Println("File did not exist")
     }else{
    new_chirp := Chirp{
      Id : 1,
      Body: body,
    }
  chirp_json, _ := json.Marshal(new_chirp)

  err := os.WriteFile(db.path, chirp_json, 0666)
    if err != nil{
      fmt.Println("Couldn't create file")
    }


  }
  db.mux.Unlock()
}
