// Copyright (c) 2020 twihike. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package structconv

import (
	"fmt"
	"os"
)

func ExampleDecodeMap() {
	type point struct {
		X int `map:"x"`
		Y int `map:"y"`
	}
	type config struct {
		AppName string `map:",required"`
		Port    int
		Addr    string `map:"-"` // omitted
		Debug   bool
		Points  []point
	}
	strMap := map[string]interface{}{
		"AppName": "myapp",
		"Port":    8080,
		"Addr":    ":8080",
		"Debug":   true,
		"Points": []map[string]int{
			{"x": 1, "y": 1},
			{"x": 2, "y": 2},
		},
	}

	var conf config
	conf.Port = 80 // Default value.
	err := DecodeMap(strMap, &conf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", conf)
	// Output:
	// {AppName:myapp Port:8080 Addr: Debug:true Points:[{X:1 Y:1} {X:2 Y:2}]}
}

func ExampleDecodeStringMap() {
	type db struct {
		Host string `strmap:"DBHost"`
		Port int    `strmap:"DBPort"`
	}
	type config struct {
		AppName string `strmap:",required"`
		Port    int    `strmap:"port"`
		Addr    string `strmap:"-"` // omitted
		Debug   bool
		DB      db
	}
	strMap := map[string]string{
		"AppName": "myapp",
		"port":    "8080",
		"Addr":    ":8080",
		"Debug":   "true",
		"DBHost":  "mydb",
		"DBPort":  "1234",
	}

	var conf config
	conf.Port = 80 // Default value.
	err := DecodeStringMap(strMap, &conf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", conf)
	// Output:
	// {AppName:myapp Port:8080 Addr: Debug:true DB:{Host:mydb Port:1234}}
}

func ExampleDecodeEnv() {
	type db struct {
		Host string `env:"DB_HOST"`
		Port int    `env:"DB_PORT"`
	}
	type config struct {
		AppName string `env:",required"`
		Port    int
		Addr    string `env:"-"` // omitted
		Debug   bool
		DB      db
	}
	envData := map[string]string{
		"APP_NAME": "myapp",
		"PORT":     "8080",
		"ADDR":     ":8080",
		"DEBUG":    "true",
		"DB_HOST":  "mydb",
		"DB_PORT":  "1234",
	}

	os.Clearenv()
	for k, v := range envData {
		os.Setenv(k, v)
	}

	var env config
	env.Port = 80 // Default value.
	err := DecodeEnv(&env)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", env)
	// Output:
	// {AppName:myapp Port:8080 Addr: Debug:true DB:{Host:mydb Port:1234}}
}
