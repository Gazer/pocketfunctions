# pocketfunctions

Pocketfunctions is a small serverless functions service

---
>Made with ❤️ by Ricardo Markiewicz // [@gazeria](https://twitter.com/gazeria).

## What is this?

This is a service that allow you to deploy functions as a service.

This documentation is still work-in-progress :)


## Dependencies

* Go 1.23
* Docker

## Run the server

```bash
$> cd server
$ server> go run main.go
```

Some limitations:

* docker cli is needed for now, the server will run docker commands directly
* We did not implemented auth yet, so do not run this in public servers
* Por 8080 is fixed, sorry, will be fixed later

## Add functions

A function needs a runtime to be executed, for now, we only have `dart`.

Check [https://github.com/Gazer/exp-pocketfunct-dart](https://github.com/Gazer/exp-pocketfunct-dart) for more information.

## Admin

We have a WorkInProgress admin panel to see some values, just open 
[http://localhost:8080/_/](http://localhost:8080/_/) in your browser.

