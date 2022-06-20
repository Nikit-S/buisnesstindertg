CREATE TABLE public.messages (
		id SERIAL PRIMARY KEY,
		chat_id varchar,
		m_id varchar,
		m_time date,
		txt varchar
);
