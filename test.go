package main

import "fmt"

type Test struct {
	First []Second
}

type Second struct {
	Xxx []int
}

func main() {

	s1 := Second{Xxx: []int{1, 2, 3, 4, 5}}
	s2 := Second{Xxx: []int{}}
	s3 := Second{Xxx: []int{}}
	s4 := Second{Xxx: []int{31, 32, 33, 34, 35}}

	t := Test{First: []Second{s1, s2, s3, s4}}
	fmt.Println(t)
	// t.First = t.First[:1]
	// fmt.Println(t)

	for _, j := range t.First {
		fmt.Println(j)
		for _, k := range j.Xxx {
			if k == nil {

			}
		}
	}

}
