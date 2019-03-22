package main

import (
  // REST API
  "net/http"
  "path/filepath"
  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"

  // DATABASE
  "database/sql"
  _ "github.com/lib/pq"

  // FORMAT
  "fmt"
  "log"
  "strings"
  "strconv"

  // CRYPTO
  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "crypto/aes"
  "crypto/cipher"
  "crypto/sha256"
  "encoding/base64"
  "encoding/hex"
  "errors"
  "io"
  "os"
)

// Task is a struct containing Task data
type objectKey struct {
  id          int
  name        string
  publickey   string
  privatekey  string
}

func main() {

  //  CRYPTO VARIABLES.
  label := []byte("")
  hash := sha256.New()
  CIPHER_KEY := []byte("0123456789012345")
  // ---------------------------- DATABASE----------------------------------- //
  var userName, dbName string
  userName = "nata"
  dbName = "keypairrsa"

  // Connect to the database.
  db, err := sql.Open("postgres", "postgresql://"+ userName + "@localhost:26257/" + dbName + "?sslmode=disable")
  if err != nil {
    log.Fatal("error connecting to the database: ", err)
  }

  // Create the table.
  if _, err := db.Exec(
    "CREATE TABLE IF NOT EXISTS keypair (id SERIAL PRIMARY KEY, name text, publickey text, privatekey text)"); err != nil {
      log.Fatal(err)
    }

    // ROUTER.
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Get("/get", func(w http.ResponseWriter, r *http.Request) {
      // RETURN ALL DATA FROM TABLE.
      w.Write([]byte(getQuery("SELECT * FROM keypair", db)))
    })

    r.Route("/get/{name}", func(r chi.Router) {
      r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        // DEBUG DATA.
        var param []string
        var name string
        param = strings.SplitAfter(r.URL.Path, "/")
        name = param[2]

        // RETURN KEYS BY NAME.
        w.Write([]byte(getQuery("SELECT * FROM keypair WHERE name='"+ name + "'", db)))
      })
    })

    r.Route("/get/{id}/{text}", func(r chi.Router){
      r.Get("/", func(w http.ResponseWriter, r *http.Request){

        // GETTING DATA.
        id, text := getText(r.URL.Path)
        var kpublic string

        //GETTING THE PUBLIC KEY.
        rows, err := db.Query("SELECT * FROM keypair WHERE id="+id)
        if err != nil {
          log.Fatal(err)
        }

        defer rows.Close()
        for rows.Next() {
          objectKey :=  objectKey{}
          if err := rows.Scan(&objectKey.id, &objectKey.name, &objectKey.publickey, &objectKey.privatekey); err != nil {
            log.Fatal(err)
          }
          kpublic = objectKey.publickey
        }

        decryptedk_public, err := decrypt(CIPHER_KEY, kpublic)
        if err != nil {
          log.Println(err)
        }

        PublicKey, x:= x509.ParsePKCS1PublicKey([]byte(decryptedk_public))
        if x != nil {
          log.Println(x)
        }

        // ENCRYPT MESSAGE.
        message := []byte(text)
        ciphertext, err := rsa.EncryptOAEP(
          hash,
          rand.Reader,
          PublicKey,
          message,
          label)
        if err != nil {
          fmt.Println(err)
          os.Exit(1)
        }

        // RETURN ENCRYPTED MESSAGE.
        w.Write([]byte(fmt.Sprintf("%x", ciphertext)))
      })
    })

    r.Route("/getplain/{id}/{text}", func(r chi.Router){
      r.Get("/", func(w http.ResponseWriter, r *http.Request){

        // GETTING DE TEXT.
        id, text := getText(r.URL.Path)

        //GETTING THE PRIVATE KEY.
        rows, err := db.Query("SELECT * FROM keypair WHERE id="+id)
        if err != nil {
          log.Fatal(err)
        }

        defer rows.Close()

        var kprivate string
        for rows.Next() {
          objectKey :=  objectKey{}
          //var name, publickey, privatekey string
          if err := rows.Scan(&objectKey.id, &objectKey.name, &objectKey.publickey, &objectKey.privatekey); err != nil {
            log.Fatal(err)
          }
          kprivate = objectKey.privatekey
        }

        decryptedk_private, err := decrypt(CIPHER_KEY, kprivate)
        if err != nil {
          log.Println(err)
        }

        PrivateKey, x2:= x509.ParsePKCS1PrivateKey([]byte(decryptedk_private))
        if x2 != nil {
          log.Println(x2)
        }

        // DECRYPT MESSAGE.
        data, err := hex.DecodeString(text)
        if err != nil {
          panic(err)
        }

        plainText, err := rsa.DecryptOAEP(
          hash,
          rand.Reader,
          PrivateKey,
          data,
          label)
        if err != nil {
          fmt.Println(err)
          os.Exit(1)
        }

        // RETURN PLAIN TEXT.
        w.Write([]byte(plainText))
      })
    })

    r.Route("/post/{name}", func(r chi.Router) {
      r.Post("/", func(w http.ResponseWriter, r *http.Request){

        var param []string
        var name, sentence string
        param = strings.SplitAfter(r.URL.Path, "/")
        name = param[2]

        // CREATE KEYS.
        PrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
        if err != nil {
          fmt.Println(err.Error)
          os.Exit(1)
        }
        PublicKey := &PrivateKey.PublicKey

        // ENCRYPT KEYS WITH AES.
        encryptedk1, err := encrypt(CIPHER_KEY, string (x509.MarshalPKCS1PublicKey(PublicKey)))
        if err != nil {
          log.Println(err)
        }

        encryptedk2, err := encrypt(CIPHER_KEY, string (x509.MarshalPKCS1PrivateKey(PrivateKey)))
        if err != nil {
          log.Println(err)
        }

        // INSERT INTO DATABASE.
        sentence = fmt.Sprintf("INSERT INTO keypair (name, publickey, privatekey) VALUES ('%s', '%s', '%s')", name, encryptedk1, encryptedk2)
        if _, err := db.Exec(sentence); err != nil {
          log.Fatal(err)
        }
        fmt.Println("Insertado con éxito")
      })
    })

    workDir, _ := os.Getwd()
    filesDir := filepath.Join(workDir, "public")
    FileServer(r, "/", http.Dir(filesDir))

    http.ListenAndServe(":3333", r)

  }

  func encrypt(key []byte, message string) (encmess string, err error) {
    plainText := []byte(message)

    block, err := aes.NewCipher(key)
    if err != nil {
      return
    }

    //IV needs to be unique, but doesn't have to be secure.
    //It's common to put it at the beginning of the ciphertext.
    cipherText := make([]byte, aes.BlockSize+len(plainText))
    iv := cipherText[:aes.BlockSize]
    if _, err = io.ReadFull(rand.Reader, iv); err != nil {
      return
    }

    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

    //returns to base64 encoded string
    encmess = base64.URLEncoding.EncodeToString(cipherText)
    return
  }

  func decrypt(key []byte, securemess string) (decodedmess string, err error) {
    cipherText, err := base64.URLEncoding.DecodeString(securemess)
    if err != nil {
      return
    }

    block, err := aes.NewCipher(key)
    if err != nil {
      return
    }

    if len(cipherText) < aes.BlockSize {
      err = errors.New("Ciphertext block size is too short!")
      return
    }

    //IV needs to be unique, but doesn't have to be secure.
    //It's common to put it at the beginning of the ciphertext.
    iv := cipherText[:aes.BlockSize]
    cipherText = cipherText[aes.BlockSize:]

    stream := cipher.NewCFBDecrypter(block, iv)
    // XORKeyStream can work in-place if the two arguments are the same.
    stream.XORKeyStream(cipherText, cipherText)

    decodedmess = string(cipherText)
    return
  }

  func FileServer(r chi.Router, path string, root http.FileSystem) {
    if strings.ContainsAny(path, "{}*") {
      panic("FileServer does not permit URL parameters.")
    }

    fs := http.StripPrefix(path, http.FileServer(root))

    if path != "/" && path[len(path)-1] != '/' {
      r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
      path += "/"
    }
    path += "*"

    r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      fs.ServeHTTP(w, r)
      }))
    }

    func getQuery(s string, db *sql.DB) string {
      CIPHER_KEY := []byte("0123456789012345")
      var keys []string
      var result string

      // GET DATA FROM DATABASE.
      rows, err := db.Query(s)
      if err != nil {
        log.Fatal(err)
      }

      defer rows.Close()
      for rows.Next() {
        objectKey :=  objectKey{}
        if err := rows.Scan(&objectKey.id, &objectKey.name, &objectKey.publickey, &objectKey.privatekey); err != nil {
          log.Fatal(err) }

          decryptedk_public, err := decrypt(CIPHER_KEY, objectKey.publickey)
          if err != nil {
            log.Println(err)
          }

          decryptedk_private, err := decrypt(CIPHER_KEY, objectKey.privatekey)
          if err != nil {
            log.Println(err)
          }

          json:="{\"id\":\"" + strconv.Itoa(objectKey.id) +
            "\",\"name\": \"" + objectKey.name + "\",\"publickey\": \"" +
            fmt.Sprintf("%x", decryptedk_public) + "\",\"privatekey\": \"" +
            fmt.Sprintf("%x", decryptedk_private) + "\"}"

            keys = append(keys, json)
          }

          // CONSTRUCT JSON ARRAY.
          result = "["
          for i := 0; i < len(keys); i++ {
            if v := i; v < len(keys)-1 {
              result = result + keys[i] + ", "
              } else {
                result = result + keys[i] + "]"
              }
            }
            return result
          }

          func getText(url string) (string, string) {
            var param1, param2 []string
            var id, text string
            text = ""

            param1 = strings.SplitAfter(url, "/")
            param2 = strings.SplitAfter(strings.Trim(param1[3], "/"), "_")
            id = strings.Trim(param1[2], "/")

            for i := 0; i < len(param2); i++ {
              if v := i; v < len(param2)-1 {
                text = text +  strings.Trim(param2[i], "_") + " "
                } else {
                  text = text + param2[i]
                }
              }
              return id, text
            }
