package handlers

import (
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "time"

    "github.com/Gazer/pocketfunctions/models"
    "github.com/gin-gonic/gin"
)

func (api *PocketAPI) ExecuteDocker() gin.HandlerFunc {
    // port will be from the API, statically assigned on create
    target := "localhost:%d"

    return func(c *gin.Context) {
        var path = c.Request.URL.Path[1:]

        fmt.Printf("Requested Name: %s\n", path)

        function, err := models.GetFunctionByName(api.Db, path)
        if err != nil {
            fmt.Printf("Error: %s\n", err.Error())
            c.String(http.StatusNotFound, err.Error())
            models.RegisterExecuted(api.Db, 0, 0, http.StatusNotFound)
            return
        }

        startTime := time.Now()
        director := func(req *http.Request) {
            // r := c.Request

            req.URL.Scheme = "http"
            req.URL.Host = fmt.Sprintf(target, 8080+function.Id)
            // req.Header["my-header"] = []string{r.Header.Get("my-header")}
            // Golang camelcases headers
            // delete(req.Header, "My-Header")
        }
        modifyHandler := func(response *http.Response) error {
            elapsed := time.Since(startTime)
            log.Printf("f(%d) executed in %d with code %d\n", function.Id, elapsed, response.StatusCode)
            if response.StatusCode == 200 {
                models.RegisterExecuted(api.Db, function.Id, elapsed, http.StatusOK)
            } else {
                models.RegisterExecuted(api.Db, function.Id, elapsed, http.StatusInternalServerError)
            }
            return nil
        }
        errorHandler := func(rw http.ResponseWriter, req *http.Request, err error) {
            elapsed := time.Since(startTime)
            models.RegisterExecuted(api.Db, function.Id, elapsed, http.StatusBadGateway)
            log.Printf("f(%d) executed in %d with code 502\n", function.Id, elapsed)
            rw.WriteHeader(http.StatusBadGateway)
        }
        proxy := &httputil.ReverseProxy{Director: director, ModifyResponse: modifyHandler, ErrorHandler: errorHandler}
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
