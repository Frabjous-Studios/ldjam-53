package main

import (
	"fmt"
	"strings"
)

var input = `"Good day! I'm here to cash this check and buy myself a nice cup of coffee."
"Hi! I'm so excited to deposit this check, I feel like I won the lottery."
"Hello! I'm just stopping by to cash my paycheck and treat myself to some ice cream."
"Hello! I'm here to cash this check and hopefully not spend it all in one place."
"Hello! I'm just stopping by to cash this check and buy some treats for my furry friend."
"Hello! I'm just stopping by to cash this check and maybe treat myself to a nice dinner."
"Good afternoon! I'm here to cash this check and hopefully start a new hobby."
"Hey, can you help me cash this check? I promise to be in a good mood all day."
"Hello! I'm just stopping by to cash this check and treat myself to a little shopping."
"Good afternoon! I'm here to cash this check and hopefully put a smile on someone's face."
"Hey, can you help me cash this check? I promise to share the good vibes."
"Hello! I'm just stopping by to cash this check and treat myself to a little self-care`

func main() {
	lines := strings.Split(input, "\n")
	fmt.Println("<< set $d to 0 >>")
	fmt.Printf("<< set $d to dice(%d) >>\n", len(lines))
	fmt.Println("<< if $d == 1 >>")
	for idx, l := range lines {
		fmt.Printf("\t%s\n", l)
		if idx != len(lines)-1 {
			fmt.Printf("<< elseif $d == %d >>\n", idx+2)
		}
	}
	fmt.Println("<<endif>>")
}
