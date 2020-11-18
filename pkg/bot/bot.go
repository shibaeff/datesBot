package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"echoBot/pkg/bot/controllers"
	"echoBot/pkg/models"
	"echoBot/pkg/store"
)

const (
	waiting          = -1
	defaultBunchSize = 5
	noPhoto          = "none"

	registerCommand = "/register"
	nextCommand     = "/next"
	usersCommand    = "/users"
	helpCommand     = "/help"
	likeCommand     = "/like"
	matchesCommand  = "/matches"
	resetCommand    = "/reset"
	profileCommand  = "/profile"
	photoCommand    = "/photo"
	startCommand    = "/start"
	cancelCommand   = "/cancel"
	facultyCommand  = "/faculty"
	aboutCommand    = "/about"
	dumpCommand     = "/dump"
	notifyAll       = "/notify"

	greetMsg          = "Добро пожаловать в бота знакомств. Начните с /register."
	notUnderstood     = "Пожалуйста, выберите действие из меню"
	alreadyRegistered = "Вы уже зарегистрированы!"
	notRegistered     = "Вы не зарегистрированы!"
	notAdmin          = "Вы не админ"
)

var (
	registerButton = tgbotapi.KeyboardButton{Text: registerCommand}
	helpButton     = tgbotapi.KeyboardButton{Text: helpCommand}
	nextButton     = tgbotapi.KeyboardButton{Text: nextCommand}
	usersButton    = tgbotapi.KeyboardButton{Text: likeCommand}
	menuButtons    = []tgbotapi.KeyboardButton{registerButton, helpButton, nextButton, usersButton}
	menuKeyboard   = tgbotapi.NewReplyKeyboard(menuButtons)
)

type Bot interface {
	Reply(message *tgbotapi.Message) (interface{}, error)
}

type bot struct {
	store            store.Store
	api              *tgbotapi.BotAPI
	genderController controllers.Controller
	photoController  controllers.Controller
	logFile          *os.File
	adminsList       []string
}

// var Users = make(map[int64]bool)

