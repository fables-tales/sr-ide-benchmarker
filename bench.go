package main

import (
    "fmt"
    "log"
    "strings"
    "bytes"
    "time"
    "net/http"
    "net/http/cookiejar"
    "github.com/howeyc/gopass"
)

func srClient() *http.Client {
    jar, err :=  cookiejar.New(nil)
    if err != nil {
        panic(err)
    }

    client := &http.Client{
        Jar: jar,
    }

    return client
}

type requestResult struct {
    GoRoutine int
    Path string
    Payload string
    ResponseTimeMS int64
}

type srRequest struct {
    Path string
    Payload string
}

func srPostRequest(client *http.Client, endpoint string, request string) string {
    reader := strings.NewReader(request)
    fullUrl := fmt.Sprintf("https://www.studentrobotics.org/ide/control.php%s", endpoint)
    response, err := client.Post(fullUrl, "text/json", reader)
    if err != nil {
        log.Print(err)
        return ""
    } else {
        buf := new(bytes.Buffer)
        buf.ReadFrom(response.Body)
        s := buf.String() // Does a complete copy of the bytes in the buffer.
        return s
    }
}

func srAuthenticate(username string, password string) *http.Client {
    client := srClient()
    body := fmt.Sprintf("{\"username\": \"%s\", \"password\":\"%s\"}", username, password)
    srPostRequest(client, "/auth/authenticate", body)
    return client
}

func srInfo(client *http.Client, team string, project string) string {
    body := fmt.Sprintf("{\"team\": \"%s\", \"project\":\"%s\"}", "HRS", "state-machine")
    return srPostRequest(client, "/proj/info", body)
}

func getTimeMilis() int64 {
    return time.Now().UnixNano()/1e6
}

func doRequestWork(results chan requestResult, clientIDs chan int, requests chan srRequest, username string, password string) {
    id := <- clientIDs
    client := srAuthenticate(username, password)
    for request := range requests {
        start := getTimeMilis()
        _ = srPostRequest(client, request.Path, request.Payload)
        end := getTimeMilis()
        r := requestResult{id,request.Path,request.Payload,end-start}
        results <- r
    }
}

func main() {
    username := "sphippen"
    fmt.Println("Password: ")
    password := string(gopass.GetPasswd())

    workers := 50

    results := make(chan requestResult, workers)
    ids := make(chan int, workers)
    requestsToMake := []srRequest{
        srRequest{"/file/compat-tree", "{\"team\":\"HRS\",\"project\":\"state-machine\",\"rev\":\"HEAD\",\"path\":\".\"}"},
    }

    requests := make(chan srRequest, workers*len(requestsToMake))

    fmt.Println("hi")
    for i := 0; i < workers; i++ {
        ids <- i
        for request := range requestsToMake {
            requests <- requestsToMake[request]
        }
    }
    close(requests)
    close(ids)
    for i := 0; i < workers; i++ {
        go doRequestWork(results, ids, requests, username, password)
    }
    for i := 0; i < workers*len(requestsToMake); i++ {
        fmt.Println(<- results)
    }
}
