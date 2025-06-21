package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./bot.db")
	if err != nil {
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

	// Обновленная таблица платежей для CloudPayments
	createPaymentsTable := `
	CREATE TABLE IF NOT EXISTS payments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_chat_id INTEGER NOT NULL,
		subscription_id INTEGER,
		amount REAL NOT NULL,
		currency TEXT DEFAULT 'RUB',
		status TEXT DEFAULT 'pending', -- 'pending', 'completed', 'failed', 'cancelled'
		payment_method TEXT DEFAULT 'cloudpayments',
		provider_transaction_id TEXT,
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

	// Создаем индексы для быстрого поиска
	createIndexes(db)

	log.Println("Таблицы созданы успешно")
}

func createIndexes(db *sql.DB) {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_payments_user_chat_id ON payments(user_chat_id);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_provider_transaction_id ON payments(provider_transaction_id);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_user_chat_id ON subscriptions(user_chat_id);`,
		`CREATE INDEX IF NOT EXISTS idx_user_logs_user_chat_id ON user_logs(user_chat_id);`,
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			log.Printf("Предупреждение при создании индекса: %v", err)
		}
	}
}