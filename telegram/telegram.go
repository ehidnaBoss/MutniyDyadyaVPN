package main

import (
	"log"
	"strconv"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("7549844817:AAEL_SfARov4mU2Gyhjdni3igNpwrHjRfEw")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	
	if message.Command() == "start" {
		showSubscriptionPlans(bot, chatID)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Подтверждаем получение callback
	bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	// Обрабатываем все кнопки независимо от состояния
	switch data {
	case "plan_1":
		showPaymentOptions(bot, chatID, "1 месяц")
	case "plan_3":
		showPaymentOptions(bot, chatID, "3 месяца")
	case "plan_6":
		showPaymentOptions(bot, chatID, "6 месяцев")
	case "pay":
		showPaymentButtons(bot, chatID)
	case "paid":
		sendKey(bot, chatID)
	case "help":
		sendHelp(bot, chatID)
	case "back_to_plans":
		showSubscriptionPlans(bot, chatID)
	case "back_to_payment":
		showPaymentButtons(bot, chatID)
	}
}

func showSubscriptionPlans(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Какую подписку хотите купить?\n1 3 6 месяцев"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1 месяц", "plan_1"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("3 месяца", "plan_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("6 месяцев", "plan_6"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showPaymentOptions(bot *tgbotapi.BotAPI, chatID int64, planText string) {
	text := "Оплата\nВы выбрали: " + planText
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Оплатить", "pay"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← Назад к тарифам", "back_to_plans"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showPaymentButtons(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Оплата\n\n\nПереведите деньги на карту:\n💳 1234 5678 9012 3456\n\nПосле оплаты нажмите кнопку ниже:"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Я оплатил!", "paid"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Помощь!", "help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← Назад к тарифам", "back_to_plans"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func sendKey(bot *tgbotapi.BotAPI, chatID int64) {
	// Генерируем ключ (в реальном проекте используйте криптографически стойкий генератор)
	key := "DEMO-KEY-" + strconv.FormatInt(chatID, 10)
	
	text := "Выдать ключ и написать инструкцию!\n\n" +
		   "🔑 Ваш ключ: `" + key + "`\n\n" +
		   "📋 Инструкция:\n" +
		   "1. Скачайте приложение\n" +
		   "2. Введите полученный ключ\n" +
		   "3. Наслаждайтесь подпиской!\n\n" +
		   "Ключ скопируется при нажатии на него."

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Купить еще", "back_to_plans"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func sendHelp(bot *tgbotapi.BotAPI, chatID int64) {
	text := "🆘 Помощь по оплате:\n\n" +
		   "❓ Проблемы с оплатой?\n" +
		   "📞 Контакты поддержки: @support_bot\n" +
		   "⏰ Время работы: 24/7\n\n" +
		   "💳 Реквизиты для оплаты:\n" +
		   "Карта: 1234 5678 9012 3456\n" +
		   "Получатель: Иван И.\n\n" +
		   "📋 Что делать после оплаты:\n" +
		   "1. Нажмите 'Я оплатил!'\n" +
		   "2. Получите ключ\n" +
		   "3. Следуйте инструкции"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← Назад к оплате", "back_to_payment"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← К тарифам", "back_to_plans"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}