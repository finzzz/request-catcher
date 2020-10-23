package main

// TODO
// request handling security
// auto cleanup log file if reached limit

import(
    "fmt"
    "log"
    "os"
    "flag"
    "strings"
    "bytes"
    "time"
    "net/http"
    "sync"
    "crypto/tls"
)

var addr string
var https string
var pubkey string
var privkey string
var output_file string
var redirect string

func main(){
    flag.StringVar(&addr, "addr", "127.0.0.1:80", "Bind to \"host:port\"")
    flag.StringVar(&https, "https", "", "HTTPS port to bind")
    flag.StringVar(&pubkey, "pubkey", "", "TLS public key")
    flag.StringVar(&privkey, "privkey", "", "TLS private key")
    flag.StringVar(&output_file, "output", "request.log", "Output file")
    flag.StringVar(&redirect, "redirect", "", "Redirect to somewhere")
    flag.Parse()

    worker := new(sync.WaitGroup)
    worker.Add(2)

    http_handler := http.HandlerFunc(catch)
    http_server := &http.Server{
            Addr:           addr,
            Handler:        http_handler,
            ReadTimeout:    10 * time.Second,
            WriteTimeout:   10 * time.Second,
            IdleTimeout:    30 * time.Second,
    }

    // spawn http server
    go func() { 
        log.Fatal(http_server.ListenAndServe()) 
        worker.Done()
    }()

    if https != "" {
        if pubkey == "" || privkey == "" {
            fmt.Println("HTTPS needs public and private key")
            os.Exit(1)
        }
        host := strings.Split(addr,":")[0]
        https_handler := http.HandlerFunc(catch)
        cert, _ := tls.LoadX509KeyPair(pubkey, privkey)
        https_server := &http.Server{
                Addr:           host + ":" + https,
                Handler:        https_handler,
                TLSConfig:      &tls.Config{
                    Certificates:   []tls.Certificate{ cert },
                },
                ReadTimeout:    10 * time.Second,
                WriteTimeout:   10 * time.Second,
                IdleTimeout:    120 * time.Second,
        }

        // spawn https server
        go func() { 
            log.Fatal(https_server.ListenAndServeTLS("",""))
            worker.Done()
        }()
    }

    worker.Wait()
}

func catch(w http.ResponseWriter, req *http.Request){
    var payload strings.Builder

    if redirect != "" {
        http.Redirect(w, req, redirect, http.StatusMovedPermanently)
    }

    now := time.Now()
    payload.WriteString("Time: " + now.Format(time.RFC1123) + "\n") // optional
    payload.WriteString("Source: " + req.RemoteAddr + "\n") // optional
    payload.WriteString(req.Method + " " + req.RequestURI + " " + req.Proto + "\n")
    payload.WriteString("Host: " + req.Host + "\n")

    for key, value := range req.Header {
        payload.WriteString(key + ": " + value[0] + "\n")   
    }

    buf := new(bytes.Buffer)
    buf.ReadFrom(req.Body)
    payload.WriteString("\n" + buf.String() + "\n")

    f, err := os.OpenFile(output_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Println(err)
    }

    defer f.Close()

    if _, err := f.WriteString(payload.String()); err != nil {
        log.Println(err)
    }
}
