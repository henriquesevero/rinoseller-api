-- Tabela de produtos
CREATE TABLE IF NOT EXISTS products (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    category      TEXT NOT NULL DEFAULT '',
    price         NUMERIC(10,2) NOT NULL DEFAULT 0,
    cost_price    NUMERIC(10,2) NOT NULL DEFAULT 0,
    stock_quantity INTEGER NOT NULL DEFAULT 0
);

-- Tabela de pedidos do catálogo público
CREATE TABLE IF NOT EXISTS orders (
    id           TEXT PRIMARY KEY,
    client_name  TEXT NOT NULL,
    client_phone TEXT NOT NULL DEFAULT '',
    total        NUMERIC(10,2) NOT NULL DEFAULT 0,
    status       TEXT NOT NULL DEFAULT 'Pronta-Entrega',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id           TEXT PRIMARY KEY,
    order_id     TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id   TEXT NOT NULL,
    product_name TEXT NOT NULL,
    quantity     INTEGER NOT NULL,
    price        NUMERIC(10,2) NOT NULL
);

-- Tabela de clientes
CREATE TABLE IF NOT EXISTS clients (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    phone      TEXT NOT NULL DEFAULT '',
    email      TEXT NOT NULL DEFAULT '',
    notes      TEXT NOT NULL DEFAULT '',
    debt       NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabela de orçamentos / pedidos de venda
CREATE TABLE IF NOT EXISTS quotes (
    id          TEXT PRIMARY KEY,
    client_id   TEXT NOT NULL REFERENCES clients(id),
    client_name TEXT NOT NULL,
    total       NUMERIC(10,2) NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'Aguardando Aprovação',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    invoiced_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS quote_items (
    id           TEXT PRIMARY KEY,
    quote_id     TEXT NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    product_id   TEXT NOT NULL,
    product_name TEXT NOT NULL,
    quantity     INTEGER NOT NULL,
    unit_price   NUMERIC(10,2) NOT NULL,
    subtotal     NUMERIC(10,2) NOT NULL
);

-- Tabela de usuários do sistema
CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'seller',
    active        BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Associação de dados aos usuários (nullable para compatibilidade com dados existentes)
ALTER TABLE products ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id);
ALTER TABLE clients  ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id);
ALTER TABLE quotes   ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id);
ALTER TABLE orders   ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id);

-- Suporte a redefinição de senha
ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_expires_at TIMESTAMPTZ;

-- CPF/CNPJ e endereço do cliente
ALTER TABLE clients ADD COLUMN IF NOT EXISTS cpf_cnpj TEXT NOT NULL DEFAULT '';
ALTER TABLE clients ADD COLUMN IF NOT EXISTS address  TEXT NOT NULL DEFAULT '';

-- Código sequencial por produto
ALTER TABLE products ADD COLUMN IF NOT EXISTS code TEXT NOT NULL DEFAULT '';

-- Forma de pagamento nos orçamentos
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS payment_type TEXT NOT NULL DEFAULT '';
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS installments INTEGER NOT NULL DEFAULT 1;

-- Ciclo de recebimento por cliente
ALTER TABLE clients ADD COLUMN IF NOT EXISTS payment_cycle_days   INTEGER      NOT NULL DEFAULT 0;
ALTER TABLE clients ADD COLUMN IF NOT EXISTS payment_cycle_amount NUMERIC(10,2) NOT NULL DEFAULT 0;

-- Kits de produtos (um produto que agrega outros produtos componentes)
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_kit BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS kit_items (
    id           TEXT PRIMARY KEY,
    kit_id       TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_id   TEXT NOT NULL,
    product_name TEXT NOT NULL,
    quantity     INTEGER NOT NULL
);

-- Snapshot da composição do kit no momento da venda (para exibição em orçamentos/pedidos/nota)
ALTER TABLE quote_items ADD COLUMN IF NOT EXISTS kit_items JSONB NOT NULL DEFAULT '[]';

