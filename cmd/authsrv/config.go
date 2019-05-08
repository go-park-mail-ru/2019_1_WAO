package main

import (
	"time"
)

var (
	secret = []byte("secretkey")
	expires = 10 * time.Minute
)