/*
 * Doc Service
 * handle doc save, read, and retrieve from S3
 *
 * Dependencies:
 * "github.com/googollee/go-rest"
 * "github.com/goamz/goamz/aws"
 * "github.com/goamz/goamz/s3"
 * 
 * Original 01/26/2016
 * Created by Todd Moses
 */

package main

import (
    "os"
    "fmt"
    "errors"
    "encoding/json"
    "encoding/base64"
    "io/ioutil"
    "net/http"
    "math/rand"
    "github.com/googollee/go-rest"
    "github.com/goamz/goamz/aws"
    "github.com/goamz/goamz/s3"
)

const _host = "0.0.0.0"
const _port = 8081
const _version = "0.0.1"
const _copyyear = "(C) 2017"

const _bucketName = "tacticiandocs"
const _publicKey = "AKIAJN2BF42Z3TE7LFKA"
const _privateKey = "eL5TnyYpD6A6fu81QfRKHrjB2jZS8qOpO+5DS3db"

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
    letterIdxBits = 6                    // 6 bits to represent a letter index
    letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
    letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

type fileSaveJson struct {
    Type string      `json:"type"`
    Data string      `json:"data"`
}

type fileGetJson struct {
    Location string  `json:"location"`
}

func parseFileGetJson(rawjson []byte) (location string, err error) {

    //ensure json is not empty
    if len(rawjson) == 0 {
        err = errors.New("Message is Empty")
        return
    }

    var fileObj fileGetJson

    err = json.Unmarshal(rawjson, &fileObj)
    //format error message
    if err != nil {
        err = errors.New("Invalid Message Format")
        return
    }

    location = fileObj.Location

    return
}

func parseFileSaveJson(rawjson []byte) (file_type string, data_b64 []byte, err error) {

    //ensure json is not empty
    if len(rawjson) == 0 {
        err = errors.New("Message is Empty")
        return
    }

    var fileObj fileSaveJson

    err = json.Unmarshal(rawjson, &fileObj)
    //format error message
    if err != nil {
        err = errors.New("Invalid Message Format")
        return
    }

    file_type = fileObj.Type
    data := fileObj.Data

    // decode base64 string into byte array
    data_b64, err = base64.StdEncoding.DecodeString(data)
    if err != nil {
        err = errors.New("Invalid File Contents")
    }

    return
}

func connectToS3Bucket(bucket_name string, aws_public string, aws_private string) (*s3.Bucket, error) {
    
    if len(bucket_name) == 0 {
        return nil, errors.New("Bucket Name Required")
    }
    
    if len(aws_public) == 0 || len(aws_private) == 0 {
        return nil, errors.New("Public and Private Key Required")
    }
    
    var auth = aws.Auth{
        AccessKey: aws_public,
        SecretKey: aws_private,
    }
    
    s3_client := s3.New(auth, aws.USEast)
    
    return s3_client.Bucket(bucket_name), nil
}

func writeToBucket(bucket *s3.Bucket, data []byte, file_name string, file_type string) error {

    if len(data) == 0 {
        err := errors.New("No Data to Save")
        return err
    }
    
    if len(file_name) == 0 {
        err := errors.New("File Name Required")
        return err
    }
    
    if len(file_type) == 0 {
        file_type = "text/plain"
    }
    
    //Options.SSE              bool
    //Options.Meta             map[string][]string
    //Options.ContentEncoding  string
    //Options.CacheControl     string
    //Options.RedirectLocation string
    //Options.ContentMD5       string
    
    var option s3.Options
    
    //Put(path string, data []byte, contType string, perm ACL, options Options) error {
    return bucket.Put(file_name, data, file_type, s3.PublicRead, option)
}

func genUniqueFileName(n int) string {
    b := make([]byte, n)
    // A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
    for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
        if remain == 0 {
            cache, remain = rand.Int63(), letterIdxMax
        }
        if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
            b[i] = letterBytes[idx]
            i--
        }
        cache >>= letterIdxBits
        remain--
    }

    return string(b)
}

func storeFileToS3(data []byte, file_type string) (location string, err error) {

    if len(data) == 0 {
        err = errors.New("No Data to Save")
        return
    }

    if len(file_type) == 0 {
        err = errors.New("File type required")
        return
    }

    // connect to bucket
    bucket, err := connectToS3Bucket(_bucketName, _publicKey, _privateKey)
    if err != nil {
        return
    }

    s3_fileName := genUniqueFileName(50)

    switch(file_type) {
    case "video/mp4":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"mp4")
    case "video/webm":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"webm")
    case "video/ogv":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"ogv")
    case "video/mp3":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"mp3")
    case "image/jpeg":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"jpg")
    case "image/png":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"png")
    case "text/plain":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"txt")
    case "application/pdf":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"pdf")
    case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"docx")
    case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"xlsx")
    case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
        s3_fileName = fmt.Sprintf("%s.%s",s3_fileName,"pptx")
    default:
        err = errors.New("Invalid file type")
        return
    }

    // save to bucket
    err = writeToBucket(bucket, data, s3_fileName, file_type)
    if err != nil {
        return
    }

    // set location identifier
    location = s3_fileName;

    return
}

