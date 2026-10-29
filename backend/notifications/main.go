package main

import (
	"fmt"
	"sync"
	"time"
)

func firstUniqueSymbol(str string) int {
	// str = "loveleetcode"
	m := make(map[rune]int)
	for _, v := range str {
		m[v]++
	}

	for i, v := range str {
		if m[v] == 1 {
			return i
		}
	}

	return -1
}

type ServerStat struct {
	Server    int
	Stability float64
}

func groupByStability(stats []ServerStat) map[string][]int {
	result := make(map[string][]int)
	for _, stat := range stats {
		val := fmt.Sprintf("%v", stat.Stability)
		if v, ok := result[val]; ok {
			result[val] = append(v, stat.Server)
			continue
		}
		result[val] = []int{stat.Server}
	}

	return result
}

func HasSum(nums []int, target int) bool {
	m := make(map[int]struct{})
	for _, num := range nums {
		diff := target - num
		if _, ok := m[diff]; ok {
			return true
		}
		m[num] = struct{}{}
	}

	return false
}

func isValid(str string) bool {
	m := map[rune]rune{
		'}': '{',
		']': '[',
		')': '(',
	}
	slc := make([]rune, 0, len(str))

	for _, c := range str {
		if v, ok := m[c]; ok {
			if slc[len(slc)-1] != v {
				return false
			}
			slc = slc[:len(slc)-1]
		} else {
			slc = append(slc, c)
		}
	}
	println(string(slc))

	return len(slc) == 0
}

func main() {
	test := firstUniqueSymbol("loveleetcode")
	res := isValid("[()]()")
	fmt.Println("res - ", res, test)
}

//type Node struct {
//	value int
//	next  *Node
//}

//func (n *Node) Next() Node {
//	return *n.next
//}
//
//func (n *Node) removeFromEnd(head *Node, del int) *Node {
//	length := 0
//	current := head
//	for current != nil {
//		length++
//		current = current.next
//	}
//
//	if length == del {
//		return head.next
//	}
//
//	current = head
//	for i := 0; i < length-del; i++ {
//		current = current.next
//	}
//
//	current.next = current.next.next
//
//	return n
//}
//
//func reverseWord(s string) string {
//	words := strings.Split(s, " ")
//	prev, next := "", ""
//
//	return ""
//}

func isPalindrome(str []rune) bool {
	prev, next := 0, len(str)-1
	for prev < next {
		if str[prev] != str[next] {
			return false
		}
		prev++
		next--
	}
	return true
}

func exec() string {
	time.Sleep(2 * time.Second)
	return "ok"
}

func main_____() {
	sm := make(chan struct{}, 3)
	for i := 0; i < 10; i++ {
		sm <- struct{}{}
		go func() {
			exec()
			<-sm
		}()
		fmt.Printf("Work %v is done\n", i)
	}
	close(sm)

	println("All works done")
}

type Transaction struct {
	ID     int64
	Amount int64
}

func readTransaction(count int) <-chan Transaction {
	out := make(chan Transaction)
	go func() {
		for i := range count {
			out <- Transaction{int64(i), int64(i + 1*10)}
		}
		close(out)
	}()
	return out
}
func filerTransaction(in <-chan Transaction) <-chan Transaction {
	out := make(chan Transaction)

	go func() {
		for tr := range in {
			if tr.Amount > 0 {
				out <- tr
			}
		}

		close(out)
	}()

	return out
}
func convertValute(in <-chan Transaction) <-chan Transaction {
	out := make(chan Transaction)

	go func() {
		for tr := range in {
			tr.Amount += tr.Amount
			out <- tr
		}

		close(out)
	}()

	return out
}
func save(in <-chan Transaction) {
	for tr := range in {
		fmt.Printf("Saving transaction %d - %v\n", tr.ID, tr.Amount)
	}
}

func main__() {
	transactions := readTransaction(10)
	filtred := filerTransaction(transactions)
	convert := convertValute(filtred)
	save(convert)
	fmt.Println("Saving transactions done")
}

type Task struct {
	ID       int
	Filename string
}

func ProcessImage(task Task) string {
	time.Sleep(3 * time.Second)
	return fmt.Sprintf("Work is %s done by id = %d", task.Filename, task.ID)
}

func ExecuteTask(taskCh <-chan Task, result chan string) {
	for task := range taskCh {
		fmt.Printf("Worker START - Task ID: %d\tFilename: %s\n", task.ID, task.Filename)
		result <- ProcessImage(task)
	}
}

func ___main() {
	const (
		numWorkers = 4
		numTasks   = 10
	)

	taskCh := make(chan Task, numTasks)
	resultCh := make(chan string)
	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ExecuteTask(taskCh, resultCh)
		}()
	}

	go func() {
		for i := 0; i < numTasks; i++ {
			taskCh <- Task{
				ID:       i,
				Filename: fmt.Sprintf("task_%d.txt", i),
			}
		}
		close(taskCh)
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for res := range resultCh {
		fmt.Println(res)
	}

	fmt.Println("DONE")
}

// ///

type Cache interface {
	set(k, v string)
	get(k string) (string, bool)
}

type MyCache struct {
	mu sync.RWMutex
	m  map[string]string
}

func NewMyCache() Cache {
	return &MyCache{m: make(map[string]string)}
}

func (m *MyCache) set(k, v string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[k] = v
}

func (m *MyCache) get(k string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[k]
	return v, ok
}

func __main() {
	myCache := NewMyCache()
	myCache.set("1", "1")
	myCache.set("2", "2")
	println(myCache.get("1"))
	println(myCache.get("3"))
}

// ///
func merge(channels ...chan int) chan int {
	resChan := make(chan int)
	wg := &sync.WaitGroup{}
	wg.Add(len(channels))
	for _, ch := range channels {
		go func() {
			defer wg.Done()
			for value := range ch {
				resChan <- value
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	return resChan
}

func _main() {
	channels := make([]chan int, 10)
	for i := range channels {
		channels[i] = make(chan int)
	}

	for i := range channels {
		go func() {
			channels[i] <- i + 1
			close(channels[i])
		}()
	}

	for v := range merge(channels...) {
		println(v)
	}
}
