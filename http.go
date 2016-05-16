package main

import (
     "net/http"
     "github.com/gorilla/sessions"
     "net/http/cookiejar"
     "strings"
     "bytes"
     "fmt"
     //"time"
     "io/ioutil"
     "net/url"
)

func doGet(url string, client *http.Client) string {
    resp, err := client.Get(url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    body, _ := ioutil.ReadAll(resp.Body)
    return string(body)
}

func doPost(url string, client *http.Client, data string, postType PostType) string {
    var req *http.Request
    
    if postType == Normal {
        postData := parsePostdata(data)
        req,_ = http.NewRequest("POST", url, strings.NewReader(postData.Encode()))
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    } else if postType == JSON {
        req, _ = http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
        req.Header.Set("Content-Type", "application/json")
    } else {
        req,_ = http.NewRequest("POST", url, strings.NewReader(data))
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    }
    
    req.Header.Set("User-Agent","Mozilla/5.0 (Macintosh; Intel Mac OS X 10.11; rv:44.0) Gecko/20100101 Firefox/44.0")
    
    //debug(httputil.DumpRequestOut(req, true))
    //spew.Dump(client)
    
    resp, err := client.Do(req)
    if err != nil {
             panic(nil)
    }
    
    body, _ := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    
    return string(body)
}

func returnJsonError(message string) string {
    return "{\"error\":\"" + message + "\"}"
}


func restoreCookieJar(session *sessions.Session, uri string) {
    url, _ := url.Parse(uri)

    existingCookies := []*http.Cookie{}

    for _, cookie := range strings.Split(session.Values["cookie"].(string), ";") {
        crumb := strings.Split(cookie, "=")
        if len(crumb) == 2 {
            newCookie := http.Cookie{Name: crumb[0], Value: crumb[1]}
            existingCookies = append(existingCookies, &newCookie)
        }
    }
    
    j, err := cookiejar.New(nil)
    if err != nil {
        panic(err)
    }

    client = &http.Client{Jar: j}
    client.Jar.SetCookies(url, existingCookies)
}

type HttpResp struct {
    Id   string
    Resp *http.Response
    Err  error
}
 
func AsyncGet(urls map[string]string) []*HttpResp {
    ch := make(chan *HttpResp)
    responses := []*HttpResp{}
 
    for track_id, url := range urls {
 
        go func(i, u string) {
            fmt.Printf("Getting: %s\n", u)
            resp, err := client.Get(u)
            ch <- &HttpResp{i, resp, err}
        }(track_id, url)
    }
 
loop:
    for {
        select {
        case r := <-ch:
            responses = append(responses, r)
            if len(responses) == len(urls) {
                break loop
            }
        //case <-time.After(50 * time.Millisecond):
        //    fmt.Printf(".")
        }
    }
    return responses
}