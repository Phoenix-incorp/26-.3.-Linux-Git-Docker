package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	bufferSize    = 10
	flushInterval = 5 * time.Second
)

type RingBuffer struct {
	data  []int
	size  int
	head  int
	tail  int
	count int
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]int, size),
		size: size,
	}
}

func (rb *RingBuffer) Push(value int) bool {
	if rb.count == rb.size {
		return false
	}
	rb.data[rb.tail] = value
	rb.tail = (rb.tail + 1) % rb.size
	rb.count++
	return true
}

func (rb *RingBuffer) PopAll() []int {
	result := make([]int, 0, rb.count)
	for rb.count > 0 {
		result = append(result, rb.data[rb.head])
		rb.head = (rb.head + 1) % rb.size
		rb.count--
	}
	return result
}

func filterNegatives(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for value := range input {
			if value >= 0 {
				output <- value
			}
		}
	}()
	return output
}

func filterNonMultiplesOfThree(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for value := range input {
			if value != 0 && value%3 == 0 {
				output <- value
			}
		}
	}()
	return output
}

func bufferStage(input <-chan int) <-chan int {
	output := make(chan int)
	ringBuffer := NewRingBuffer(bufferSize)

	go func() {
		defer close(output)
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		for {
			select {
			case value, ok := <-input:
				if !ok {
					for _, bufferedValue := range ringBuffer.PopAll() {
						output <- bufferedValue
					}
					return
				}
				ringBuffer.Push(value)

			case <-ticker.C:
				for _, bufferedValue := range ringBuffer.PopAll() {
					output <- bufferedValue
				}
			}
		}
	}()
	return output
}

func dataSource() <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Введите целые числа. Для завершения введите 'exit'.")
		for scanner.Scan() {
			input := scanner.Text()
			if input == "exit" {
				break
			}
			value, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("Некорректное значение, попробуйте снова.")
				continue
			}
			output <- value
		}
	}()
	return output
}

func dataSink(input <-chan int) {
	for value := range input {
		fmt.Printf("Получены данные: %d\n", value)
	}
}

func main() {
	input := dataSource()
	filteredNegatives := filterNegatives(input)
	filteredMultiples := filterNonMultiplesOfThree(filteredNegatives)
	buffered := bufferStage(filteredMultiples)
	dataSink(buffered)
}
