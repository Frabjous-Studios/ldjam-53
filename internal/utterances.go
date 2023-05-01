package internal

import "math/rand"

var HandsTrash = []string{
	"Oh, sorry. Could you throw that away for me?",
	"Oh, is this the latest currency trend? Garbage is the new Crypto Nyan-coin?",
	"I didn't know I could deposit my trash here, can I also get a compost account?",
	"I see the bank is really cutting costs, I guess I'll have to use this as toilet paper.",
	"Wow, talk about a trashy bank...literally.",
	"This is great! Now I can finally pay off my debts with something that's worth even less than money.",
	"I appreciate the thought, but I already have enough garbage in my life.",
	"Thanks, now I have something to feed to my pet raccoon.",
	"I thought the bank was supposed to help me clean up my finances, not add to the mess.",
	"I'm pretty sure I can't use this to buy my morning coffee, but I'll try anyway.",
	"This must be some sort of new recycling initiative...thanks for the fancy paperweight!",
	"I think I'll frame this and hang it on my wall. It'll be a constant reminder of my financial decisions.",
	"Well, I guess I should be grateful. At least this garbage won't depreciate in value like money does.",
	"Thanks for the gift, but I prefer my garbage in a more eco-friendly wrapping.",
}

var WrongSlip = []string{
	"What? You mean I have to get back in line?!",
	"This isn't even the right slip?",
	"Wow, I didn't know you guys accepted Monopoly money.",
	"Wait, this isn't even my check. How did you manage to mess this up?",
	"Oh; it's wrong? Do you even know how to read?!",
	"This is clearly not the right slip. I'm sorry, but this is actually the wrong slip.",
	"Could you please give me the correct one?",
	"Uh... Do you need a calculator, or...?",
	"Looks like someone needs to go back to basic math.",
	"Did you take a course on how to make mistakes? Uh, this isn't even my name.",
	"You need to double-check your work, buddy.",
	"Did you even read this before giving it back to me?",
	"I don't think this is what I wrote on the slip. I guess I'll have to try again tomorrow.",
}

var BossDismissal = []string{
	"What?! You think you can dismiss me?!!? I wasn't through talking!",
	"What do you think you're doing? You can't just dismiss me like that!",
	"Oh, I see you're trying to get fired. Good luck with that.",
	"Wait, did you just... dismiss me? I'm your boss, you know.",
	"Are you kidding me? You can't just ignore me and expect everything to be okay.",
	"What, do you think that thing is my 'off' button?",
	"Ugh... Next time, try using your words instead of the bell.",
	"Is this your way of asking for a break?",
	"I think someone needs a reminder of who's in charge here.",
	"Are we playing a game of 'how to get fired' now?",
	"I'm glad we're clear on who's in control here...",
	"Sorry, I don't think I understand 'dismiss the boss' day.'",
}

var CashBackDeposit = []string{
	"I'd like to deposit that, actually.",
	"Excuse me, but I needed that money for my deposit. Can you please be more careful?",
	"Oh sure, just give away my money to anyone who asks. Thanks a lot!",
	"Wait, you're actually giving it back? I was expecting to have to argue with you for it.",
	"Are you kidding me? Do you even know how banking works?",
	"I'm sorry, but I actually need that money to make my deposit. Could you please give it back to me?",
	"Oh, you're giving away free money now?",
	"I didn't know we were running a charity here.",
	"Thanks for the tip, but I prefer my cash in the bank.",
	"Well, I guess I won't be depositing anything today.",
	"Do you need a calculator to count?",
	"You know I wanted to deposit this, right?",
	"Can I have my money back now?",
	"Looks like I'll be holding onto my cash a little longer.",
}

var FreeMoney = []string{
	"Wow! Cool! Hope I don't get robbed on the way home!",
	"Whoa, I didn't know you were that generous.",
	"Heyyy! Best bank EVER!",
	"Do you even know how to count? This is way more than I asked for.",
	"Wow, looks like we're having a clearance sale!",
	"Are you trying to bribe me or something?",
	"You just made my day, thank you!",
	"Jackpot!",
	"I won the lottery!",
	"Wow! I'm keeping it. Goodbye!",
	"Well, I wasn't expecting this today.",
}

func randSlice[T any](ts []T) T {
	return ts[rand.Intn(len(ts))]
}
