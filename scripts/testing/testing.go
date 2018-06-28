package testing

import (
	"rboot"
	"time"
)

func parse(bot *rboot.Robot) error {
	switch bot.Matcher {
	case `123`:
		bot.Send(`1 or 2 or 3`)

	case `abc`:
		bot.Send(`a or b or c`)

	}

	return nil
}

func hook(bot rboot.Robot) {
	bot.Timer(10 * time.Second)
	bot.Handle(`/timer/10s`, func(evt rboot.Event) {
		//data := evt.Data.(rboot.TimerData)

		bot.Send(`10s`)
	})
}

func init() {
	rboot.RegisterScript(`testing`, &rboot.Script{
		Action: parse,
		Ruleset: map[string]string{
			`123`: `1|2|3`,
			`abc`: `a|b|c`,
		},
		Hook: hook,
	})
}
