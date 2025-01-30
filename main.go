package main

import (
	"bufio"
	"fmt"
	"log"
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
	log.Println("Создан кольцевой буфер")
	return &RingBuffer{
		data: make([]int, size),
		size: size,
	}
}

func (rb *RingBuffer) Push(value int) bool {
	if rb.count == rb.size {
		log.Println("Буфер заполнен, невозможно добавить значение", value)
		return false
	}
	rb.data[rb.tail] = value
	rb.tail = (rb.tail + 1) % rb.size
	rb.count++
	log.Println("Добавлено в буфер:", value)
	return true
}

func (rb *RingBuffer) PopAll() []int {
	result := make([]int, 0, rb.count)
	for rb.count > 0 {
		result = append(result, rb.data[rb.head])
		rb.head = (rb.head + 1) % rb.size
		rb.count--
	}
	log.Println("Буфер очищен, отправлено значений:", len(result))
	return result
}

func filterNegatives(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for value := range input {
			if value >= 0 {
				log.Println("Пропущено положительное число:", value)
				output <- value
			} else {
				log.Println("Фильтр отрицательных отсеял:", value)
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
				log.Println("Пропущено число, кратное 3:", value)
				output <- value
			} else {
				log.Println("Фильтр кратных 3 отсеял:", value)
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
					log.Println("Входной канал закрыт, сбрасываем буфер")
					for _, bufferedValue := range ringBuffer.PopAll() {
						output <- bufferedValue
					}
					return
				}
				ringBuffer.Push(value)

			case <-ticker.C:
				log.Println("Таймер сработал, сбрасываем буфер")
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
				log.Println("Получена команда завершения")
				break
			}
			value, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("Некорректное значение, попробуйте снова.")
				log.Println("Ошибка преобразования строки в число:", input)
				continue
			}
			log.Println("Прочитано значение:", value)
			output <- value
		}
	}()
	return output
}

func dataSink(input <-chan int) {
	for value := range input {
		log.Println("Получены данные:", value)
		fmt.Printf("Получены данные: %d\n", value)
	}
}

func main() {
	log.Println("Запуск программы")
	input := dataSource()
	filteredNegatives := filterNegatives(input)
	filteredMultiples := filterNonMultiplesOfThree(filteredNegatives)
	buffered := bufferStage(filteredMultiples)
	dataSink(buffered)
	log.Println("Программа завершена")
}

