// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello, 世界")
	timeout := time.Now()
	fmt.Println(timeout)
	time.Sleep(2 * time.Second)
	cur := time.Now().Unix()
	fmt.Println(cur)
}
