CREATE TABLE cities
(
    id      uuid primary key default gen_random_uuid(),
    name    VARCHAR(100) NOT NULL,
    lat     FLOAT        NOT NULL,
    lon     FLOAT        NOT NULL,
    country VARCHAR(100) NOT NULL,

    CONSTRAINT name_country_unique UNIQUE (name, country)
);

CREATE TABLE weather
(
    city_id   uuid      NOT NULL,
    temp      FLOAT     NOT NULL,
    date      TIMESTAMP NOT NULL,
    data_json json      NOT NULL,

    CONSTRAINT city_fk FOREIGN KEY (city_id) REFERENCES cities (id),
    CONSTRAINT city_date_unique UNIQUE (city_id, date)
);

CREATE TABLE users
(
    uuid     uuid primary key default gen_random_uuid(),
    email    VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100)        NOT NULL
);

CREATE table user_favorites
(
    user_id uuid NOT NULL,
    city_id uuid NOT NULL,

    CONSTRAINT user_fk FOREIGN KEY (user_id) REFERENCES users (uuid),
    CONSTRAINT city_fk FOREIGN KEY (city_id) REFERENCES cities (id),
    CONSTRAINT user_fav_unique UNIQUE (user_id, city_id)
);
