# go-structconv

[![ci status](https://github.com/twihike/go-structconv/workflows/ci/badge.svg)](https://github.com/twihike/go-structconv/actions) [![license](https://img.shields.io/github/license/twihike/go-structconv)](LICENSE)

A converter between struct and other data.

## Installation

```shell
go get -u github.com/twihike/go-structconv
```

## Usage

`structconv.DecodeMap`

```go
package main

import (
    "fmt"

    "github.com/twihike/go-structconv/structconv"
)

type example1 struct {
    A int `map:"AA,required"`
    B []example2
}

type example2 struct {
    C int
    D string
    E string `map:"-"` // Omitted.
}

func main() {
    var e example1
    structconv.DecodeMap(map[string]interface{}{
        "AA": 1,
        "B": []map[string]interface{}{
            {
                "C": 2,
                "D": "foo",
                "E": "FOO",
            },
            {
                "C": 3,
                "D": "bar",
                "E": "BAR",
            },
        },
    }, &e, nil)
    fmt.Println(e) // {1 [{2 foo } {3 bar }]}
}
```

`structconv.DecodeStringMap`

```go
package main

import (
    "fmt"

    "github.com/twihike/go-structconv/structconv"
)

type config struct {
    AppName string `strmap:",required"`
    AppPort int
    DB      db
}

type db struct {
    Host     string `strmap:"DBHost,required"`
    Username int    `strmap:"DBUsername,required"`
    Password string `strmap:"-"` // Omitted.
}

func main() {
    m := map[string]string{
        "AppName":    "myapp",
        "AppPort":    "8080",
        "DBHost":     "mydb",
        "DBUsername": "1234",
        "DBPassword": "mypw",
    }
    var conf config
    conf.AppPort = 80 // Default value.
    err := structconv.DecodeStringMap(m, &conf, nil)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(conf) //{myapp 8080 {mydb 1234 }}
}
```

`structconv.DecodeEnv`

```shell
export APP_NAME=awesomeapp
export PORT=8080
export DB_HOST=mydb
export DB_USERNAME=1234
export DB_PASSWORD=mypw
```

```go
package main

import (
    "fmt"

    "github.com/twihike/go-structconv/structconv"
)

type config struct {
    AppName string `env:",required"`
    AppPort int    `env:"PORT"`
    DB      db
}

type db struct {
    Host     string `env:"DB_HOST,required"`
    Username int    `env:"DB_USERNAME,required"`
    Password string `env:"-"` // Omitted.
}

func main() {
    var conf config
    conf.AppPort = 80 // Default value.
    err := structconv.DecodeEnv(&conf, nil)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(conf) // {awesomeapp 8080 {mydb 1234 }}
}
```

## License

Copyright (c) 2020 twihike. All rights reserved.

This project is licensed under the terms of the MIT license.
