--
-- import using something like:
--
-- $ createdb bank ; psql bank -f ./datamodel.sql
--
drop table if exists account cascade;
create table account (
	id			serial unique not null primary key,
	name		text,
	email		text
);

insert into account (id, name, email) values
	(default, 'dustin', 'dustin@gmail.com'),
	(default, 'deb',		'deb@hotmail.com'),
	(default, 'mark',		'mark@msn.com'),
	(default, 'tom',		'tom@mail.ru')
returning *;

drop table if exists transaction cascade;
create table transaction (
	id							serial unique not null primary key,
	account					integer references account (id) on delete cascade,
	"timestamp"			timestamp without time zone default localtimestamp,
	description			text,
	amount					money,
	fee							money
);

insert into transaction (id, account, "timestamp", description, amount, fee) values
	(default, (select id from account where email='deb@hotmail.com'), localtimestamp - '30 days'::interval, 'Deposit', 2000.00, 0.0),
	(default, (select id from account where email='deb@hotmail.com'), localtimestamp - '15 days'::interval, 'Best Buy', 1500.00, 0.0),
	(default, (select id from account where email='deb@hotmail.com'), localtimestamp - '7 days'::interval, 'Amex', -3500.25, 0.0)
returning *;

insert into transaction (id, account, "timestamp", description, amount, fee) values
	(default, (select id from account where email='dustin@gmail.com'), localtimestamp - '30 days'::interval, 'Deposit', 2000.00, 0.0),
	(default, (select id from account where email='dustin@gmail.com'), localtimestamp - '29 days'::interval, '7-11', -30.00, 0.0),
	(default, (select id from account where email='dustin@gmail.com'), localtimestamp - '27 days'::interval, 'Publix', -50.00, 0.0),
	(default, (select id from account where email='dustin@gmail.com'), localtimestamp - '20 days'::interval, 'amazon.com', -24.00, 0.0),
	(default, (select id from account where email='dustin@gmail.com'), localtimestamp - '15 days'::interval, 'ATM', -50.00, 3.00),
	(default, (select id from account where email='dustin@gmail.com'), localtimestamp - '2 days'::interval, 'Dividends', 200.00, 0.0),
	(default, (select id from account where email='dustin@gmail.com'), localtimestamp - '1 days'::interval, 'Cryptocurrency Transaction', -200.00, 0.0)
returning *;

create materialized view ledger as 
	select
		id, account, "timestamp"::date, upper(description) as description, amount, fee, sum(amount) over (partition by account order by "timestamp") as balance
	from
		transaction;
