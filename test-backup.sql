-- PostgreSQL database dump
-- Ультимативный стресс-тест для restore-assert (Уровень: Хардкор)

SET statement_timeout = 0;
SET client_encoding = 'UTF8';

-- =======================================================================
-- 1. ТЯЖЕЛЫЕ РАСШИРЕНИЯ (Extensions)
-- =======================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- Для генерации UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- Для хеширования паролей
CREATE EXTENSION IF NOT EXISTS "citext";    -- Регистронезависимый текст

-- =======================================================================
-- 2. СЕТКА РОЛЕЙ И СЛОЖНЫЕ ПРАВА (Roles & RBAC)
-- =======================================================================
DO $$ 
BEGIN 
    -- Роль 1: Только чтение
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app_readonly') THEN 
        CREATE ROLE app_readonly; 
    END IF;
    -- Роль 2: Только запись (Data Writer)
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app_writer') THEN 
        CREATE ROLE app_writer; 
    END IF;
    -- Роль 3: Менеджер схем (Админ конкретной структуры)
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'schema_manager') THEN 
       CREATE ROLE schema_manager; 
   END IF;
END $$;

-- =======================================================================
-- 3. МНОГОУРОВНЕВЫЕ СХЕМЫ (Schemas)
-- =======================================================================
CREATE SCHEMA IF NOT EXISTS core_data;
CREATE SCHEMA IF NOT EXISTS analytics_priv;
CREATE SCHEMA IF NOT EXISTS billing_secure;

-- Передаем владение схемами менеджеру (чтобы проверить права)
ALTER SCHEMA core_data OWNER TO schema_manager;
ALTER SCHEMA analytics_priv OWNER TO schema_manager;
ALTER SCHEMA billing_secure OWNER TO schema_manager;

-- =======================================================================
-- 4. СЛОЖНЫЕ ТАБЛИЦЫ С КРОСС-СХЕМНЫМИ ЗАВИСИМОСТЯМИ
-- =======================================================================

-- Таблица А (Схема core_data) - Пользователи
CREATE TABLE IF NOT EXISTS core_data.users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username citext UNIQUE NOT NULL,
    pass_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица Б (Схема billing_secure) - Счета (Зависит от core_data.users)
CREATE TABLE IF NOT EXISTS billing_secure.wallets (
    wallet_id SERIAL PRIMARY KEY,
    owner_id UUID NOT NULL,
    balance NUMERIC(15,2) DEFAULT 0.00 CHECK (balance >= 0),
    CONSTRAINT fk_wallet_owner FOREIGN KEY (owner_id) REFERENCES core_data.users(user_id) ON DELETE CASCADE
);

-- Таблица В (Схема analytics_priv) - Логи транзакций (Зависит от billing_secure.wallets)
CREATE TABLE IF NOT EXISTS analytics_priv.transaction_logs (
    log_id SERIAL PRIMARY KEY,
    wallet_id INT NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    tx_type VARCHAR(10) CHECK (tx_type IN ('deposit', 'withdraw')),
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_tx_wallet FOREIGN KEY (wallet_id) REFERENCES billing_secure.wallets(wallet_id)
);

-- =======================================================================
-- 5. МАССИВНЫЕ ВСТАВКИ И ПЕРЕКРЕСТНЫЕ ДАННЫЕ (Data Inserts)
-- =======================================================================

-- Шаг 1: Создаем пользователей и фиксируем их UUID через переменные сессии
-- (Так как UUID генерируются случайно, мы используем хитрый трюк Postgres)

WITH inserted_users AS (
    INSERT INTO core_data.users (username, pass_hash) VALUES 
    ('alex_dev', crypt('my_secret_pass', gen_salt('bf'))),
    ('murad_admin', crypt('super_secure_99', gen_salt('bf'))),
    ('guest_user', crypt('123456', gen_salt('bf')))
    RETURNING user_id, username
)
-- Шаг 2: Тут же создаем им кошельки в другой схеме
INSERT INTO billing_secure.wallets (owner_id, balance)
SELECT user_id, 1500.00 FROM inserted_users WHERE username = 'alex_dev'
UNION ALL
SELECT user_id, 99999.50 FROM inserted_users WHERE username = 'murad_admin'
UNION ALL
SELECT user_id, 0.00 FROM inserted_users WHERE username = 'guest_user';

-- Шаг 3: Генерируем пачку транзакций в третьей схеме!
INSERT INTO analytics_priv.transaction_logs (wallet_id, amount, tx_type)
SELECT wallet_id, 500.00, 'deposit' FROM billing_secure.wallets WHERE balance > 10000
UNION ALL
SELECT wallet_id, 100.00, 'withdraw' FROM billing_secure.wallets WHERE balance < 2000 AND balance > 0;

-- =======================================================================
-- 6. ТОНКИНГ ПРАВ (Grants & Permissions)
-- =======================================================================

-- Права для Readonly роли (Видит всё во всех схемах, но ничего не может менять)
GRANT USAGE ON SCHEMA core_data TO app_readonly;
GRANT USAGE ON SCHEMA billing_secure TO app_readonly;
GRANT USAGE ON SCHEMA analytics_priv TO app_readonly;

GRANT SELECT ON ALL TABLES IN SCHEMA core_data TO app_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA billing_secure TO app_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA analytics_priv TO app_readonly;

-- Права для Writer роли (Может добавлять транзакции, но не может трогать юзеров!)
GRANT USAGE ON SCHEMA billing_secure TO app_writer;
GRANT USAGE ON SCHEMA analytics_priv TO app_writer;

GRANT SELECT, UPDATE ON billing_secure.wallets TO app_writer;
GRANT SELECT, INSERT ON analytics_priv.transaction_logs TO app_writer;

-- =======================================================================
-- 7. ПРЕДСТАВЛЕНИЯ (Views) И ОГРАНИЧЕНИЯ НА НИХ
-- =======================================================================
-- Создаем вьюху, которая соединяет данные из ТРЕХ схем
CREATE VIEW core_data.v_rich_users_report AS 
SELECT 
    u.username,
    w.balance,
    t.amount AS last_tx_amount
FROM core_data.users u
JOIN billing_secure.wallets w ON u.user_id = w.owner_id
LEFT JOIN analytics_priv.transaction_logs t ON w.wallet_id = t.wallet_id
WHERE w.balance > 1000;

-- Даем доступ к этой вьюхе роли только для чтения
GRANT SELECT ON core_data.v_rich_users_report TO app_readonly;
