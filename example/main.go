package example

import "fmt"

func main() {
	fmt.Println(Reverse("Hello, World!"))
}

func Reverse(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}
	return
}
