package main

import "fmt"

// https://stackoverflow.com/questions/30226438

func nextPerm(p []int) {
	for i := len(p) - 1; i >= 0; i-- {
		if i == 0 || p[i] < len(p)-i-1 {
			p[i]++
			return
		}
		p[i] = 0
	}
}

func getPerm(orig, p []int) []int {
	result := append([]int{}, orig...)
	for i, v := range p {
		result[i], result[i+v] = result[i+v], result[i]
	}
	return result
}

func printAllPerms(orig []int) {
	for p := make([]int, len(orig)); p[0] < len(p); nextPerm(p) {
		fmt.Println(getPerm(orig, p))
	}
}
