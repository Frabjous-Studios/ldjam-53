package main

import (
	"fmt"
)

var lines = []string{
	"Please, you have to help me. I'm in dire need of that money!",
	"I can't believe this is happening. What am I supposed to do now?",
	"Please don't tell me you're refusing to give me my money back.",
	"I don't know what to do. I need that money to pay my bills.",
	"Why won't you help me? Don't you understand how much trouble I'm in?",
	"This is a nightmare. How am I going to survive without that money?",
	"Please, please, please...there has to be something you can do.",
	"I'm begging you, please don't take that money away from me.",
	"I can't even think straight right now. This is too much to handle.",
	"What am I supposed to do now? I have nothing left.",
	"I don't know what's going to happen to me if I don't get that money back.",
	"Please tell me you're joking. This can't be happening.",
	"I'm so scared. I don't know what to do.",
	"Please help me. I need that money to survive.",
	"I can't believe this is happening. How did things get so bad?",
	"I'm shaking right now. This is too much for me to handle.",
	"Why is this happening to me? What did I do to deserve this?",
	"I don't think I can make it through this. I need that money back.",
	"I feel like I'm drowning. Please help me before it's too late.",
	"I don't know who else to turn to. Please don't turn your back on me.",
}

func main() {
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
