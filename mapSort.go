package main

import (
	"fmt"
	"sort"
)

// C ...
type C map[int]int64

func (s C) Len() int           { return len(s) }
func (s C) Swap(i, j i
	
	)      { s[i], s[j] = s[j], s[i] }
func (s C) Less(i, j int) bool { return s[i] < s[j] }

func main() {
	var c = C{4: 6, 3: 9, 6: 1, 2: 10}

	sort.Sort(c)
	// for k, v := range c {
	// 	if v == 0 {
	// 		fmt.Println("delete v: ", v)
	// 		delete(c, k)
	// 	}
	// }

	fmt.Println(c)

}
