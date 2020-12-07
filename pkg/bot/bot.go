package bot

import (
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"echoBot/pkg/bot/controllers"
	"echoBot/pkg/models"
	"echoBot/pkg/store"
	"echoBot/pkg/timelogger"
)

const (
	waiting          = -1
	defaultBunchSize = 5
	noPhoto          = "none"

	timeLoggingFileName = "time.csv"

	registerCommand   = "/register"
	nextCommand       = "/next"
	usersCommand      = "/users"
	helpCommand       = "/help"
	likeCommand       = "/like"
	matchesCommand    = "/matches"
	resetCommand      = "/reset"
	profileCommand    = "/profile"
	photoCommand      = "/photo"
	startCommand      = "/start"
	cancelCommand     = "/cancel"
	facultyCommand    = "/faculty"
	aboutCommand      = "/about"
	logCommand        = "/log"
	dumpCommand       = "/dump"
	notifyAll         = "/notify"
	reregisterCommand = "/reregister"
	feedbackCommand   = "/feedback"
	numbers           = "/numbers"
	deleteCommand     = "/delete"
	pauseCommand      = "/pause"

	greetMsg          = "Привет! ✨\nЭто бот знакомств МГУ. Работает аналогично Тиндеру 😉\n\nДля регистрации вызывай: /register, для отмены: /cancel. Бот запросит имя, фоточку и пару слов о себе.\n\nПредложения и баги пишите в /feedback."
	notUnderstood     = "Пожалуйста, выберите действие из меню"
	alreadyRegistered = "Вы уже зарегистрированы!"
	notRegistered     = "Вы не зарегистрированы!"
	notAdmin          = "Вы не админ"
	pleaseSendAgain   = "Пожалуйста, сделайте запрос еще раз"
)

var (
	profileButton = tgbotapi.KeyboardButton{Text: profileCommand}
	helpButton    = tgbotapi.KeyboardButton{Text: helpCommand}
	matchesButton = tgbotapi.KeyboardButton{Text: matchesCommand}
	nextButton    = tgbotapi.KeyboardButton{Text: nextCommand}
	menuButtons   = []tgbotapi.KeyboardButton{profileButton, helpButton, matchesButton, nextButton}
	menuKeyboard  = tgbotapi.NewReplyKeyboard(menuButtons)
)

type Bot interface {
	ReplyMessage(message *tgbotapi.Message) (interface{}, error)
	HandleCallbackQuery(query *tgbotapi.CallbackQuery) (interface{}, error)
	GetStore() store.Store
}

type bot struct {
	store            store.Store
	api              *tgbotapi.BotAPI
	genderController controllers.Controller
	photoController  controllers.Controller
	aboutController  controllers.Controller
	logFile          *os.File
	timeloggers      map[string]timelogger.TimeLogger
	adminsList       []string
	actionsLog       *log.Logger
}

func (b *bot) GetStore() store.Store {
	return b.store
}

func (b *bot) HandleCallbackQuery(query *tgbotapi.CallbackQuery) (reply interface{}, err error) {
	_, err = b.store.GetUser(int64(query.From.ID))
	return nil, nil
}

func (b *bot) ReplyMessage(message *tgbotapi.Message) (reply interface{}, err error) {
	// check if user is registered
	// unregistered users are allowed only to do /start, /help, /register
	_, err = b.store.GetUser(message.Chat.ID)
	// Putting user in the db
	if err != nil {
		u := models.User{
			Name:       message.Chat.FirstName,
			Faculty:    "",
			Gender:     "",
			WantGender: "",
			About:      "",
			Id:         message.Chat.ID,
			PhotoLink:  "",
			RegiStep:   waiting,
			UserName:   message.Chat.UserName,
		}
		b.store.PutUser(u)
	}
	text := message.Text
	if text[0] == '/' {
		split := strings.Split(text, " ")
		// in case of paired commands
		if len(split) == 1 {
			switch text {
			case startCommand:
				return replyWithText(greetMsg), nil
			case helpCommand:
				return replyWithText("Начинаем регистарцию"), nil
			}
		}
		if len(split) == 2 {

		}
		reply = replyWithText("Неизвестная команда")
		return
	}
	return replyWithText(notUnderstood), nil
}

func (b *bot) setTimeLoggers() {
	b.timeloggers = make(map[string]timelogger.TimeLogger)
	b.timeloggers[startCommand] = timelogger.NewTimeLogger(startCommand, timeLoggingFileName)
}

func (b *bot) setActionLoggers() {
	file, err := os.OpenFile("actions.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Cannot create or open log file")
	}
	b.actionsLog = log.New(file, "Common Logger:\t", log.Ldate|log.Ltime|log.Lshortfile)
}

func NewBot(store store.Store, api *tgbotapi.BotAPI, logFile *os.File, admins []string) (b Bot) {
	b = &bot{
		store:            store,
		api:              api,
		genderController: &controllers.GenderController{},
		photoController:  &controllers.PhotoController{},
		aboutController:  &controllers.AboutController{},
		logFile:          logFile,
		adminsList:       admins,
	}
	b.(*bot).setTimeLoggers()
	b.(*bot).setActionLoggers()
	return b
}
