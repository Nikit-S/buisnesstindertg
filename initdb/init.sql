CREATE TABLE IF NOT EXISTS public.messages (
	id serial NOT NULL,
	chat_id varchar NULL,
	m_id varchar NULL,
	is_bot bool NULL,
	from_id varchar NULL,
	m_time date NULL,
	txt varchar NULL,
	CONSTRAINT messages_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.chats (
	id varchar NOT NULL,
	first_name varchar NULL,
	last_name varchar NULL,
	username varchar NULL,
	phone varchar NULL,
	description varchar NULL,
	perm varchar NULL,
	part bool NULL,
	CONSTRAINT chats_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.pairs (
	id varchar NOT NULL,
	first_name varchar NULL,
	last_name varchar NULL,
	username varchar NULL,
	phone varchar NULL,
	description varchar NULL,
	CONSTRAINT pairs_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.user_pair_history (
	id varchar REFERENCES public.chats(id),
	pair_id varchar REFERENCES public.chats(id),
	want boolean not null,
	PRIMARY KEY (id, pair_id)
);
