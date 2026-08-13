create table if not exists poker_users
(
    id          bigint generated always as identity
        primary key,
    username    text                                   not null,
    created_at  timestamp with time zone default now() not null,
    disabled_at timestamp with time zone,
    twitch_user text                                   not null
);

create table if not exists just_poker_game
(
    game_id        bigserial
        primary key,
    starting_chips jsonb                                  not null,
    player_ids     bigint[],
    created_at     timestamp with time zone default now() not null,
    ended_at       timestamp with time zone,
    player_count   integer                                not null
);

create table if not exists poker_api_keys
(
    id           bigserial
        primary key,
    key_id       text                                   not null,
    user_id      bigint                                 not null
        references poker_users
            on delete cascade,
    username     text                                   not null,
    mac          bytea                                  not null,
    created_at   timestamp with time zone default now() not null,
    revoked_at   timestamp with time zone,
    last_used_at timestamp with time zone
);