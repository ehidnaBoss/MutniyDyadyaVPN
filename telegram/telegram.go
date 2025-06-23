package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"mutinydydayvpn/internals/payments"
	"mutinydydayvpn/internals/user"
)

type Bot struct {
	api            *tgbotapi.BotAPI
	paymentService *payments.Service
	userService    *user.Service
}

func NewBot(token string, paymentService *payments.Service, userService *user.Service) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &Bot{
		api:            bot,
		paymentService: paymentService,
		userService:    userService,
	}
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}
	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	ctx := context.Background()
	chatID := message.Chat.ID
	
	// Регистрируем пользователя если его нет
	b.userService.CreateOrUpdateUser(ctx, &user.CreateUserRequest{
		ChatID:    chatID,
		Username:  message.From.UserName,
		FirstName: message.From.FirstName,
		LastName:  message.From.LastName,
	})
	
	if message.Command() == "start" {
		b.showSubscriptionPlans(chatID)
	}
}

func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	ctx := context.Background()
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Подтверждаем получение callback
	b.api.Request(tgbotapi.NewCallback(callback.ID, ""))

	switch {
	case data == "plan_1":
		b.showPaymentOptions(chatID, "1 месяц", 299.0, 1)
	case data == "plan_3":
		b.showPaymentOptions(chatID, "3 месяца", 699.0, 3)
	case data == "plan_6":
		b.showPaymentOptions(chatID, "6 месяцев", 1199.0, 6)
	case strings.HasPrefix(data, "pay_"):
		// Извлекаем план из callback data
		parts := strings.Split(data, "_")
		if len(parts) >= 2 {
			planMonths, _ := strconv.Atoi(parts[1])
			b.createPayment(ctx, chatID, planMonths)
		}
	case data == "help":
		b.sendHelp(chatID)
	case data == "back_to_plans":
		b.showSubscriptionPlans(chatID)
	}
}

func (b *Bot) showSubscriptionPlans(chatID int64) {
	text := `🚀 Выберите подписку на VPN:

🔹 1 месяц - 299₽
🔹 3 месяца - 699₽ (экономия 198₽)
🔹 6 месяцев - 1199₽ (экономия 595₽)

✅ Безлимитный трафик
✅ Высокая скорость
✅ Серверы по всему миру`
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1 месяц - 200₽", "plan_1"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("3 месяца - 550₽", "plan_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("6 месяцев - 1050₽", "plan_6"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) showPaymentOptions(chatID int64, planText string, amount float64, months int) {
	text := fmt.Sprintf(`💳 Оплата подписки

📦 Тариф: %s
💰 Стоимость: %.0f₽

Нажмите кнопку ниже для перехода к оплате:`, planText, amount)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Перейти к оплате", fmt.Sprintf("pay_%d", months)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← Назад к тарифам", "back_to_plans"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) createPayment(ctx context.Context, chatID int64, months int) {
	var amount float64
	switch months {
	case 1:
		amount = 200.0
	case 3:
		amount = 550.0
	case 6:
		amount = 1050.0
	default:
		amount = 200.0
	}

	// Создаем заявку на оплату
	req := payments.CreatePaymentRequest{
		UserChatID:     chatID,
		SubscriptionID: months, // временно используем как ID плана
		Amount:         amount,
	}

	payment, err := b.paymentService.CreatePayment(ctx, req)
	if err != nil {
		b.sendError(chatID, "Ошибка создания платежа. Попробуйте позже.")
		return
	}

	// Отправляем ссылку на оплату
	text := fmt.Sprintf(`💳 Платеж создан!

💰 Сумма: %.0f₽
🆔 Номер платежа: %d

Для оплаты используйте CloudPayments или переведите на карту:
💳 2200 7007 1234 5678

После оплаты ключ будет выдан автоматически.`, payment.Amount, payment.ID)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← К тарифам", "back_to_plans"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) sendError(chatID int64, errorText string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+errorText)
	b.api.Send(msg)
}

func (b *Bot) sendHelp(chatID int64) {
	text := `🆘 Помощь по оплате:

💳 Способы оплаты:
• Банковские карты (Visa, MasterCard, МИР)
• Переводы на карту

📞 Поддержка: @support_username
⏰ Время работы: 24/7

💡 Часто задаваемые вопросы:
• Ключ приходит автоматически после оплаты
• Подписка активируется сразу
• Возврат средств возможен в течение 24 часов`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← К тарифам", "back_to_plans"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// SendKey отправляет ключ пользователю после успешной оплаты
func (b *Bot) SendKey(chatID int64, key string, planMonths int) {
	text := fmt.Sprintf(`🎉 Оплата прошла успешно!

🔑 Ваш ключ: %s

📱 Инструкция по подключению:
1. Скачайте приложение Outline
2. Нажмите "+" и вставьте ключ
3. Подключайтесь к VPN!

⏰ Подписка действует: %d месяц(ев)
💡 Ключ скопируется при нажатии на него`, key, planMonths)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛒 Купить еще", "back_to_plans"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}