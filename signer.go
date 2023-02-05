package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, j := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go jobContainer(j, in, out, wg)
		in = out
	}
	wg.Wait()
}

func jobContainer(j job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	j(in, out)
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for data := range in {
		wg.Add(1)
		go func(data interface{}) {
			defer wg.Done()
			str := strconv.Itoa(data.(int))

			mu.Lock()
			md5 := DataSignerMd5(str)
			mu.Unlock()

			chan1 := make(chan string)
			go func(out chan string, str string) {
				out <- DataSignerCrc32(str)
			}(chan1, str)

			chan2 := make(chan string)
			go func(out chan string, md5 string) {
				out <- DataSignerCrc32(md5)
			}(chan2, md5)

			out <- (<-chan1 + "~" + <-chan2)
		}(data)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		go func(data interface{}) {
			defer wg.Done()
			waitG := &sync.WaitGroup{}

			array := make([]string, 6)

			for th := 0; th < 6; th++ {
				waitG.Add(1)
				str := data.(string)

				go func(str string, th int) {
					defer waitG.Done()
					array[th] = DataSignerCrc32(strconv.Itoa(th) + str)
				}(str, th)
			}
			waitG.Wait()
			out <- strings.Join(array, "")
		}(data)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var strs []string
	for data := range in {
		strs = append(strs, data.(string))
	}
	sort.Strings(strs)
	out <- strings.Join(strs, "_")
}
