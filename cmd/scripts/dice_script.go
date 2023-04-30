package main

import (
	"fmt"
	"strings"
)

var input = `Whatever, bye.
I can't believe how long this took, you need to work on your efficiency. Bye.
Thanks for nothing. Bye.
I hope you do better next time. Bye.
This is ridiculous, I'll be finding another bank. Bye.
I can't believe I wasted my time here. Bye.
You really need to work on your customer service skills. Bye.
I hope you get your act together. Bye.
I don't have time for this nonsense. Bye.
You really need to speed things up. Bye.
I'm not impressed, bye.
I can't believe how incompetent you are. Bye.
This is unacceptable, bye.
I don't have patience for this kind of service. Bye.
I'm going to let your supervisor know how terrible this was. Bye.
I'm so disappointed in this experience. Bye.
I hope you take some customer service classes. Bye.
I can't believe how unprofessional this was. Bye.
You've just lost a customer. Bye.
This was a waste of my time. Bye.
I hope the next customer has a better experience than me. Bye.
I'm done with this bank. Bye.
I'm glad to be done dealing with you. Bye.
You really need to learn how to do your job. Bye.
I don't have any patience left for this kind of service. Bye.
I'm not coming back here, bye.
You've just lost my trust. Bye.
I can't believe how bad this was. Bye.
I'll be taking my business elsewhere. Bye.
You need to step up your game. Bye.
I hope you do better in the future. Bye.
I don't have any faith in this bank. Bye.
This is the worst service I've ever received. Bye.
I hope this experience was a learning lesson for you. Bye.
I'm going to tell everyone I know about this terrible service. Bye.`

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
