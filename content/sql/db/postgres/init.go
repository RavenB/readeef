package postgres

var (
	initSQL = []string{`
CREATE TABLE IF NOT EXISTS readeef (
	db_version INTEGER
)`, `
CREATE TABLE IF NOT EXISTS users (
	login TEXT PRIMARY KEY,
	first_name TEXT,
	last_name TEXT,
	email TEXT,
	admin BOOLEAN DEFAULT 'f',
	active BOOLEAN DEFAULT 't',
	profile_data BYTEA,
	hash_type TEXT,
	salt BYTEA,
	hash BYTEA,
	md5_api BYTEA
)`, `
CREATE TABLE IF NOT EXISTS feeds (
	id SERIAL PRIMARY KEY,
	link TEXT NOT NULL UNIQUE,
	title TEXT,
	description TEXT,
	hub_link TEXT,
	site_link TEXT,
	update_error TEXT,
	subscribe_error TEXT
)`, `
CREATE TABLE IF NOT EXISTS feed_images (
	id SERIAL PRIMARY KEY,
	feed_id INTEGER NOT NULL,
	title TEXT,
	url TEXT,
	width INTEGER,
	height INTEGER,

	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS articles (
	id BIGSERIAL PRIMARY KEY,
	feed_id INTEGER,
	link TEXT,
	guid TEXT,
	title TEXT,
	description TEXT,
	date TIMESTAMP WITH TIME ZONE,

	UNIQUE(feed_id, link),
	UNIQUE(feed_id, guid),
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_feeds (
	user_login TEXT,
	feed_id INTEGER,

	PRIMARY KEY(user_login, feed_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_feeds_tags (
	user_login TEXT,
	feed_id INTEGER,
	tag TEXT,

	PRIMARY KEY(user_login, feed_id, tag),
	FOREIGN KEY(user_login, feed_id) REFERENCES users_feeds(user_login, feed_id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_read (
	user_login TEXT,
	article_id BIGINT,

	PRIMARY KEY(user_login, article_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_fav (
	user_login TEXT,
	article_id BIGINT,

	PRIMARY KEY(user_login, article_id),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS articles_scores (
	article_id BIGINT,
	score  BIGINT,
	score1 BIGINT,
	score2 BIGINT,
	score3 BIGINT,
	score4 BIGINT,
	score5 BIGINT,

	PRIMARY KEY(article_id),
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS hubbub_subscriptions (
	feed_id INTEGER,
	link TEXT,
	lease_duration INTEGER,
	verification_time TIMESTAMP WITH TIME ZONE,
	subscription_failure BOOLEAN DEFAULT 'f',

	PRIMARY KEY(feed_id),
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS domain_https_support (
	domain TEXT,
	https BOOLEAN DEFAULT 'f',

	PRIMARY KEY(domain)
)`,
	}
)
