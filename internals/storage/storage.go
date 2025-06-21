package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"mutinydydayvpn/internals/domain"
)

func initDB() *sql.DB{
	db,err := sql.Open("sqlite3", "./bot.db")
	if err != nil{
		log.Fatal("Ошибка подключения к БД:", err)
	}
	
	// Проверяем соединение
	if err = db.Ping(); err != nil {
		log.Fatal("Ошибка ping БД:", err)
	}
	createTables(db)
	
	log.Println("База данных инициализирована")
	return db

}

func createTables(db *sql.DB) {
	// Таблица пользователей
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER UNIQUE NOT NULL,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Таблица подписок
	createSubscriptionsTable := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_chat_id INTEGER NOT NULL,
		plan_type TEXT NOT NULL, -- '1month', '3months', '6months'
		activation_key TEXT UNIQUE NOT NULL,
		is_active BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		FOREIGN KEY (user_chat_id) REFERENCES users(chat_id)
	);`

	// Таблица платежей
	createPaymentsTable := `
	CREATE TABLE IF NOT EXISTS payments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_chat_id INTEGER NOT NULL,
		subscription_id INTEGER,
		amount REAL NOT NULL,
		status TEXT DEFAULT 'pending', -- 'pending', 'confirmed', 'failed'
		payment_method TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		confirmed_at DATETIME,
		FOREIGN KEY (user_chat_id) REFERENCES users(chat_id),
		FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
	);`

	// Таблица логов активности
	createLogsTable := `
	CREATE TABLE IF NOT EXISTS user_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_chat_id INTEGER NOT NULL,
		action TEXT NOT NULL, -- 'start', 'select_plan', 'payment', 'key_generated'
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_chat_id) REFERENCES users(chat_id)
	);`

	// Выполняем создание таблиц
	tables := []string{
		createUsersTable,
		createSubscriptionsTable,
		createPaymentsTable,
		createLogsTable,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			log.Fatal("Ошибка создания таблицы:", err)
		}
	}

	log.Println("Таблицы созданы успешно")
}

// Функции для работы с пользователями
func createUser(db *sql.DB, chatID int64, username, firstName, lastName string) error {
	query := `
	INSERT OR IGNORE INTO users (chat_id, username, first_name, last_name) 
	VALUES (?, ?, ?, ?)`
	
	_, err := db.Exec(query, chatID, username, firstName, lastName)
	return err
}

func getUser(db *sql.DB, chatID int64) (*domain.User, error) {
	query := `SELECT id, chat_id, username, first_name, last_name, created_at FROM users WHERE chat_id = ?`
	
	user := &User{}
	err := db.QueryRow(query, chatID).Scan(
		&user.ID, &user.ChatID, &user.Username, 
		&user.FirstName, &user.LastName, &user.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// Функции для работы с подписками
func createSubscription(db *sql.DB, chatID int64, planType, key string) error {
	query := `
	INSERT INTO subscriptions (user_chat_id, plan_type, activation_key, is_active) 
	VALUES (?, ?, ?, ?)`
	
	_, err := db.Exec(query, chatID, planType, key, true)
	return err
}

func getActiveSubscriptions(db *sql.DB, chatID int64) ([]domain.Subscription, error) {
	query := `
	SELECT id, user_chat_id, plan_type, activation_key, is_active, created_at, expires_at 
	FROM subscriptions 
	WHERE user_chat_id = ? AND is_active = true`
	
	rows, err := db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		err := rows.Scan(
			&sub.ID, &sub.UserChatID, &sub.PlanType, 
			&sub.ActivationKey, &sub.IsActive, &sub.CreatedAt, &sub.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	
	return subscriptions, nil
}

// Функция для логирования действий пользователя
func logUserAction(db *sql.DB, chatID int64, action, details string) error {
	query := `INSERT INTO user_logs (user_chat_id, action, details) VALUES (?, ?, ?)`
	_, err := db.Exec(query, chatID, action, details)
	return err
}