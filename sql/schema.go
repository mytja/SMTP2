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
	reply_to     INTEGER
);
CREATE TABLE IF NOT EXISTS sentmsgs (
	id           INTEGER       PRIMARY KEY,
    to_email     VARCHAR(250)  NOT NULL,
	from_email   VARCHAR(250)  NOT NULL,
	title        VARCHAR(250)  NOT NULL,
	body         TEXT          NOT NULL,
	pass         VARCHAR(30)   NOT NULL,
	reply_to     INTEGER
);
CREATE TABLE IF NOT EXISTS users (
    id           INTEGER       PRIMARY KEY,
    email        VARCHAR(250)  NOT NULL,
    pass         VARCHAR(250)  NOT NULL
);
`
