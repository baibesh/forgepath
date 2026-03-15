package bot

import (
	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/db"
)

func RequireUser(database *db.DB) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			user, err := database.GetUser(c.Sender().ID)
			if err != nil {
				return c.Send("Please /start first!")
			}
			c.Set("user", user)
			return next(c)
		}
	}
}

func UserFromContext(c tele.Context) *db.User {
	u, ok := c.Get("user").(*db.User)
	if !ok {
		return nil
	}
	return u
}
