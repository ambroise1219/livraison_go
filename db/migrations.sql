-- =============================================
-- MIGRATIONS SQLITE - ILEX LIVRAISON
-- Migration complète depuis SurrealDB vers SQLite
-- =============================================

-- Activer les clés étrangères
PRAGMA foreign_keys = ON;

-- =============================================
-- TABLE USER
-- =============================================
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    phone TEXT NOT NULL UNIQUE,
    address TEXT,
    role TEXT NOT NULL DEFAULT 'CLIENT' CHECK (role IN ('CLIENT', 'LIVREUR', 'ADMIN', 'GESTIONNAIRE', 'MARKETING')),
    referred_by_id TEXT REFERENCES users(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    profile_picture_id TEXT,
    last_name TEXT NOT NULL DEFAULT '',
    first_name TEXT NOT NULL DEFAULT '',
    email TEXT,
    date_of_birth DATETIME,
    lieu_residence TEXT,
    is_profile_completed BOOLEAN NOT NULL DEFAULT 0,
    is_driver_complete BOOLEAN NOT NULL DEFAULT 0,
    is_driver_vehicule_complete BOOLEAN NOT NULL DEFAULT 0,
    cni_recto TEXT,
    cni_verso TEXT,
    permis_recto TEXT,
    permis_verso TEXT,
    driver_status TEXT NOT NULL DEFAULT 'OFFLINE' CHECK (driver_status IN ('OFFLINE', 'ONLINE', 'BUSY', 'AVAILABLE')),
    last_known_lat REAL,
    last_known_lng REAL,
    last_seen_at DATETIME
);