-- Histórico de recebimentos por cliente
CREATE TABLE IF NOT EXISTS client_payments (
    id               TEXT PRIMARY KEY,
    client_id        TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    user_id          TEXT REFERENCES users(id),
    amount           NUMERIC(10,2) NOT NULL,
    notes            TEXT NOT NULL DEFAULT '',
    count_as_revenue BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE client_payments ADD COLUMN IF NOT EXISTS count_as_revenue BOOLEAN NOT NULL DEFAULT FALSE;

-- Produtos iniciais (executado apenas se a tabela estiver vazia)
INSERT INTO products (id, name, category, price, cost_price, stock_quantity)
VALUES
    ('seed-prod-1', 'Kit Progressiva BB Cream',          'Químicos',      185.00, 110.00, 10),
    ('seed-prod-2', 'Shampoo Lavatório 5L',               'Lavatório',      89.00,  50.00, 15),
    ('seed-prod-3', 'Pó Descolorante Azul',               'Químicos',       75.00,  40.00, 20),
    ('seed-prod-4', 'Oxidante 40 Vol 900ml',              'Químicos',       28.00,  15.00, 30),
    ('seed-prod-5', 'Reparador de Pontas Óleo de Argan',  'Finalizadores',  55.00,  30.00, 12)
ON CONFLICT (id) DO NOTHING;

-- Contas a pagar (compras de distribuidores, despesas)
CREATE TABLE IF NOT EXISTS expenses (
    id             TEXT PRIMARY KEY,
    user_id        TEXT REFERENCES users(id),
    description    TEXT NOT NULL,
    supplier       TEXT NOT NULL DEFAULT '',
    amount         NUMERIC(10,2) NOT NULL,
    payment_method TEXT NOT NULL DEFAULT '',
    due_date       TIMESTAMPTZ,
    status         TEXT NOT NULL DEFAULT 'Pendente',
    paid_at        TIMESTAMPTZ,
    notes          TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Movimentações manuais de capital: aportes (entradas) e retiradas (saídas) de caixa
CREATE TABLE IF NOT EXISTS capital_contributions (
    id          TEXT PRIMARY KEY,
    user_id     TEXT REFERENCES users(id),
    description TEXT NOT NULL DEFAULT '',
    amount      NUMERIC(10,2) NOT NULL,
    type        TEXT NOT NULL DEFAULT 'aporte',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE capital_contributions ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'aporte';
ALTER TABLE capital_contributions ADD COLUMN IF NOT EXISTS hidden BOOLEAN NOT NULL DEFAULT false;

-- Limite de dívida por cliente: 0 significa sem limite
ALTER TABLE clients ADD COLUMN IF NOT EXISTS debt_limit NUMERIC(10,2) NOT NULL DEFAULT 0;

-- Timestamp de entrega, separado de invoiced_at (antes a entrega sobrescrevia
-- a data de faturamento por engano)
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ;

-- Índices para acelerar queries por user_id (filtro principal em todas as listagens)
CREATE INDEX IF NOT EXISTS idx_products_user_id            ON products(user_id);
CREATE INDEX IF NOT EXISTS idx_clients_user_id             ON clients(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id              ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_quotes_user_id              ON quotes(user_id);
CREATE INDEX IF NOT EXISTS idx_expenses_user_id            ON expenses(user_id);
CREATE INDEX IF NOT EXISTS idx_capital_contributions_user  ON capital_contributions(user_id);
CREATE INDEX IF NOT EXISTS idx_client_payments_user_id     ON client_payments(user_id);

-- Índices para buscas por chave estrangeira
CREATE INDEX IF NOT EXISTS idx_quotes_client_id            ON quotes(client_id);
CREATE INDEX IF NOT EXISTS idx_quote_items_quote_id        ON quote_items(quote_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id        ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_kit_items_kit_id            ON kit_items(kit_id);
CREATE INDEX IF NOT EXISTS idx_client_payments_client_id   ON client_payments(client_id);

-- Índices para ordenação (ORDER BY created_at DESC)
CREATE INDEX IF NOT EXISTS idx_quotes_created_at           ON quotes(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_created_at           ON orders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_created_at         ON expenses(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_capital_created_at          ON capital_contributions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_client_payments_created_at  ON client_payments(created_at DESC);

-- Índice para busca de usuário por reset_token
CREATE INDEX IF NOT EXISTS idx_users_reset_token           ON users(reset_token) WHERE reset_token IS NOT NULL;

-- Índice para casar pedidos do catálogo com o cliente pelo telefone
CREATE INDEX IF NOT EXISTS idx_orders_client_phone         ON orders(client_phone) WHERE client_phone <> '';

-- Marca do produto (ex: Wella, L'Oréal) — usado para filtro na tela de produtos
ALTER TABLE products ADD COLUMN IF NOT EXISTS brand TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS brand_catalogs (
    id          TEXT PRIMARY KEY,
    user_id     TEXT REFERENCES users(id),
    brand_name  TEXT NOT NULL,
    drive_url   TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

