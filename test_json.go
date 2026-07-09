package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Notification struct {
	CreatedAt time.Time `json:"CreatedAt" gorm:"autoCreateTime"`
}

func main() {
	n := Notification{CreatedAt: time.Now()}
	b, _ := json.Marshal(n)
	fmt.Println(string(b))
}
