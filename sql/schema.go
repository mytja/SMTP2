package sql

var schema string = `
CREATE TABLE IF NOT EXISTS receivedmsgs (
	id           INTEGER       PRIMARY KEY,
    uri          VARCHAR(250)  NOT NULL,
    to_email     VARCHAR(250)  NOT NULL,
	from_email   VARCHAR(250)  NOT NULL,
	title        VARCHAR(250)  NOT NULL,
	server_id    INTEGER       NOT NULL,
	server_pass  VARCHAR(30)   NOT NULL,
	warning      VARCHAR(75)
);
CREATE TABLE IF NOT EXISTS messages (
	id           INTEGER       PRIMARY KEY,
	original_id  INTEGER       NOT NULL,
	server_id    INTEGER       NOT NULL,
	reply_pass   VARCHAR(100)  NOT NULL,
	reply_id     VARCHAR(100)  NOT NULL,
	type         VARCHAR(10)   NOT NULL,
	is_draft     BOOLEAN       NOT NULL
);
CREATE TABLE IF NOT EXISTS sentmsgs (
	id           INTEGER       PRIMARY KEY,
    to_email     VARCHAR(250)  NOT NULL,
	from_email   VARCHAR(250)  NOT NULL,
	title        VARCHAR(250)  NOT NULL,
	body         TEXT          NOT NULL,
	pass         VARCHAR(30)   NOT NULL
);
CREATE TABLE IF NOT EXISTS users (
    id           INTEGER       PRIMARY KEY,
    email        VARCHAR(250)  NOT NULL,
    pass         VARCHAR(250)  NOT NULL
);
`