func readFromS3(location string) ([]byte, error) {

    if len(location) == 0 {
        err := errors.New("No file location given")
        return nil, err
    }

    // connect to bucket
    bucket, err := connectToS3Bucket(_bucketName, _publicKey, _privateKey)
    if err != nil {
        return nil, err
    }

    return bucket.Get(location)
}

func docGetHandler(w http.ResponseWriter, r *http.Request) {

    //get body
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        //handle and log error
        http.Error(w, "Invalid Request Format", http.StatusBadRequest)
        return
    }
    
    location, err := parseFileGetJson(body)
    if err != nil {
        //handle and log error
        http.Error(w, "Invalid Request Body", http.StatusBadRequest)
        return
    }

    // read from S3
    data, err := readFromS3(location)
    if err != nil {
        //handle and log error
        http.Error(w, "File not found", http.StatusBadRequest)
        return
    }

    // base64 encode byte array
    encoded := base64.StdEncoding.EncodeToString(data)

    ret_json := "{"
    ret_json = ret_json + fmt.Sprintf("\"name\":\"%s\",",location)
    ret_json = ret_json + fmt.Sprintf("\"b64_string\":\"%s\"",encoded)
    ret_json = ret_json + "}"

    //set headers
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("content-type", "application/json")
    w.Header().Set("X-POWERED-BY", "Tactician Inc")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "x-api-key, origin, x-requested-with, content-type, accept, referer, user-agent")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
    w.Header().Set("Connection", "close")
    w.WriteHeader(200)
    
    //send response    
    w.Write([]byte(ret_json))
}

func docSaveHandler(w http.ResponseWriter, r *http.Request) {

    //get body
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        //handle and log error
        http.Error(w, "Invalid Request Format", http.StatusBadRequest)
        return
    }
    
    //try to parse json
    file_type, data, err := parseFileSaveJson(body)
    if err != nil {
        //handle and log error
        http.Error(w, "Invalid Request Body", http.StatusBadRequest)
        return
    }

    // store to S3
    // storeFileToS3(data []byte, file_type string) (location string, err error)
    location, err := storeFileToS3(data, file_type)
    if err != nil {
        fmt.Println(err)
        //handle and log error
        http.Error(w, "Unable to save file", http.StatusBadRequest)
        return
    }

    //set headers
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("content-type", "application/json")
    w.Header().Set("X-POWERED-BY", "Tactician Inc")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "x-api-key, origin, x-requested-with, content-type, accept, referer, user-agent")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
    w.Header().Set("Connection", "close")
    w.WriteHeader(200)
    
    //send response    
    w.Write([]byte(location))
}

func baseHandler(w http.ResponseWriter, r *http.Request) {

    body := fmt.Sprintf("<h1>Tactician</h1><p>Doc Service v%s. Copyright %s by Tactician Inc</p><hr><p>For more info see <a href=\"http://tactician.com\">http://tactician.com</a></p>", _version, _copyyear)
    msgcnt := fmt.Sprintf("<html><head><title>Tactician Status</title><meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\"></head><body>%s</body></html>", body)
    
    fmt.Fprintf(w, "%s", msgcnt)
}

func httpListener(host string, port int) (err error) {
    
    //safety check
    if len(host) == 0 || port <= 0 {
        err = errors.New("Host and Port Required in Configuration")
        return
    }
    
    //implment go-rest
    r := rest.New()
    
    // add router. router must before mime parser because parser need handler's parameter inserted by router.
    r.Use(rest.NewRouter())
    
    //display to console
    fmt.Println("-> Listening:", host, "on Port", port)

    //build address
    addr := fmt.Sprintf("%s:%d", host, port)
    
    //create get handlers
    r.Get("/", baseHandler)
    
    //save doc
    r.Post("/doc/save/", docSaveHandler)
    //get doc
    r.Post("/doc/get/", docGetHandler)

    //listen for connections
    http.ListenAndServe(addr, r)
    
    return
}

func main() {

    fmt.Println("================================================")
    fmt.Println("Doc Service", _version, _copyyear,"by Tactician Inc")
    fmt.Println("================================================")

    err := httpListener(_host, _port)
    if err != nil {
        //if error occurs exit
        fmt.Println(fmt.Sprintf("ERROR OnStart: %q [Exit]",err))
        os.Exit(1)
    }
}