-- Indexes pour users
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_referred_by_id ON users(referred_by_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_driver_status ON users(role, driver_status);
CREATE INDEX IF NOT EXISTS idx_users_location ON users(last_known_lat, last_known_lng);

-- =============================================
-- TABLE OTP
-- =============================================
CREATE TABLE IF NOT EXISTS otps (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    phone TEXT NOT NULL,
    code TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_otps_phone_code_expires ON otps(phone, code, expires_at);

-- =============================================
-- TABLE VEHICLE
-- =============================================
CREATE TABLE IF NOT EXISTS vehicles (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    type TEXT NOT NULL CHECK (type IN ('MOTO', 'VOITURE', 'CAMIONNETTE')),
    user_id TEXT NOT NULL REFERENCES users(id),
    nom TEXT,
    plaque_immatriculation TEXT UNIQUE,
    couleur TEXT,
    marque TEXT,
    modele TEXT,
    annee INTEGER,
    carte_grise_recto TEXT,
    carte_grise_verso TEXT,
    vignette_recto TEXT,
    vignette_verso TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vehicles_user_id ON vehicles(user_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_plaque ON vehicles(plaque_immatriculation);

-- =============================================
-- TABLE LOCATION
-- =============================================
CREATE TABLE IF NOT EXISTS locations (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    address TEXT NOT NULL,
    lat REAL,
    lng REAL
);

CREATE INDEX IF NOT EXISTS idx_locations_coordinates ON locations(lat, lng);

-- =============================================
-- TABLE DELIVERY
-- =============================================
CREATE TABLE IF NOT EXISTS deliveries (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    client_id TEXT NOT NULL REFERENCES users(id),
    livreur_id TEXT REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'PICKED_UP', 'DELIVERED', 'CANCELLED', 'ZONE_ASSIGNED', 'PICKUP_IN_PROGRESS', 'PICKUP_COMPLETED', 'DELIVERY_IN_PROGRESS', 'ASSIGNED_TO_HELPER', 'HELPERS_CONFIRMED', 'ARRIVED_AT_PICKUP', 'LOADING_IN_PROGRESS', 'LOADING_COMPLETED', 'IN_TRANSIT', 'ARRIVED_AT_DESTINATION', 'UNLOADING_IN_PROGRESS', 'UNLOADING_COMPLETED', 'ARRIVED_AT_DROPOFF', 'EN_ROUTE', 'DISPATCH_IN_PROGRESS', 'SORTED', 'SORTING_IN_PROGRESS')),
    type TEXT NOT NULL CHECK (type IN ('SIMPLE', 'EXPRESS', 'GROUPEE', 'DEMENAGEMENT')),
    pickup_id TEXT NOT NULL REFERENCES locations(id),
    dropoff_id TEXT NOT NULL REFERENCES locations(id),
    distance_km REAL,
    duration_min REAL,
    vehicle_type TEXT NOT NULL CHECK (vehicle_type IN ('MOTO', 'VOITURE', 'CAMIONNETTE')),
    base_price REAL,
    waiting_min REAL,
    final_price REAL NOT NULL DEFAULT 0 CHECK (final_price >= 0),
    payment_method TEXT NOT NULL CHECK (payment_method IN ('CASH', 'MOBILE_MONEY_ORANGE', 'MOBILE_MONEY_MTN', 'MOBILE_MONEY_MOOV', 'MOBILE_MONEY_WAVE')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    paid_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_deliveries_client_type_status ON deliveries(client_id, type, status);
CREATE INDEX IF NOT EXISTS idx_deliveries_livreur_type_status ON deliveries(livreur_id, type, status);
CREATE INDEX IF NOT EXISTS idx_deliveries_pickup_id ON deliveries(pickup_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_dropoff_id ON deliveries(dropoff_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_created_at ON deliveries(created_at);

-- =============================================
-- TABLE PACKAGE
-- =============================================
CREATE TABLE IF NOT EXISTS packages (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    delivery_id TEXT NOT NULL REFERENCES deliveries(id),
    description TEXT,
    weight_kg REAL,
    size TEXT,
    fragile BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_packages_delivery_id ON packages(delivery_id);

-- =============================================
-- TABLE DRIVER_LOCATION (WITHOUT ROWID pour performance)
-- =============================================
CREATE TABLE IF NOT EXISTS driver_locations (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    driver_id TEXT NOT NULL REFERENCES users(id),
    lat REAL,
    lng REAL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_available BOOLEAN NOT NULL DEFAULT 1,
    vehicle_type TEXT NOT NULL CHECK (vehicle_type IN ('MOTO', 'VOITURE', 'CAMIONNETTE'))
) WITHOUT ROWID;

-- Indices optimisés pour géolocalisation et disponibilité
CREATE INDEX IF NOT EXISTS idx_driver_locations_driver_available ON driver_locations(driver_id, is_available);
CREATE INDEX IF NOT EXISTS idx_driver_locations_available_location ON driver_locations(is_available, lat, lng) WHERE is_available = 1;
CREATE INDEX IF NOT EXISTS idx_driver_locations_vehicle_available ON driver_locations(vehicle_type, is_available, lat, lng) WHERE is_available = 1;
CREATE INDEX IF NOT EXISTS idx_driver_locations_timestamp ON driver_locations(timestamp DESC);

-- =============================================
-- TABLE TRACKING (WITHOUT ROWID pour performance)
-- =============================================
CREATE TABLE IF NOT EXISTS trackings (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    delivery_id TEXT NOT NULL REFERENCES deliveries(id),
    lat REAL,
    lng REAL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) WITHOUT ROWID;

-- Indices optimisés pour requêtes fréquentes
CREATE INDEX IF NOT EXISTS idx_trackings_delivery_timestamp ON trackings(delivery_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_trackings_timestamp ON trackings(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_trackings_location ON trackings(lat, lng) WHERE lat IS NOT NULL AND lng IS NOT NULL;

-- =============================================
-- TABLE REFERRAL
-- =============================================
CREATE TABLE IF NOT EXISTS referrals (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    referrer_id TEXT NOT NULL REFERENCES users(id),
    referee_phone TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    code TEXT NOT NULL,
    completed_at DATETIME,
    expires_at DATETIME,
    message TEXT,
    referee_id TEXT REFERENCES users(id),
    reward_claimed_at DATETIME,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'EXPIRED', 'CANCELLED', 'REWARD_CLAIMED'))
);

CREATE INDEX IF NOT EXISTS idx_referrals_referrer_id ON referrals(referrer_id);
CREATE INDEX IF NOT EXISTS idx_referrals_referee_phone ON referrals(referee_phone);
CREATE INDEX IF NOT EXISTS idx_referrals_code ON referrals(code);
CREATE INDEX IF NOT EXISTS idx_referrals_status ON referrals(status);

-- =============================================
-- TABLE PROMO
-- =============================================
CREATE TABLE IF NOT EXISTS promos (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    code TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_date DATETIME NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    max_usage INTEGER,
    min_purchase_amount REAL,
    start_date DATETIME NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('PERCENTAGE', 'FIXED_AMOUNT', 'FREE_DELIVERY')),
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    usage_count INTEGER,
    value REAL NOT NULL DEFAULT 0 CHECK (value >= 0)
);

CREATE INDEX IF NOT EXISTS idx_promos_code ON promos(code);
CREATE INDEX IF NOT EXISTS idx_promos_is_active ON promos(is_active);
CREATE INDEX IF NOT EXISTS idx_promos_start_date ON promos(start_date);
CREATE INDEX IF NOT EXISTS idx_promos_end_date ON promos(end_date);

-- =============================================
-- TABLE PROMO_USAGE
-- =============================================
CREATE TABLE IF NOT EXISTS promo_usages (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    promo_id TEXT NOT NULL REFERENCES promos(id),
    user_id TEXT NOT NULL REFERENCES users(id),
    used_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    amount REAL NOT NULL DEFAULT 0 CHECK (amount >= 0),
    discount REAL NOT NULL DEFAULT 0 CHECK (discount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_promo_usages_promo_id ON promo_usages(promo_id);
CREATE INDEX IF NOT EXISTS idx_promo_usages_user_id ON promo_usages(user_id);
CREATE INDEX IF NOT EXISTS idx_promo_usages_used_at ON promo_usages(used_at);

-- =============================================
-- TABLE PRICING_RULE
-- =============================================
CREATE TABLE IF NOT EXISTS pricing_rules (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    vehicle_type TEXT NOT NULL UNIQUE CHECK (vehicle_type IN ('MOTO', 'VOITURE', 'CAMIONNETTE')),
    base_price REAL NOT NULL DEFAULT 0 CHECK (base_price >= 0),
    included_km REAL NOT NULL DEFAULT 0 CHECK (included_km >= 0),
    per_km REAL NOT NULL DEFAULT 0 CHECK (per_km >= 0),
    waiting_free INTEGER NOT NULL DEFAULT 0 CHECK (waiting_free >= 0),
    waiting_rate REAL NOT NULL DEFAULT 0 CHECK (waiting_rate >= 0)
);

-- =============================================
-- TABLE NOTIFICATION
-- =============================================
CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL REFERENCES users(id),
    type TEXT NOT NULL CHECK (type IN ('SMS', 'EMAIL', 'PUSH')),
    title TEXT,
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENT', 'FAILED')),
    metadata TEXT, -- JSON
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_status ON notifications(user_id, status);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

-- =============================================
-- TABLE FILE
-- =============================================
CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    public_id TEXT NOT NULL,
    url TEXT NOT NULL,
    secure_url TEXT NOT NULL,
    format TEXT,
    width INTEGER,
    height INTEGER,
    resource_type TEXT NOT NULL,
    size INTEGER NOT NULL DEFAULT 0 CHECK (size >= 0),
    original_filename TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_files_public_id ON files(public_id);
CREATE INDEX IF NOT EXISTS idx_files_resource_type ON files(resource_type);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);

-- =============================================
-- TABLE SUBSCRIPTION
-- =============================================
CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL REFERENCES users(id),
    type TEXT NOT NULL CHECK (type IN ('BASIC', 'STANDARD', 'PREMIUM')),
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'EXPIRED', 'CANCELLED')),
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    price REAL NOT NULL DEFAULT 0 CHECK (price >= 0),
    delivery_credits INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payment_id TEXT REFERENCES payments(id)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_status ON subscriptions(user_id, status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_payment_id ON subscriptions(payment_id);

-- =============================================
-- TABLE PAYMENT
-- =============================================
CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL REFERENCES users(id),
    amount REAL NOT NULL DEFAULT 0 CHECK (amount >= 0),
    method TEXT NOT NULL CHECK (method IN ('CASH', 'WAVE', 'MOBILE_MONEY_ORANGE', 'MOBILE_MONEY_MTN', 'MOBILE_MONEY_MOOV')),
    payment_type TEXT NOT NULL CHECK (payment_type IN ('DELIVERY_PAYMENT', 'WALLET_RECHARGE', 'SUBSCRIPTION', 'FINE', 'BONUS')),
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'REFUNDED')),
    transaction_id TEXT,
    reference TEXT NOT NULL UNIQUE,
    metadata TEXT, -- JSON
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_payments_user_status ON payments(user_id, status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
CREATE INDEX IF NOT EXISTS idx_payments_reference ON payments(reference);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments(transaction_id);
CREATE INDEX IF NOT EXISTS idx_payments_payment_type ON payments(payment_type);
CREATE INDEX IF NOT EXISTS idx_payments_method ON payments(method);

-- =============================================
-- TABLE WALLET
-- =============================================
CREATE TABLE IF NOT EXISTS wallets (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL UNIQUE REFERENCES users(id),
    balance REAL NOT NULL DEFAULT 0.0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =============================================
-- TABLE WALLET_TRANSACTION
-- =============================================
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    wallet_id TEXT NOT NULL REFERENCES wallets(id),
    amount REAL NOT NULL,
    type TEXT NOT NULL,
    description TEXT,
    reference TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_type ON wallet_transactions(type);

-- =============================================
-- TABLE REFRESH_TOKEN
-- =============================================
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL REFERENCES users(id),
    token_value TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_value ON refresh_tokens(token_value);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- =============================================
-- TABLE GROUPED_DELIVERY
-- =============================================
CREATE TABLE IF NOT EXISTS grouped_deliveries (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    delivery_id TEXT NOT NULL UNIQUE REFERENCES deliveries(id),
    total_zones INTEGER NOT NULL CHECK (total_zones > 0),
    completed_zones INTEGER NOT NULL DEFAULT 0,
    discount_percentage REAL NOT NULL DEFAULT 30.0,
    original_price REAL NOT NULL DEFAULT 0 CHECK (original_price >= 0),
    final_price REAL NOT NULL DEFAULT 0 CHECK (final_price >= 0),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =============================================
-- TABLE DELIVERY_ZONE
-- =============================================
CREATE TABLE IF NOT EXISTS delivery_zones (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    grouped_delivery_id TEXT NOT NULL REFERENCES grouped_deliveries(id),
    zone_number INTEGER NOT NULL CHECK (zone_number > 0),
    recipient_name TEXT NOT NULL,
    recipient_phone TEXT NOT NULL,
    pickup_location_id TEXT NOT NULL REFERENCES locations(id),
    delivery_location_id TEXT NOT NULL REFERENCES locations(id),
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'PICKED_UP', 'DELIVERED', 'CANCELLED', 'ZONE_ASSIGNED', 'PICKUP_IN_PROGRESS', 'PICKUP_COMPLETED', 'DELIVERY_IN_PROGRESS', 'ASSIGNED_TO_HELPER', 'HELPERS_CONFIRMED', 'ARRIVED_AT_PICKUP', 'LOADING_IN_PROGRESS', 'LOADING_COMPLETED', 'IN_TRANSIT', 'ARRIVED_AT_DESTINATION', 'UNLOADING_IN_PROGRESS', 'UNLOADING_COMPLETED', 'ARRIVED_AT_DROPOFF', 'EN_ROUTE', 'DISPATCH_IN_PROGRESS', 'SORTED', 'SORTING_IN_PROGRESS')),
    price REAL NOT NULL DEFAULT 0 CHECK (price >= 0),
    picked_up_at DATETIME,
    delivered_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_delivery_zones_grouped_zone ON delivery_zones(grouped_delivery_id, zone_number);
CREATE INDEX IF NOT EXISTS idx_delivery_zones_grouped_delivery_id ON delivery_zones(grouped_delivery_id);
CREATE INDEX IF NOT EXISTS idx_delivery_zones_status ON delivery_zones(status);

-- =============================================
-- TABLE USER_ADDRESS
-- =============================================
CREATE TABLE IF NOT EXISTS user_addresses (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL REFERENCES users(id),
    name TEXT,
    description TEXT,
    address TEXT NOT NULL,
    lat REAL,
    lng REAL,
    type TEXT NOT NULL DEFAULT 'OTHER' CHECK (type IN ('HOME', 'WORK', 'BUSINESS', 'OTHER')),
    is_default BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_addresses_user_id ON user_addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_user_addresses_is_default ON user_addresses(is_default);

-- =============================================
-- TABLE INCIDENT
-- =============================================
CREATE TABLE IF NOT EXISTS incidents (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    delivery_id TEXT NOT NULL REFERENCES deliveries(id),
    driver_id TEXT NOT NULL REFERENCES users(id),
    client_id TEXT NOT NULL REFERENCES users(id),
    type TEXT NOT NULL CHECK (type IN ('ACCIDENT', 'BREAKDOWN', 'TRAFFIC', 'WEATHER', 'CUSTOMER_NOT_AVAILABLE', 'WRONG_ADDRESS', 'PAYMENT_ISSUE', 'PACKAGE_DAMAGED', 'PACKAGE_LOST', 'VEHICLE_ISSUE', 'ROAD_BLOCKED', 'CUSTOMER_REFUSED')),
    severity TEXT NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    description TEXT,
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED', 'ESCALATED')),
    reported_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    resolution TEXT,
    reported_by TEXT NOT NULL REFERENCES users(id),
    assigned_to TEXT REFERENCES users(id),
    priority TEXT NOT NULL DEFAULT 'MEDIUM' CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT'))
);

CREATE INDEX IF NOT EXISTS idx_incidents_delivery_id ON incidents(delivery_id);
CREATE INDEX IF NOT EXISTS idx_incidents_driver_id ON incidents(driver_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
CREATE INDEX IF NOT EXISTS idx_incidents_type ON incidents(type);
CREATE INDEX IF NOT EXISTS idx_incidents_reported_at ON incidents(reported_at);
CREATE INDEX IF NOT EXISTS idx_incidents_assigned_to ON incidents(assigned_to);

-- =============================================
-- TABLE RATING
-- =============================================
CREATE TABLE IF NOT EXISTS ratings (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    delivery_id TEXT NOT NULL REFERENCES deliveries(id),
    client_id TEXT NOT NULL REFERENCES users(id),
    driver_id TEXT NOT NULL REFERENCES users(id),
    client_rating INTEGER NOT NULL CHECK (client_rating >= 1 AND client_rating <= 5),
    driver_rating INTEGER NOT NULL CHECK (driver_rating >= 1 AND driver_rating <= 5),
    client_comment TEXT,
    driver_comment TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_anonymous BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ratings_delivery_id ON ratings(delivery_id);
CREATE INDEX IF NOT EXISTS idx_ratings_client_id ON ratings(client_id);
CREATE INDEX IF NOT EXISTS idx_ratings_driver_id ON ratings(driver_id);
CREATE INDEX IF NOT EXISTS idx_ratings_client_rating ON ratings(client_rating);
CREATE INDEX IF NOT EXISTS idx_ratings_driver_rating ON ratings(driver_rating);
CREATE INDEX IF NOT EXISTS idx_ratings_created_at ON ratings(created_at);

-- =============================================
-- TABLE MOVING_SERVICE
-- =============================================
CREATE TABLE IF NOT EXISTS moving_services (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    delivery_id TEXT NOT NULL REFERENCES deliveries(id),
    vehicle_size TEXT NOT NULL CHECK (vehicle_size IN ('MINI', 'SMALL', 'MEDIUM', 'LARGE', 'EXTRA_LARGE')),
    helpers_count INTEGER NOT NULL DEFAULT 1,
    floors INTEGER NOT NULL DEFAULT 1,
    has_elevator BOOLEAN NOT NULL DEFAULT 0,
    needs_disassembly BOOLEAN NOT NULL DEFAULT 0,
    has_fragile_items BOOLEAN NOT NULL DEFAULT 0,
    additional_services TEXT, -- JSON array
    special_instructions TEXT,
    estimated_volume REAL,
    helpers_cost REAL NOT NULL DEFAULT 0.0,
    vehicle_cost REAL NOT NULL DEFAULT 0.0,
    service_cost REAL NOT NULL DEFAULT 0.0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_moving_services_delivery_id ON moving_services(delivery_id);
CREATE INDEX IF NOT EXISTS idx_moving_services_vehicle_size ON moving_services(vehicle_size);

-- =============================================
-- TRIGGERS POUR MISE À JOUR AUTOMATIQUE
-- =============================================

-- Trigger pour mettre à jour updated_at automatiquement
CREATE TRIGGER IF NOT EXISTS update_users_updated_at 
    AFTER UPDATE ON users 
    FOR EACH ROW 
    BEGIN 
        UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_deliveries_updated_at 
    AFTER UPDATE ON deliveries 
    FOR EACH ROW 
    BEGIN 
        UPDATE deliveries SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_grouped_deliveries_updated_at 
    AFTER UPDATE ON grouped_deliveries 
    FOR EACH ROW 
    BEGIN 
        UPDATE grouped_deliveries SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_delivery_zones_updated_at 
    AFTER UPDATE ON delivery_zones 
    FOR EACH ROW 
    BEGIN 
        UPDATE delivery_zones SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_user_addresses_updated_at 
    AFTER UPDATE ON user_addresses 
    FOR EACH ROW 
    BEGIN 
        UPDATE user_addresses SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_wallets_updated_at 
    AFTER UPDATE ON wallets 
    FOR EACH ROW 
    BEGIN 
        UPDATE wallets SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_files_updated_at 
    AFTER UPDATE ON files 
    FOR EACH ROW 
    BEGIN 
        UPDATE files SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_subscriptions_updated_at 
    AFTER UPDATE ON subscriptions 
    FOR EACH ROW 
    BEGIN 
        UPDATE subscriptions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_promos_updated_at 
    AFTER UPDATE ON promos 
    FOR EACH ROW 
    BEGIN 
        UPDATE promos SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- =============================================
-- DONNÉES INITIALES
-- =============================================

-- Insérer les règles de tarification par défaut
INSERT OR IGNORE INTO pricing_rules (vehicle_type, base_price, included_km, per_km, waiting_free, waiting_rate) VALUES
('MOTO', 500.0, 2.0, 100.0, 10, 50.0),
('VOITURE', 800.0, 3.0, 150.0, 15, 75.0),
('CAMIONNETTE', 1200.0, 5.0, 200.0, 20, 100.0);

-- Insérer un utilisateur admin par défaut
INSERT OR IGNORE INTO users (id, phone, role, first_name, last_name, is_profile_completed) VALUES
('admin-001', '+225000000000', 'ADMIN', 'Admin', 'System', 1);
