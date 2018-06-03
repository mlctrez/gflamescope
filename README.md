# gflamescope

[![Go Report Card](https://goreportcard.com/badge/github.com/mlctrez/gflamescope)](https://goreportcard.com/report/github.com/mlctrez/gflamescope)

This is a go powered version of [Netflix/flamescope](https://github.com/Netflix/flamescope)

* Summary

This project uses the ui components from Netflix/flamescope and replaces the python parts
with a version written in go.

* Building

See `build.sh` for how this is assembled. The binary is self contained in that you don't need to
have the static html/css/js files in a directory when starting the server. Just copy the built 
binary and execute it to get the flamescope server running on :8080.

This project is built using vgo. Replacing vgo with go in the build.sh should be enough if you don't
like that.

* Usage

```text
Usage of gflamescope:
  -root string
        examples root (default "examples")
```