func (b *bot) Reply(message *tgbotapi.Message) (reply interface{}, err error) {
	user, err := b.store.GetUser(message.Chat.ID)
	if err != nil {
		reply = replyWithText(greetMsg)
		err = b.store.PutUser(&models.User{
			Name:       message.Chat.FirstName,
			Faculty:    "",
			Gender:     "",
			WantGender: "",
			About:      "",
			Id:         message.Chat.ID,
			PhotoLink:  "",
			RegiStep:   waiting,
			UserName:   message.Chat.UserName,
		})
		return
	}
	if user.RegiStep != waiting && user.RegiStep < regOver {
		if message.Text == cancelCommand {
			if user.RegiStep != regPhoto {
				b.store.DeleteUser(user.Id)
				reply = replyWithText("Откат регистрации")
				return
			} else {
				b.store.UpdUserField(user.Id, "registep", regOver)
				reply = replyWithText("Откатываемся к старой информации")
				return
			}
		}
		reply = b.registerFlow(user, message)
		return
	}
	if message.Text[0] == '/' {
		split := strings.Split(message.Text, " ")
		switch split[0] {
		case notifyAll:
			if !b.ensureAdmin(user.UserName) {
				reply = replyWithText(notAdmin)
				return
			}
			reply, _ = b.notifyUsers(split[1])
			return
		case dumpCommand:
			if !b.ensureAdmin(user.UserName) {
				reply = replyWithText(notAdmin)
				return
			}
			offset, e := strconv.Atoi(split[1])
			if e != nil {
				reply = replyWithText("Неправильный оффсет")
				return
			}
			logs, e := b.grabLogs(offset)
			if e != nil {
				reply = replyWithText("Неправильный оффсет")
				return
			}
			reply = replyWithText(logs)
			return
		case aboutCommand:
			about := strings.Split(message.Text, " ")[1]
			err = b.store.UpdUserField(user.Id, "about", about)
			if err != nil {
				reply = replyWithText("Ошибка обновления!")
				return
			}
			reply = replyWithText(fmt.Sprintf("Обновили информацию на %s", about))
			return
		case facultyCommand:
			faculty := strings.Split(message.Text, " ")[1]
			err = b.store.UpdUserField(user.Id, "faculty", faculty)
			if err != nil {
				reply = replyWithText("Ошибка обновления!")
				return
			}
			reply = replyWithText(fmt.Sprintf("Обновили факультет на %s", faculty))
			return
		case startCommand:
			reply = replyWithText(greetMsg)
			return
		case helpCommand:
			reply = replyWithText(helpMsg)
			return
		case registerCommand:
			if user.RegiStep >= regOver {
				reply = replyWithText(alreadyRegistered)
				return
			}
			reply = b.registerFlow(user, message)
			return
		case nextCommand:
			if user.RegiStep < regOver {
				reply = replyWithText(notRegistered)
				return
			}
			newuser, e := b.store.GetAny(user.Id)
			if e != nil {
				reply = replyWithText("Не можем подобрать вариант")
				return
			}
			e = b.store.PutSeen(user.Id, newuser.Id)
			if e != nil {
				reply = replyWithText("Не можем подобрать вариант")
				return
			}
			reply = replyWithPhoto(newuser, message.Chat.ID)
			return
		case likeCommand:
			entry, e := b.store.GetSeen(user.Id)
			if e != nil {
				reply = replyWithText("failed to put your like")
				return
			}
			likee := entry.Whome[len(entry.Whome)-1]
			e = b.store.PutLike(user.Id, likee)
			if e != nil {
				reply = replyWithText("failed to put your like")
				return
			}
			reply = replyWithText("Успешный лайк!")
			likee_entry, e := b.store.GetLikes(likee)
			if e == nil {
				likee_likes := likee_entry.Whome
				_, ok1 := find(likee_likes, user.Id)
				if ok1 {
					user_entry, e := b.store.GetLikes(user.Id)
					if e != nil {
						return
					}
					_, ok1 = find(user_entry.Whome, likee)
					if ok1 {
						likee_user, e := b.store.GetUser(likee)
						if e != nil {
							reply = replyWithText("Такого пользователя уже нет")
						}
						reply = replyWithText(fmt.Sprintf(matchMsg, likee_user.UserName))
						e = b.store.GetMatchesRegistry().AddToList(user.Id, likee_user.Id)
						e = b.store.GetMatchesRegistry().AddToList(likee_user.Id, user.Id)
						return
					} else {
						return
					}
				} else {
					return
				}

			}
			return
		case usersCommand:
			if user.RegiStep < regOver {
				reply = replyWithText(notRegistered)
				return
			}
			usersString, err := b.listUsers()
			if err != nil {
				return nil, err
			}
			reply = replyWithText(usersString)
			return reply, nil
		case matchesCommand:
			matches, _ := b.prepareMatches(user.Id)
			reply = replyWithText(matches)
			return
		case resetCommand:
			b.store.DeleteFromRegistires(user.Id)
			reply = replyWithText("Ваши оценки сброшены!")
			return
		case profileCommand:
			reply = replyWithText(user.String())
			return
		case photoCommand:
			err = b.store.UpdUserField(user.Id, "photolink", noPhoto)
			if err != nil {
				reply = replyWithText("Ошибка обновления фото")
				return
			}
			err = b.store.UpdUserField(user.Id, "registep", regPhoto)
			if err != nil {
				reply = replyWithText("Ошибка обновления фото")
				return
			}
			reply = replyWithText("Ждем ваше фото!")
			return
		}
	}
	reply = replyWithText(notUnderstood)
	return
}

func (b *bot) listUsers() (str string, err error) {
	users, err := b.store.GetBunch(defaultBunchSize)
	if err != nil {
		log.Fatal(err)
		return
	}
	var raw []string
	for _, user := range users {
		raw = append(raw, user.String())
	}
	return strings.Join(raw, "\n"), nil
}

func NewBot(store store.Store, api *tgbotapi.BotAPI, logFile *os.File, admins []string) (b Bot) {
	b = &bot{
		store:            store,
		api:              api,
		genderController: &controllers.GenderController{},
		photoController:  &controllers.PhotoController{},
		logFile:          logFile,
		adminsList:       admins,
	}
	return b
}
