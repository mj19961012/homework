package main

import (
	"fmt"
	"time"
)

func consumer(ch <-chan int) {
	for value := range ch {
		fmt.Println("consumer:%d", value)
	}
}

func produce(ch chan<- int) {
	for i := 0; i < 10; i++ {
		ch <- i
		time.Sleep(1 * time.Second)
		fmt.Println("produce:%d", i)
	}
	close(ch)
}
func main() {
	array := [5]string{"I", "am", "stupid", "and", "weak"}
	for i := 0; i < len(array); i++ {
		fmt.Println(array[i])
	}
	array[2] = "smart"
	array[4] = "strong"

	for i := 0; i < len(array); i++ {
		fmt.Println(array[i])
	}

	ch := make(chan int, 10)
	go produce(ch)
	consumer(ch)
}
